// cmd/focus-helper/main.go
// REFACTORED: Sunday, August 10, 2025
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"focus-helper/pkg/actions"
	"focus-helper/pkg/activity"
	"focus-helper/pkg/config"
	"focus-helper/pkg/database"
	"focus-helper/pkg/language"
	"focus-helper/pkg/llm"
	"focus-helper/pkg/models"
	"focus-helper/pkg/notifications"
	"focus-helper/pkg/persona"
	"focus-helper/pkg/state"
	"focus-helper/pkg/variables"

	_ "github.com/mattn/go-sqlite3"
)

var (
	appConfig       models.Config
	db              *sql.DB
	activityMonitor *activity.Monitor
	actionExecutor  *actions.Executor
	notifier        notifications.Notifier
)

func main() {
	clearTempAudioOnExit()

	debugFlag := flag.Bool("debug", false, "Set to true to enable debug mode")
	profileFlag := flag.String("profile", "default", "Profile name to load from profiles.json")
	flag.Parse()
	profilePath := filepath.Join(config.GetUserConfigPath(), "profiles.json")
	profiles, err := config.LoadProfiles(profilePath)
	if err != nil {
		log.Fatalf("Error load profile: %v", err)
	}
	cfg, err := config.GetProfileByName(profiles, *profileFlag)
	if err != nil {
		log.Fatalf("Profile '%s' not found: %v", *profileFlag, err)
	}
	if *debugFlag {
		cfg.DEBUG = true
		cfg.MinRandomQuestion = models.Duration{Duration: (cfg.MinRandomQuestion.Duration / 2) * time.Second}
		cfg.MaxRandomQuestion = models.Duration{Duration: (cfg.MaxRandomQuestion.Duration / 2) * time.Second}
		for i := range cfg.AlertLevels {
			cfg.AlertLevels[i].Threshold = models.Duration{Duration: (cfg.AlertLevels[i].Threshold.Duration / 2) * time.Second}
		}
		log.Printf("WARNING: all times is set to half and converted to seconds.")
	}

	appConfig = *cfg
	setupLogger()

	log.Println("--- Starting focus helper ---")
	log.Printf("PERSON ACTIVE: %s", appConfig.PersonaName)

	if appConfig.DEBUG {
		log.Println("!!!!!!!!!! RUNNING IN DEBUG MODE !!!!!!!!!!")
	}

	db, err = database.Init(appConfig.DatabaseFile)
	if err != nil {
		log.Fatalf("Fail to start database: %v", err)
	}
	defer db.Close()

	activityMonitor = activity.NewMonitor()
	notifier = notifications.NewDesktopNotifier()

	llmAdapter, err := llm.NewAdapter(appConfig.IAModel)
	if err != nil {
		log.Fatalf("Fail to start LLM adapter: %v", err)
	}
	variablesProcessor := variables.NewProcessor()

	executorDeps := actions.ExecutorDependencies{
		AppConfig:    &appConfig,
		VarProcessor: variablesProcessor,
		Notifier:     notifier,
		LLMAdapter:   llmAdapter,
	}
	actionExecutor = actions.NewExecutor(executorDeps)
	tempAudioDir := filepath.Join(config.GetUserConfigPath(), config.TEMP_AUDIO_DIR)
	if err := os.MkdirAll(tempAudioDir, 0755); err != nil {
		log.Fatalf("Failed to create temp audio directory: %v", err)
	}
	fs := http.FileServer(http.Dir(tempAudioDir))
	http.Handle("/temp_audio/", http.StripPrefix("/temp_audio/", fs))
	go func() {
		log.Printf("Running audio server at http://localhost:%s", config.SERVER_PORT)
		if err := http.ListenAndServe(":"+config.SERVER_PORT, nil); err != nil {
			log.Fatalf("Fail to start audio server: %v", err)
		}
	}()

	currentPersona, err := persona.GetPersona(appConfig.PersonaName, variablesProcessor)
	if err != nil {
		log.Fatalf("Failed to get current person: %v", err)
	}
	langsPath := filepath.Join(config.GetUserConfigPath(), "langs")
	lm, err := language.NewManager(langsPath, appConfig.PersonaName, appConfig.Language)
	if err != nil {
		log.Fatalf("Failed to get current person: %v", err)
	}
	state.Instance = &state.AppState{
		LastActivityTime:         time.Now(),
		ContinuousUsageStartTime: time.Now(),
		WarnedThresholds:         make(map[time.Duration]bool),
		Persona:                  currentPersona,
		Language:                 lm,
	}
	go monitorActivityLoop(state.Instance)
	if appConfig.WellbeingQuestionsEnabled {
		go schedulerLoop()
	} else {
		log.Println("Questions disabled.")
	}
	setupCustomVariables(variablesProcessor, state.Instance)
	welcomeAction := models.ActionConfig{
		Type: config.ActionSpeak,
		Text: state.Instance.Language.Get("hello_prompt"),
	}
	go actionExecutor.Execute(welcomeAction)
	select {}
}

