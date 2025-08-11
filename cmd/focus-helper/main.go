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
	"time"

	"focus-helper/pkg/actions"
	"focus-helper/pkg/activity"
	"focus-helper/pkg/config"
	"focus-helper/pkg/database"
	"focus-helper/pkg/language"
	"focus-helper/pkg/llm"
	"focus-helper/pkg/notifications"
	"focus-helper/pkg/persona"
	"focus-helper/pkg/variables"

	_ "github.com/mattn/go-sqlite3"
)

// Global variables for core services
var (
	appConfig       config.Config
	db              *sql.DB
	activityMonitor *activity.Monitor
	actionExecutor  *actions.Executor
	notifier        notifications.Notifier
)

// AppState remains the same
type AppState struct {
	lastActivityTime         time.Time
	continuousUsageStartTime time.Time
	warnedThresholds         map[time.Duration]bool
	currentHyperfocusState   *config.HyperfocusState
	currentPersona           persona.Persona
	currentLanguage          *language.LanguageManager
}

func main() {
	debugFlag := flag.Bool("debug", false, "Set to true to enable debug mode")
	flag.Parse()
	appConfig = config.Init(*debugFlag)
	setupLogger()

	log.Println("--- Iniciando o Focus Helper ---")
	log.Printf("PERSONA ATIVA: %s", appConfig.PersonaName)
	if appConfig.DEBUG {
		log.Println("!!!!!!!!!! RODANDO EM MODO DEBUG !!!!!!!!!!")
	}

	var err error
	db, err = database.Init(appConfig.DatabaseFile)
	if err != nil {
		log.Fatalf("Falha ao inicializar banco de dados: %v", err)
	}
	defer db.Close()

	activityMonitor = activity.NewMonitor()
	notifier = notifications.NewDesktopNotifier()

	llmAdapter, err := llm.NewAdapter(appConfig.IAModel)
	if err != nil {
		log.Fatalf("FATAL: Nao foi possivel criar o adaptador LLM: %v", err)
	}
	variablesProcessor := variables.NewProcessor()

	executorDeps := actions.ExecutorDependencies{
		AppConfig:    &appConfig,
		LangManager:  language.NewManager,
		VarProcessor: variablesProcessor,
		Notifier:     notifier,
		LLMAdapter:   llmAdapter,
	}
	actionExecutor = actions.NewExecutor(executorDeps)
	if err := os.MkdirAll(config.TEMP_AUDIO_DIR, 0755); err != nil {
		log.Fatalf("Failed to create temp audio directory: %v", err)
	}
	fs := http.FileServer(http.Dir(config.TEMP_AUDIO_DIR))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))
	go func() {
		log.Printf("Servidor de áudio rodando em http://localhost:%s", config.SERVER_PORT)
		if err := http.ListenAndServe(":"+config.SERVER_PORT, nil); err != nil {
			log.Fatalf("Falha ao iniciar servidor de áudio: %v", err)
		}
	}()

	currentPersona, err := persona.GetPersona(config.AppConfig.PersonaName, variablesProcessor)
	if err != nil {
		log.Fatalf("Failed to get current person: %v", err)
	}
	lm, err := language.NewManager("pkg/language", config.AppConfig.PersonaName, config.AppConfig.Language)
	if err != nil {
		log.Fatalf("Failed to get current person: %v", err)
	}
	state := &AppState{
		lastActivityTime:         time.Now(),
		continuousUsageStartTime: time.Now(),
		warnedThresholds:         make(map[time.Duration]bool),
		currentPersona:           currentPersona,
		currentLanguage:          lm,
	}

	go monitorActivityLoop(state)
	if appConfig.WellbeingQuestionsEnabled {
		go schedulerLoop()
	} else {
		log.Println("Questoes de bem estar desativadas.")
	}

	setupCustomVariables(variablesProcessor, state)

	welcomeAction := config.ActionConfig{
		Type: config.ActionSpeak,
		Text: "Bem-vindo de volta, %username%. %person% na escuta",
	}
	go actionExecutor.Execute(welcomeAction)
	select {}
}

