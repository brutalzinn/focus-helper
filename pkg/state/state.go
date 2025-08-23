package state

import (
	"context"
	"database/sql"
	"fmt"
	"focus-helper/pkg/language"
	"focus-helper/pkg/llm"
	"focus-helper/pkg/models"
	"focus-helper/pkg/notifications"
	"focus-helper/pkg/persona"
	"focus-helper/pkg/variables"
	"sync"
	"time"
)

var currentState *AppState

type AppEvent struct {
	Type    string
	Payload any
}

type AppState struct {
	Notifier                 *notifications.DesktopNotifier
	LLMAdapter               llm.LLMAdapter
	Hyperfocus               *models.HyperfocusState
	Language                 *language.LanguageManager
	VarProcessor             *variables.Processor
	DB                       *sql.DB
	AppConfig                *models.Config
	Persona                  persona.Persona
	IsListening              bool
	IsActionRunning          bool
	WarnedIndexes            map[int]bool
	SubjectFrequency         map[string]int
	EventChannel             chan AppEvent
	LastActivityTime         time.Time
	ContinuousUsageStartTime time.Time
}

type AppStateDependencies struct {
	AppConfig    *models.Config
	Persona      persona.Persona
	DB           *sql.DB
	Notifier     *notifications.DesktopNotifier
	Language     *language.LanguageManager
	LLMAdapter   llm.LLMAdapter
	VarProcessor *variables.Processor
}

func NewAppState(deps AppStateDependencies) *AppState {
	currentState = &AppState{
		EventChannel:             make(chan AppEvent, 10),
		WarnedIndexes:            make(map[int]bool),
		SubjectFrequency:         make(map[string]int),
		Hyperfocus:               nil,
		LastActivityTime:         time.Now(),
		ContinuousUsageStartTime: time.Now(),
		IsActionRunning:          false,
		IsListening:              true,
		Notifier:                 deps.Notifier,
		AppConfig:                deps.AppConfig,
		Language:                 deps.Language,
		Persona:                  deps.Persona,
		LLMAdapter:               deps.LLMAdapter,
		VarProcessor:             deps.VarProcessor,
		DB:                       deps.DB,
	}
	return currentState
}

func GetAppState() *AppState {
	return currentState
}

func (appState *AppState) EventLoop(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("Event loop started.")
	for {
		select {
		case event, ok := <-appState.EventChannel:
			if !ok {
				fmt.Println("Event channel closed. Exiting event loop.")
				return
			}
			fmt.Printf("Event Received: Type=%s\n", event.Type)
			switch event.Type {
			case "STOP_LISTENING":
				appState.IsListening = false
				fmt.Println(">> State changed: Now NOT listening for activity.")
			case "START_LISTENING":
				appState.IsListening = true
				fmt.Println(">> State changed: Resumed listening for activity.")
			}
		case <-ctx.Done():
			fmt.Println("Shutdown signal received. Exiting event loop.")
			return
		}
	}
}