func setupCustomVariables(processor *variables.Processor, state *state.AppState) {
	processor.RegisterHandler("level", func(context ...string) string {
		return state.Language.Get(state.Hyperfocus.Level)
	})
	processor.RegisterHandler("activity_duration", func(context ...string) string {
		usageDuration := time.Since(state.ContinuousUsageStartTime)
		activityDuration := formatDuration(usageDuration)
		return activityDuration
	})
	processor.RegisterHandler("mode", func(context ...string) string {
		if appConfig.DEBUG {
			return state.Language.Get("debug_on")
		}
		return state.Language.Get("debug_off")
	})
	processor.RegisterHandler("username", func(context ...string) string {
		return appConfig.Username
	})
	processor.RegisterHandler("person", func(context ...string) string {
		return state.Persona.GetName()
	})
	processor.RegisterHandler("date", func(context ...string) string {
		now := time.Now()
		monthName := state.Language.Get(fmt.Sprintf("months.%d", now.Month()))
		dateFormat := state.Language.Get("date_format")
		result := strings.ReplaceAll(dateFormat, "{day}", fmt.Sprintf("%d", now.Day()))
		result = strings.ReplaceAll(result, "{month}", monthName)
		result = strings.ReplaceAll(result, "{year}", fmt.Sprintf("%d", now.Year()))
		return result
	})
	processor.RegisterHandler("time", func(context ...string) string {
		loc, _ := time.LoadLocation(appConfig.TimeLocation)
		now := time.Now().In(loc)
		if appConfig.Language == "pt-br" {
			hour := now.Hour()
			min := now.Minute()
			var periodKey string
			switch {
			case hour >= 0 && hour < 6:
				periodKey = "time_periods.early_morning"
			case hour >= 6 && hour < 12:
				periodKey = "time_periods.morning"
			case hour == 12:
				periodKey = "time_periods.noon"
			case hour > 12 && hour < 18:
				periodKey = "time_periods.afternoon"
			default:
				periodKey = "time_periods.night"
			}
			period := state.Language.Get(periodKey)
			displayHour := hour
			if displayHour == 0 {
				displayHour = 12
			} else if displayHour > 12 {
				displayHour -= 12
			}
			hourWord := state.Language.Get(fmt.Sprintf("hour_words.%d", displayHour))
			if hourWord == "" || hourWord == fmt.Sprintf("!!MISSING_KEY:hour_words.%d!!", displayHour) {
				hourWord = fmt.Sprintf("%d", displayHour)
			}
			if min == 0 {
				log.Printf("São %s %s", strings.ToUpper(hourWord), period)
				return fmt.Sprintf("%s %s", hourWord, period)
			} else {
				log.Printf("São %s e %02d %s", hourWord, min, period)
				return fmt.Sprintf("%s e %02d %s", hourWord, min, period)
			}
		}
		return now.Format(state.Language.Get("time_format"))
	})
}

func setupLogger() {
	mode := ""
	if appConfig.DEBUG {
		mode = "_debug"
	}
	appConfig.DatabaseFile = filepath.Join(config.GetUserConfigPath(), fmt.Sprintf("focus_helper_%s.db", mode))
	appConfig.LogFile = filepath.Join(config.GetUserConfigPath(), fmt.Sprintf("focus_helper_%s.log", mode))
	f, err := os.OpenFile(appConfig.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Erro ao abrir arquivo de log: %v", err)
	}
	multiWriter := io.MultiWriter(os.Stdout, f)
	log.SetOutput(multiWriter)
}

