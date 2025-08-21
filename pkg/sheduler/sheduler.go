package sheduler

import (
	"database/sql"
	"focus-helper/pkg/actions"
	"focus-helper/pkg/database"
	"focus-helper/pkg/models"
	"focus-helper/pkg/notifications"
	"log"
	"math/rand"
	"time"
)

func SchedulerLoop(appConfig *models.Config, db *sql.DB, actionExecutor *actions.Executor, notifier notifications.Notifier) {
	randomDuration := time.Duration(rand.Int63n(int64(appConfig.MaxRandomQuestion.Duration-appConfig.MinRandomQuestion.Duration))) + appConfig.MinRandomQuestion.Duration
	ticker := time.NewTicker(randomDuration)
	defer ticker.Stop()

	for range ticker.C {
		askWellbeingQuestion(db, actionExecutor, notifier)
		newDuration := time.Duration(rand.Int63n(int64(appConfig.MaxRandomQuestion.Duration-appConfig.MinRandomQuestion.Duration))) + appConfig.MinRandomQuestion.Duration
		ticker.Reset(newDuration)
		log.Printf("Next wellbeing question in %v.", newDuration.Round(time.Second))
	}
}

func askWellbeingQuestion(db *sql.DB, actionExecutor *actions.Executor, notifier notifications.Notifier) {
	go func() {
		questionText := "How are you feeling right now, %username%? Would you like to take a wellbeing break?"
		action := models.ActionConfig{
			Type:   models.ActionSpeak,
			Prompt: questionText,
		}
		actionExecutor.Execute(action)

		answeredYes, err := notifier.Question("Wellbeing Break", "How are you feeling?")
		if err != nil {
			log.Printf("Error displaying question popup: %v", err)
			return
		}
		answer := "No"
		if answeredYes {
			answer = "Yes"
		}
		database.LogWellbeingCheck(db, questionText, answer)
	}()
}
