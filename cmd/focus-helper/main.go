// cmd/focus-helper/main.go
// REFACTORED: Sunday, August 10, 2025
package main

import (
	"database/sql"
	"flag"
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
	"focus-helper/pkg/variables"

	_ "github.com/mattn/go-sqlite3"
)

const (
	serverPort   = "8088"
	tempAudioDir = "temp_audio"
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

	executorDeps := actions.ExecutorDependencies{
		AppConfig:    &appConfig,
		LangManager:  language.NewManager,
		VarProcessor: variables.NewProcessor(),
		Notifier:     notifier,
		LLMAdapter:   llmAdapter,
	}
	actionExecutor = actions.NewExecutor(executorDeps)
	if err := os.MkdirAll(tempAudioDir, 0755); err != nil {
		log.Fatalf("Failed to create temp audio directory: %v", err)
	}
	fs := http.FileServer(http.Dir(tempAudioDir))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))
	go func() {
		log.Printf("Servidor de áudio rodando em http://localhost:%s", serverPort)
		if err := http.ListenAndServe(":"+serverPort, nil); err != nil {
			log.Fatalf("Falha ao iniciar servidor de áudio: %v", err)
		}
	}()

	// --- Start main application loops ---
	state := &AppState{
		lastActivityTime:         time.Now(),
		continuousUsageStartTime: time.Now(),
		warnedThresholds:         make(map[time.Duration]bool),
	}

	go monitorActivityLoop(state)
	if appConfig.WellbeingQuestionsEnabled {
		go schedulerLoop()
	} else {
		log.Println("Questoes de bem estar desativadas.")
	}

	welcomeAction := config.ActionConfig{
		Type:   config.ActionSpeak,
		Prompt: "Bem-vindo de volta, %username%. %person% na escuta. Hoje é %date%.",
	}
	go actionExecutor.Execute(welcomeAction)
	log.Println("Focus Helper esta rodando em background.")
	select {}
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
				// Execute all actions defined for this alert level
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
