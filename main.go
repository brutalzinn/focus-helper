package main

import (
	"database/sql"
	"flag"
	"io"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/brutalzinn/focus-helper/actions"
	"github.com/brutalzinn/focus-helper/activity"
	"github.com/brutalzinn/focus-helper/audio"
	"github.com/brutalzinn/focus-helper/config"
	"github.com/brutalzinn/focus-helper/database"
	"github.com/brutalzinn/focus-helper/integrations"
	"github.com/brutalzinn/focus-helper/notifications"
)

var appConfig config.Config
var db *sql.DB
var activityMonitor *activity.Monitor
var atcPromptManager *integrations.PromptManager

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
	if appConfig.DEBUG {
		log.Println("!!!!!!!!!! RODANDO EM MODO DEBUG !!!!!!!!!!")
	}

	var err error
	db, err = database.Init(appConfig.DatabaseFile)
	if err != nil {
		log.Fatalf("Falha ao inicializar banco de dados: %v", err)
	}
	defer db.Close()

	audio.InitSpeaker()
	activityMonitor = activity.NewMonitor()
	atcPromptManager = integrations.NewATCPromptManager()

	state := &AppState{
		lastActivityTime:         time.Now(),
		continuousUsageStartTime: time.Now(),
		warnedThresholds:         make(map[time.Duration]bool),
		currentHyperfocusState:   nil,
	}

	go monitorActivityLoop(state)
	if appConfig.WellbeingQuestionsEnabled {
		go schedulerLoop()
	} else {
		log.Println("Questões de bem estar desativadas.")
	}

	audio.PlayRadioSimulation("Bem-vindo ao Focus Helper. Estamos prontos para ajudar você a manter o foco e o bem-estar.", 1, 1, 1, "radio_static.wav")
	log.Println("Focus Helper está rodando em background.")
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
	ticker := time.NewTicker(config.AppConfig.ActivityCheckRate)
	defer ticker.Stop()
	for range ticker.C {
		isIdle := time.Since(state.lastActivityTime) > config.AppConfig.IdleTimeout
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
		for _, level := range config.AppConfig.AlertLevels {
			if level.Enabled && usageDuration >= level.Threshold && !state.warnedThresholds[level.Threshold] {
				log.Printf("Alerta de hiperfoco acionado: %s (duração: %v)", level.Level, usageDuration)
				if state.currentHyperfocusState == nil || state.currentHyperfocusState.Level != level.Level {
					state.currentHyperfocusState = &config.HyperfocusState{
						Level:     level.Level,
						StartTime: time.Now(),
					}
				}
				go actions.Execute(level, state.currentHyperfocusState)
				state.warnedThresholds[level.Threshold] = true
			}
		}
	}
}

func schedulerLoop() {
	randomDuration := time.Duration(rand.Int63n(int64(config.AppConfig.MaxRandomQuestion-config.AppConfig.MinRandomQuestion))) + config.AppConfig.MinRandomQuestion
	ticker := time.NewTicker(randomDuration)
	log.Printf("Próxima pergunta de bem-estar agendada em %v.", randomDuration.Round(time.Second))
	defer ticker.Stop()
	for range ticker.C {
		askWellbeingQuestion()
		newDuration := time.Duration(rand.Int63n(int64(config.AppConfig.MaxRandomQuestion-config.AppConfig.MinRandomQuestion))) + config.AppConfig.MinRandomQuestion
		ticker.Reset(newDuration)
		log.Printf("Próxima pergunta de bem-estar reagendada em %v.", newDuration.Round(time.Second))
	}
}

func askWellbeingQuestion() {
	go func() {
		finalPrompt := atcPromptManager.FormatPrompt("Como você está se sentindo agora? Você gostaria de fazer uma pausa para o bem-estar?")
		questionText, err := integrations.GenerateTextWithLlama(config.AppConfig.Llama.Model, finalPrompt)
		if err != nil {
			log.Printf("Erro ao gerar pergunta com Llama, usando fallback: %v", err)
			questionText = "Que tal uma pausa para um copo d'água?"
		}
		answeredYes := notifications.ShowQuestionPopup("Pausa para o Bem-estar", questionText)
		answer := "Não"
		if answeredYes {
			answer = "Sim"
		}
		database.LogWellbeingCheck(db, questionText, answer)
	}()
}

func resetState(state *AppState) {
	log.Println("Usuário retornou da ociosidade. Reiniciando contadores.")

	now := time.Now()
	state.continuousUsageStartTime = now
	state.lastActivityTime = now
	state.warnedThresholds = make(map[time.Duration]bool)
	state.currentHyperfocusState = nil
	prompt := integrations.NewATCPromptManager()
	text := prompt.FormatPrompt("Informe ao Alfa-Um que ele retornou da ociosidade e que seus contadores foram reiniciados.")
	response, err := integrations.GenerateTextWithLlama(config.AppConfig.Llama.Model, text)
	if err != nil {
		log.Printf("Erro ao gerar resposta com Llama: %v", err)
		response = "Usuário ativo novamente."
	}
	go audio.PlayRadioSimulation(response, 1, 1, 1, "radio_static.wav")
}
