package main

import (
	"database/sql"
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
var lastActivityTime time.Time
var continuousUsageStartTime time.Time
var activityMonitor *activity.Monitor

func main() {
	appConfig = config.LoadConfig(true)

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

	now := time.Now()
	lastActivityTime = now
	continuousUsageStartTime = now

	go monitorActivityLoop()
	if appConfig.WellbeingQuestionsEnabled {
		go schedulerLoop()
	} else {
		log.Println("Questões de bem estar desativadas.")
	}

	audio.PlayRadioSimulation("Bem-vindo ao Focus Helper. Estamos prontos para ajudar você a manter o foco e o bem-estar.", 1, 1)
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

func monitorActivityLoop() {
	ticker := time.NewTicker(appConfig.ActivityCheckRate)
	defer ticker.Stop()
	warnedThresholds := make(map[time.Duration]bool)

	for range ticker.C {
		if activityMonitor.HasActivity() {
			if time.Since(lastActivityTime) > appConfig.IdleTimeout {
				log.Println("Usuário retornou da ociosidade. Reiniciando contadores.")
				continuousUsageStartTime = time.Now()
				warnedThresholds = make(map[time.Duration]bool)
			}
			lastActivityTime = time.Now()
		}

		if time.Since(lastActivityTime) > appConfig.IdleTimeout {
			continue
		}
		usageDuration := time.Since(continuousUsageStartTime)
		for _, level := range appConfig.AlertLevels {
			if level.Enabled && usageDuration > level.Threshold && !warnedThresholds[level.Threshold] {
				go actions.Execute(level, appConfig)
				warnedThresholds[level.Threshold] = true
			}
		}
	}
}

func schedulerLoop() {
	randomDuration := time.Duration(rand.Int63n(int64(appConfig.MaxRandomQuestion-appConfig.MinRandomQuestion))) + appConfig.MinRandomQuestion
	ticker := time.NewTicker(randomDuration)
	log.Printf("Próxima pergunta de bem-estar agendada em %v.", randomDuration.Round(time.Second))
	defer ticker.Stop()

	for range ticker.C {
		askWellbeingQuestion()
		newDuration := time.Duration(rand.Int63n(int64(appConfig.MaxRandomQuestion-appConfig.MinRandomQuestion))) + appConfig.MinRandomQuestion
		ticker.Reset(newDuration)
		log.Printf("Próxima pergunta de bem-estar reagendada em %v.", newDuration.Round(time.Second))
	}
}

func askWellbeingQuestion() {
	go func() {
		prompt := "Você é um assistente de bem-estar. Gere UMA ÚNICA pergunta curta de confirmação para um usuário de computador sobre pausas ou saúde (postura, água, olhos). Sem aspas."
		questionText, err := integrations.GenerateTextWithLlama(appConfig.LlamaModel, prompt)
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
