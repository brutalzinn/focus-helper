package state

import (
	"database/sql"
	"fmt"
	"focus-helper/pkg/language"
	"focus-helper/pkg/llm"
	"focus-helper/pkg/models"
	"focus-helper/pkg/notifications"
	"focus-helper/pkg/persona"
	"focus-helper/pkg/variables"
	"time"
)

type AppEvent struct {
	Type    string
	Payload any
}

type AppState struct {
	Notifier                 *notifications.DesktopNotifier
	LLMAdapter               *llm.LLMAdapter
	Hyperfocus               *models.HyperfocusState
	Language                 *language.LanguageManager
	TextProcessor            *variables.Processor
	DB                       *sql.DB
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
	Persona       persona.Persona
	Language      *language.LanguageManager
	LLMAdapter    *llm.LLMAdapter
	TextProcessor *variables.Processor
}

func NewAppState(deps AppStateDependencies) *AppState {
	return &AppState{
		EventChannel:             make(chan AppEvent, 10),
		IsActionRunning:          false,
		IsListening:              true,
		LastActivityTime:         time.Now(),
		ContinuousUsageStartTime: time.Now(),
		WarnedIndexes:            make(map[int]bool),
		SubjectFrequency:         make(map[string]int),
		Notifier:                 notifications.NewDesktopNotifier(),
		Hyperfocus:               nil,
		Language:                 deps.Language,
		Persona:                  deps.Persona,
		LLMAdapter:               deps.LLMAdapter,
		TextProcessor:            deps.TextProcessor,
	}
}

func (appState *AppState) EventLoop() {
	for event := range appState.EventChannel {
		fmt.Printf("Event Received: Type=%s\n", event.Type)

		switch event.Type {
		case "STOP_LISTENING":
			appState.IsListening = false
			fmt.Println(">> State changed: Now NOT listening for activity.")
		case "START_LISTENING":
			appState.IsListening = true
			fmt.Println(">> State changed: Resumed listening for activity.")
		}
	}
	fmt.Println("Event loop finished.")
}