func setupCustomVariables(processor *variables.Processor, state *AppState) {

	processor.RegisterHandler("level", func(context ...string) string {
		return state.currentHyperfocusState.Level
	})
	processor.RegisterHandler("activity_duration", func(context ...string) string {
		usageDuration := time.Since(state.continuousUsageStartTime)
		activityDuration := formatDuration(usageDuration)
		return activityDuration
	})
	processor.RegisterHandler("username", func(context ...string) string {
		return config.AppConfig.Username
	})
	processor.RegisterHandler("person", func(context ...string) string {
		return state.currentPersona.GetName()
	})
	processor.RegisterHandler("date", func(context ...string) string {
		now := time.Now()
		monthsPtBr := []string{"", "janeiro", "fevereiro", "março", "abril", "maio", "junho", "julho", "agosto", "setembro", "outubro", "novembro", "dezembro"}
		return fmt.Sprintf("%d de %s de %d", now.Day(), monthsPtBr[now.Month()], now.Year())
	})
	processor.RegisterHandler("time", func(context ...string) string {
		loc, _ := time.LoadLocation("America/Sao_Paulo")
		return time.Now().In(loc).Format("3:04 PM")
	})
}

func setupLogger() {
	f, err := os.OpenFile(appConfig.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Erro ao abrir arquivo de log: %v", err)
	}
	multiWriter := io.MultiWriter(os.Stdout, f)
	log.SetOutput(multiWriter)
}

func monitorActivityLoop(state *AppState) {
	ticker := time.NewTicker(appConfig.ActivityCheckRate)
	defer ticker.Stop()
	for range ticker.C {
		isIdle := time.Since(state.lastActivityTime) > appConfig.IdleTimeout
		if activityMonitor.HasActivity() {
			if isIdle {
				resetState(state)
			}
			state.lastActivityTime = time.Now()
		}
		if isIdle {
			continue
		}
		usageDuration := time.Since(state.continuousUsageStartTime)
		for _, level := range appConfig.AlertLevels {
			if level.Enabled && usageDuration >= level.Threshold && !state.warnedThresholds[level.Threshold] {
				log.Printf("Alerta de hiperfoco acionado: %s (duracao: %v)", level.Level, usageDuration)
				if state.currentHyperfocusState == nil || state.currentHyperfocusState.Level != level.Level {
					state.currentHyperfocusState = &config.HyperfocusState{
						Level:     level.Level,
						StartTime: time.Now(),
					}
				}
				for _, action := range level.Actions {
					go actionExecutor.Execute(action)
				}
				state.warnedThresholds[level.Threshold] = true
			}
		}
	}
}

func schedulerLoop() {
	randomDuration := time.Duration(rand.Int63n(int64(appConfig.MaxRandomQuestion-appConfig.MinRandomQuestion))) + appConfig.MinRandomQuestion
	ticker := time.NewTicker(randomDuration)
	log.Printf("Proxima pergunta de bem-estar agendada em %v.", randomDuration.Round(time.Second))
	defer ticker.Stop()
	for range ticker.C {
		askWellbeingQuestion()
		newDuration := time.Duration(rand.Int63n(int64(appConfig.MaxRandomQuestion-appConfig.MinRandomQuestion))) + appConfig.MinRandomQuestion
		ticker.Reset(newDuration)
		log.Printf("Proxima pergunta de bem-estar reagendada em %v.", newDuration.Round(time.Second))
	}
}

func askWellbeingQuestion() {
	go func() {
		questionText := "Como você está se sentindo agora, %username%? Gostaria de fazer uma pausa para o bem-estar?"
		action := config.ActionConfig{
			Type:   config.ActionSpeak,
			Prompt: questionText,
		}
		// The executor handles speaking the question
		actionExecutor.Execute(action)

		// The notifier handles showing the popup
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

func resetState(state *AppState) {
	log.Println("Usuario retornou da ociosidade. Reiniciando contadores.")
	now := time.Now()
	state.continuousUsageStartTime = now
	state.lastActivityTime = now
	state.warnedThresholds = make(map[time.Duration]bool)
	state.currentHyperfocusState = nil

	action := config.ActionConfig{
		Type:   config.ActionSpeak,
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
