package config

import (
	"encoding/json"
	"fmt"
	"focus-helper/pkg/models"
	"os"
	"time"
)

const (
	ActionPopup   models.ActionType = "POPUP"
	ActionSound   models.ActionType = "SOUND"
	ActionSpeak   models.ActionType = "SPEAK_VOICE"
	ActionSpeakIA models.ActionType = "SPEAK_IA"
	ActionWebHook models.ActionType = "WEBHOOK"
)

const (
	SERVER_PORT    = "8088"
	TEMP_AUDIO_DIR = "temp_audio"
)

func LoadProfiles(filename string) ([]models.Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error to open profile: %w", err)
	}
	defer file.Close()
	var profiles []models.Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&profiles); err != nil {
		return nil, fmt.Errorf("error to decode profiles: %w", err)
	}
	return profiles, nil
}

func GetProfileByName(profiles []models.Config, name string) (*models.Config, error) {
	for _, p := range profiles {
		if p.ProfileName == name {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("perfil '%s' não encontrado", name)
}

func DefaultProfile() models.Config {
	return models.Config{
		ProfileName:  "atc_tower",
		TimeLocation: "America/Sao_Paulo",
		PersonaName:  "atc_tower",
		Language:     "pt-br",
		Username:     "Piloto-Alfa-Um",
		IAModel: models.IAModel{
			Type:  "ollama",
			Model: "llama3.2",
			URL:   "http://localhost:11434/api/generate",
		},
		IdleTimeout:               models.Duration{Duration: 3 * time.Minute}, // Explicitly wrap time.Duration
		ActivityCheckRate:         models.Duration{Duration: 30 * time.Second},
		MinRandomQuestion:         models.Duration{Duration: 45 * time.Minute},
		MaxRandomQuestion:         models.Duration{Duration: 90 * time.Minute},
		WellbeingQuestionsEnabled: true,
		ReduceOSSounds:            true,
		DatabaseFile:              "./focus_helper.db",
		LogFile:                   "./focus_helper.log",
		AlertLevels: []models.AlertLevel{
			{
				Enabled:   true,
				Level:     "low",
				Threshold: models.Duration{Duration: 45 * time.Minute},
				Actions: []models.ActionConfig{
					{Type: ActionSound, SoundFile: "alert_level_1.mp3"},
					{Type: ActionSpeakIA, Prompt: "Piloto-Alfa-Um, aqui é a Torre. Apenas um lembrete para verificar seus sistemas e fazer uma pequena pausa, se necessário."},
				},
			},
			{
				Enabled:    true,
				Level:      "medium",
				Threshold:  models.Duration{Duration: 90 * time.Minute},
				Multiplier: 1.5,
				Actions: []models.ActionConfig{
					{Type: ActionSound, SoundFile: "autopilot.mp3"},
					{Type: ActionSpeakIA, Prompt: "Piloto-Alfa-Um, você está em um longo período de foco. Recomendamos uma pausa para hidratação e alongamento."},
				},
			},
			{
				Enabled:   true,
				Level:     "high",
				Threshold: models.Duration{Duration: 2*time.Hour + 30*time.Minute},

				Multiplier: 2.5,
				Actions: []models.ActionConfig{
					{Type: ActionSound, SoundFile: "alert_level_3.mp3"},
					{Type: ActionPopup, PopupTitle: "Alerta de Foco Intenso", PopupMessage: "Você está trabalhando continuamente por um longo período. Considere fazer uma pausa mais longa."},
					{Type: ActionSpeakIA, Prompt: "Piloto-Alfa-Um, detectamos sinais de hiperfoco. É crucial fazer uma pausa para manter a performance e o bem-estar."},
				},
			},
			{
				Enabled:    true,
				Level:      "CRITICAL",
				Threshold:  models.Duration{Duration: 4 * time.Hour},
				Multiplier: 5.0,
				Actions: []models.ActionConfig{
					{Type: ActionSpeakIA, Prompt: "Mayday, Mayday, Mayday. Piloto-Alfa-Um, risco de burnout detectado. Desligue o piloto automático e faça uma pausa obrigatória imediatamente."},
				},
			},
		},
	}
}
