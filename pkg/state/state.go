package state

import (
	"focus-helper/pkg/language"
	"focus-helper/pkg/llm"
	"focus-helper/pkg/models"
	"focus-helper/pkg/persona"
	"time"
)

type AppState struct {
	LastActivityTime         time.Time
	ContinuousUsageStartTime time.Time
	LLMAdapter               *llm.LLMAdapter
	Hyperfocus               *models.HyperfocusState
	Persona                  persona.Persona
	Language                 *language.LanguageManager
	WarnedIndexes            map[int]bool
	SubjectFrequency         map[string]int
}

func NewAppState() *AppState {
	return &AppState{
		LastActivityTime:         time.Now(),
		ContinuousUsageStartTime: time.Now(),
		WarnedIndexes:            make(map[int]bool),
		SubjectFrequency:         make(map[string]int),
		Hyperfocus:               nil,
		LLMAdapter:               nil,
	}
}
