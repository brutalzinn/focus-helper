package config

import (
	"log"
	"time"
)

var AppConfig Config

type ActionType string

const (
	ActionPopup         ActionType = "POPUP"
	ActionSound         ActionType = "SOUND"
	ActionATC           ActionType = "ATC_VOICE"
	ActionHomeAssistant ActionType = "HOME_ASSISTANT"
)

type ActionConfig struct {
	Type             ActionType          `json:"type"`
	RandomChance     float64             `json:"random_chance,omitempty"`
	BackgroundVolume float64             `json:"background_volume,omitempty"`
	BackgroundFile   string              `json:"background_file,omitempty"`
	VoiceVolume      float64             `json:"voice_volume,omitempty"`
	LlamaPrompt      string              `json:"llama_prompt,omitempty"`
	SoundFile        string              `json:"sound_file,omitempty"`
	PopupTitle       string              `json:"popup_title,omitempty"`
	PopupMessage     string              `json:"popup_message,omitempty"`
	HomeAssistant    HomeAssistantConfig `json:"home_assistant,omitempty"`
}

type AlertLevel struct {
	Enabled              bool
	Level                string
	Multiplier           float64 `json:"multiplier,omitempty"`
	Threshold            time.Duration
	TriggerHomeAssistant bool
	Actions              []ActionConfig
}

type LlamaConfig struct {
	Model      string `json:"model"`
	BasePrompt string `json:"base_prompt"`
}

type HomeAssistantConfig struct {
	Enabled    bool   `json:"enabled"`
	WebhookURL string `json:"webhook_url"`
}

type HyperfocusState struct {
	Level     string
	StartTime time.Time
	EndTime   time.Time
}
type MiscConfig struct {
	WarnedThresholds       map[time.Duration]bool
	CurrentHyperfocusState *HyperfocusState
}

type Config struct {
	DEBUG                     bool
	IdleTimeout               time.Duration
	ActivityCheckRate         time.Duration
	MinRandomQuestion         time.Duration
	MaxRandomQuestion         time.Duration
	DatabaseFile              string
	LogFile                   string
	Llama                     LlamaConfig
	HomeAssistant             HomeAssistantConfig
	WellbeingQuestionsEnabled bool
	ReduceOSSounds            bool
	Misc                      MiscConfig
	AlertLevels               []AlertLevel
}

func Init(debugMode bool) Config {
	log.Println("Initializing configuration...")
	if debugMode {
		AppConfig = loadDebugConfig()
		log.Println("DEBUG configuration loaded.")
	} else {
		AppConfig = loadProdConfig()
		log.Println("PRODUCTION configuration loaded.")
	}
	AppConfig.Misc.WarnedThresholds = make(map[time.Duration]bool)
	AppConfig.Misc.CurrentHyperfocusState = nil
	return AppConfig
}

