package state

import (
	"focus-helper/pkg/language"
	"focus-helper/pkg/models"
	"focus-helper/pkg/persona"
	"time"
)

type AppState struct {
	LastActivityTime         time.Time
	ContinuousUsageStartTime time.Time
	WarnedThresholds         map[time.Duration]bool
	Hyperfocus               *models.HyperfocusState
	Persona                  persona.Persona
	Language                 *language.LanguageManager
}

var Instance *AppState
