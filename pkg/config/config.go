package config

import (
	"encoding/json"
	"fmt"
	"focus-helper/pkg/models"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	SERVER_PORT        = "8088"
	TEMP_AUDIO_DIR     = "temp_audio"
	ASSETS_DIR         = "assets"
	PROFILES_FILE_NAME = "profiles.json"
)

var currentConfig *models.Config

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
		IADetectorEnabled:         false,
		DatabaseFile:              "./focus_helper.db",
		LogFile:                   "./focus_helper.log",
		AlertLevels: []models.AlertLevel{
			{
				Enabled:   true,
				Level:     "low",
				Threshold: models.Duration{Duration: 45 * time.Minute},
				Actions: []models.ActionConfig{
					{Type: models.ActionSound, SoundFile: "alert_level_1.mp3"},
					{Type: models.ActionSpeakIA, Prompt: "Piloto-Alfa-Um, aqui é a Torre. Apenas um lembrete para verificar seus sistemas e fazer uma pequena pausa, se necessário."},
				},
			},
			{
				Enabled:   true,
				Level:     "medium",
				Threshold: models.Duration{Duration: 90 * time.Minute},
				Actions: []models.ActionConfig{
					{Type: models.ActionSound, SoundFile: "autopilot.mp3"},
					{Type: models.ActionSpeakIA, Prompt: "Piloto-Alfa-Um, você está em um longo período de foco. Recomendamos uma pausa para hidratação e alongamento."},
				},
			},
			{
				Enabled:   true,
				Level:     "high",
				Threshold: models.Duration{Duration: 2*time.Hour + 30*time.Minute},
				Actions: []models.ActionConfig{
					{Type: models.ActionSound, SoundFile: "alert_level_3.mp3"},
					{Type: models.ActionPopup, PopupTitle: "Alerta de Foco Intenso", PopupMessage: "Você está trabalhando continuamente por um longo período. Considere fazer uma pausa mais longa."},
					{Type: models.ActionSpeakIA, Prompt: "Piloto-Alfa-Um, detectamos sinais de hiperfoco. É crucial fazer uma pausa para manter a performance e o bem-estar."},
				},
			},
			{
				Enabled:   true,
				Level:     "CRITICAL",
				Threshold: models.Duration{Duration: 4 * time.Hour},
				Actions: []models.ActionConfig{
					{Type: models.ActionSpeakIA, Prompt: "Mayday, Mayday, Mayday. Piloto-Alfa-Um, risco de burnout detectado. Desligue o piloto automático e faça uma pausa obrigatória imediatamente."},
				},
			},
		},
	}
}

func GetUserConfigPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("Error getting user config directory: %v", err)
	}
	userConfigPath := filepath.Join(configDir, "focushelper")
	err = os.MkdirAll(userConfigPath, 0755)
	if err != nil {
		log.Fatalf("Error creating user config directory: %v", err)
	}
	return userConfigPath
}

func LoadConfig(profileName string, debug bool) (*models.Config, error) {
	profilePath := filepath.Join(GetUserConfigPath(), PROFILES_FILE_NAME)

	if debug {
		projectFolder, err := os.Getwd()
		if err != nil {
			log.Print("Error on load project folder when debug mode is enabled")
		}
		profilePath = filepath.Join(projectFolder, PROFILES_FILE_NAME)
	}
	profiles, err := LoadProfiles(profilePath)
	if err != nil {
		return nil, fmt.Errorf("error loading profiles: %w", err)
	}

	cfg, err := GetProfileByName(profiles, profileName)
	if err != nil {
		return nil, fmt.Errorf("profile '%s' not found: %w", profileName, err)
	}

	if debug {
		cfg.DEBUG = true
		log.Println("DEBUG mode enabled: Overriding time settings for faster testing.")
		cfg.MinRandomQuestion = models.Duration{Duration: 5 * time.Second}
		log.Printf("DEBUG: MinRandomQuestion set to %s", cfg.MinRandomQuestion.Duration)
		cfg.MaxRandomQuestion = models.Duration{Duration: 10 * time.Second}
		log.Printf("DEBUG: MaxRandomQuestion set to %s", cfg.MaxRandomQuestion.Duration)
		cfg.DatabaseFile = filepath.Join("focus_helper_debug.db")
		for i := range cfg.AlertLevels {
			newThreshold := time.Duration((i+1)*15) * time.Second
			cfg.AlertLevels[i].Threshold = models.Duration{Duration: newThreshold}
			log.Printf("DEBUG: Alert level '%s' threshold set to %s", cfg.AlertLevels[i].Level, cfg.AlertLevels[i].Threshold.Duration)
		}
	}
	currentConfig = cfg
	return cfg, nil
}

func GetConfig() *models.Config {
	return currentConfig
}