func monitorActivityLoop(state *state.AppState) {
	ticker := time.NewTicker(appConfig.ActivityCheckRate.Duration)
	defer ticker.Stop()
	for range ticker.C {
		isIdle := time.Since(state.LastActivityTime) > appConfig.IdleTimeout.Duration
		if activityMonitor.HasActivity() {
			if isIdle {
				resetState(state)
			}
			state.LastActivityTime = time.Now()
		}
		if isIdle {
			continue
		}
		usageDuration := time.Since(state.ContinuousUsageStartTime)
		for _, alertLevel := range appConfig.AlertLevels {
			if alertLevel.Enabled && usageDuration >= alertLevel.Threshold.Duration && !state.WarnedThresholds[alertLevel.Threshold.Duration] {
				log.Printf("[WARNING]HYPERFOCUS DETECTED: %s (duracao: %v)", alertLevel.Level, usageDuration)
				if state.Hyperfocus == nil || state.Hyperfocus.Level != alertLevel.Level {
					state.Hyperfocus = &models.HyperfocusState{
						Level:     alertLevel.Level,
						StartTime: time.Now(),
					}
				}
				for _, action := range alertLevel.Actions {
					go actionExecutor.Execute(action)
				}
				state.WarnedThresholds[alertLevel.Threshold.Duration] = true
			}
		}
	}
}

func schedulerLoop() {
	randomDuration := time.Duration(rand.Int63n(int64(appConfig.MaxRandomQuestion.Duration-appConfig.MinRandomQuestion.Duration))) + appConfig.MinRandomQuestion.Duration
	ticker := time.NewTicker(randomDuration)
	defer ticker.Stop()
	for range ticker.C {
		askWellbeingQuestion()
		newDuration := time.Duration(rand.Int63n(int64(appConfig.MaxRandomQuestion.Duration-appConfig.MinRandomQuestion.Duration))) + appConfig.MinRandomQuestion.Duration
		ticker.Reset(newDuration)
		log.Printf("Proxima pergunta de bem-estar reagendada em %v.", newDuration.Round(time.Second))
	}
}

func askWellbeingQuestion() {
	go func() {
		questionText := "Como você está se sentindo agora, %username%? Gostaria de fazer uma pausa para o bem-estar?"
		action := models.ActionConfig{
			Type:   config.ActionSpeak,
			Prompt: questionText,
		}
		actionExecutor.Execute(action)
		answeredYes, err := notifier.Question("Pausa para o Bem-estar", "Como você está se sentindo?")
		if err != nil {
			log.Printf("Erro ao exibir pop-up de pergunta: %v", err)
			return
		}
		answer := "Não"
		if answeredYes {
			answer = "Sim"
		}
		database.LogWellbeingCheck(db, questionText, answer)
	}()
}

func resetState(state *state.AppState) {
	log.Println("User is back. Reset app state.")
	now := time.Now()
	state.ContinuousUsageStartTime = now
	state.LastActivityTime = now
	state.WarnedThresholds = make(map[time.Duration]bool)
	state.Hyperfocus = nil
	action := models.ActionConfig{
		Type:   config.ActionSpeakIA,
		Prompt: "Informe ao %username% que ele retornou da ociosidade e que seus contadores foram reiniciados.",
	}
	go actionExecutor.Execute(action)
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	} else if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func clearTempAudioOnExit() {
	tempAudioDir := filepath.Join(config.GetUserConfigPath(), config.TEMP_AUDIO_DIR)
	if _, err := os.Stat(tempAudioDir); os.IsNotExist(err) {
		log.Printf("Directory %s does not exist. Skipping deletion.", tempAudioDir)
		return
	}

	defer func() {
		err := os.RemoveAll(tempAudioDir)
		if err != nil {
			log.Printf("Error clearing temp_audio: %v", err)
		} else {
			fmt.Println("All files inside temp_audio have been cleared.")
		}
	}()

	err := filepath.Walk(tempAudioDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			err := os.Remove(path)
			if err != nil {
				log.Printf("Failed to remove file: %v", err)
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error clearing files in temp_audio: %v", err)
	}
}
