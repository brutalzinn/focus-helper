package sheduler

import (
	"context"
	"focus-helper/src/pkg/actions"
	"focus-helper/src/pkg/database"
	"focus-helper/src/pkg/models"
	"focus-helper/src/pkg/state"
	"log"
	"math/rand"
	"sync"
	"time"
)

func SchedulerLoop(ctx context.Context, wg *sync.WaitGroup, appState *state.AppState) {
	defer wg.Done()
	log.Println("Wellbeing scheduler started.")
	randomDuration := time.Duration(rand.Int63n(int64(appState.AppConfig.MaxRandomQuestion.Duration-appState.AppConfig.MinRandomQuestion.Duration))) + appState.AppConfig.MinRandomQuestion.Duration
	ticker := time.NewTicker(randomDuration)
	defer ticker.Stop()
	log.Printf("Next wellbeing question in %v.", randomDuration.Round(time.Second))
	for {
		select {
		case <-ticker.C:
			askWellbeingQuestion(appState)
			newDuration := time.Duration(rand.Int63n(int64(appState.AppConfig.MaxRandomQuestion.Duration-appState.AppConfig.MinRandomQuestion.Duration))) + appState.AppConfig.MinRandomQuestion.Duration
			ticker.Reset(newDuration)
			log.Printf("Next wellbeing question in %v.", newDuration.Round(time.Second))
		case <-ctx.Done():
			log.Println("Stopping wellbeing scheduler due to shutdown signal...")
			return
		}
	}
}

func askWellbeingQuestion(appState *state.AppState) {
	log.Println("Asking a wellbeing question...")

	questionText := appState.Language.Get("ask_sheduler")
	action := models.ActionConfig{
		Type:   models.ActionSpeak,
		Prompt: questionText,
	}
	if err := actions.Execute(action); err != nil {
		log.Printf("Cant execute this action %v: %v", models.ActionSpeak, err)
	}
	answeredYes, err := appState.Notifier.Question("Wellbeing Break", "How are you feeling?")
	if err != nil {
		log.Printf("Error displaying question popup: %v", err)
		return
	}
	answer := "No"
	if answeredYes {
		answer = "Yes"
	}
	database.LogWellbeingCheck(appState.DB, questionText, answer)
}