func loadProdConfig() Config {
	return Config{
		DEBUG:             false,
		IdleTimeout:       3 * time.Minute,  // Tempo para considerar o usuário ocioso
		ActivityCheckRate: 30 * time.Second, // Verificação menos frequente para economizar recursos
		ReduceOSSounds:    true,

		MinRandomQuestion: 45 * time.Minute,
		MaxRandomQuestion: 90 * time.Minute,
		DatabaseFile:      "./focus_helper.db",
		LogFile:           "./focus_helper.log",
		Llama: LlamaConfig{
			Model: "llama3.2:latest",
		},
		AlertLevels: []AlertLevel{
			{
				Enabled:   true,
				Level:     "LOW",
				Threshold: 45 * time.Minute,
				Actions: []ActionConfig{
					{Type: ActionSound, SoundFile: "alert_level_1.mp3"},
					{Type: ActionATC, VoiceVolume: 1.0, BackgroundVolume: 0.3, BackgroundFile: "radio_static.wav", LlamaPrompt: "Piloto-Alfa-Um, aqui é a Torre. Apenas um lembrete para verificar seus sistemas e fazer uma pequena pausa, se necessário."},
				},
			},
			{
				Enabled:    true,
				Level:      "MEDIUM",
				Threshold:  90 * time.Minute,
				Multiplier: 1.5,
				Actions: []ActionConfig{
					{Type: ActionSound, SoundFile: "autopilot.mp3"},
					{Type: ActionATC, VoiceVolume: 1.0, BackgroundVolume: 0.4, BackgroundFile: "radio_static.wav", LlamaPrompt: "Piloto-Alfa-Um, você está em um longo período de foco. Recomendamos uma pausa para hidratação e alongamento."},
				},
			},
			{
				Enabled:    true,
				Level:      "HIGH",
				Threshold:  2*time.Hour + 30*time.Minute,
				Multiplier: 2.5,
				Actions: []ActionConfig{
					{Type: ActionSound, SoundFile: "alert_level_3.mp3"},
					{Type: ActionPopup, PopupTitle: "Alerta de Foco Intenso", PopupMessage: "Você está trabalhando continuamente por um longo período. Considere fazer uma pausa mais longa."},
					{Type: ActionATC, VoiceVolume: 1.0, BackgroundVolume: 0.5, BackgroundFile: "radio_static.wav", LlamaPrompt: "Piloto-Alfa-Um, detectamos sinais de hiperfoco. É crucial fazer uma pausa para manter a performance e o bem-estar."},
				},
			},
			{
				Enabled:    true,
				Level:      "CRITICAL",
				Threshold:  4 * time.Hour,
				Multiplier: 5.0,
				Actions: []ActionConfig{
					{Type: ActionATC, VoiceVolume: 1.0, BackgroundVolume: 1.0, BackgroundFile: "radio_static.wav", LlamaPrompt: "Mayday, Mayday, Mayday. Piloto-Alfa-Um, risco de burnout detectado. Desligue o piloto automático e faça uma pausa obrigatória imediatamente."}, // <-- Adicionado VoiceVolume
				},
			},
		},
	}
}
func loadDebugConfig() Config {
	return Config{
		DEBUG:             true,
		IdleTimeout:       30 * time.Second,
		ActivityCheckRate: 5 * time.Second,
		MinRandomQuestion: 30 * time.Second,
		MaxRandomQuestion: 60 * time.Second,
		ReduceOSSounds:    true,
		Llama: LlamaConfig{
			Model:      "llama3.2:latest",
			BasePrompt: "Piloto-Alfa-Um, você está em uma missão de foco intenso. Mantenha a calma e siga as instruções da torre.",
		},
		WellbeingQuestionsEnabled: false,
		DatabaseFile:              "./focus_helper_debug.db",
		LogFile:                   "./focus_helper_debug.log",
		AlertLevels: []AlertLevel{
			{
				Enabled:    true,
				Level:      "LOW",
				Threshold:  10 * time.Second,
				Multiplier: 1.0,
				Actions: []ActionConfig{
					{Type: ActionSound, SoundFile: "alert_level_1.mp3"},
					{Type: ActionATC, VoiceVolume: 1.0, BackgroundFile: "radio_static.wav", BackgroundVolume: 1.0, LlamaPrompt: "Piloto-Alfa-Um detecção de Windshear ou hiperfoco próximo."}, // <-- VoiceVolume ajustado
				},
			},
			{
				Enabled:    true,
				Level:      "MEDIUM",
				Threshold:  25 * time.Second,
				Multiplier: 1.5,
				Actions: []ActionConfig{
					{Type: ActionSound, SoundFile: "autopilot.mp3"},
					{Type: ActionATC, VoiceVolume: 1.2, BackgroundFile: "radio_static.wav", BackgroundVolume: 1.0, LlamaPrompt: "Piloto-Alfa-Um você está em hipertoco. Solicito que siga para o próximo aeroporto cozinha e solicite ajuda."}, // <-- Adicionado VoiceVolume
				},
			},
			{
				Enabled:    true,
				Level:      "HIGH",
				Threshold:  45 * time.Second,
				Multiplier: 2.0,
				Actions: []ActionConfig{
					{Type: ActionSound, SoundFile: "alert_level_3.mp3"},
					{Type: ActionPopup, PopupTitle: "Teste de alerta nível 3", PopupMessage: "Siga as instruções da torre de comando imediatamente!"},
					{Type: ActionATC, VoiceVolume: 1.5, BackgroundFile: "radio_static.wav", BackgroundVolume: 1.0, LlamaPrompt: "Piloto-Alfa-Um está perdendo o controle. Siga as instruções na tela."},                    
					{Type: ActionATC, VoiceVolume: 1.5, BackgroundFile: "radio_static.wav", BackgroundVolume: 1.0, LlamaPrompt: "Piloto-Alfa-Um você deve prosseguir para o aeroporto mais próximo e pousar imediatamente."},
				},
			},
			{
				Enabled:    true,
				Level:      "CRITICAL",
				Threshold:  60 * time.Second,
				Multiplier: 5.0,
				Actions: []ActionConfig{
					{Type: ActionSound, SoundFile: "autopilot.mp3"},
					{Type: ActionATC, VoiceVolume: 2.0, BackgroundVolume: 1.0, BackgroundFile: "radio_static.wav", LlamaPrompt: "Piloto-Alfa-Um você perdeu o controle. Siga as instruções na tela."},                 
					{Type: ActionATC, VoiceVolume: 2.0, BackgroundVolume: 1.0, BackgroundFile: "radio_static.wav", LlamaPrompt: "Piloto-Alfa-Um solicito que desligue o piloto automático e siga as ordens da torre."}, 
				},
			},
		},
	}
}
