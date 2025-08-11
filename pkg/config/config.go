package config

import (
	"log"
	"time"
)

var AppConfig Config

type ActionType string

const (
	ActionPopup   ActionType = "POPUP"
	ActionSound   ActionType = "SOUND"
	ActionSpeak   ActionType = "SPEAK_VOICE"
	ActionWebHook ActionType = "WEBHOOK"
)

type ActionConfig struct {
	Type         ActionType `json:"type"`
	RandomChance float64    `json:"random_chance,omitempty"`
	SoundFile    string     `json:"sound_file,omitempty"`
	Volume       float64    `json:"volume,omitempty"`
	///IA
	Prompt string `json:"llama_prompt,omitempty"`
	//dialogs
	PopupTitle   string `json:"popup_title,omitempty"`
	PopupMessage string `json:"popup_message,omitempty"`
	///webhook
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    map[string]string `json:"body,omitempty"`
}

type AlertLevel struct {
	Enabled              bool
	Level                string
	Multiplier           float64 `json:"multiplier,omitempty"`
	Threshold            time.Duration
	TriggerHomeAssistant bool
	Actions              []ActionConfig
}

type IAModel struct {
	Model   string            `json:"model"`
	Type    string            `json:"type"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

type HyperfocusState struct {
	Level     string    /// determina o nivel do hiperfoco
	StartTime time.Time /// hora de inicio do hiperfoco
	EndTime   time.Time /// hora de fim do hiperfoco
}
type MiscConfig struct {
	WarnedThresholds       map[time.Duration]bool
	CurrentHyperfocusState *HyperfocusState
}

type Config struct {
	DEBUG                     bool
	PersonaName               string
	IAModel                   IAModel
	IdleTimeout               time.Duration
	ActivityCheckRate         time.Duration
	MinRandomQuestion         time.Duration
	MaxRandomQuestion         time.Duration
	DatabaseFile              string
	LogFile                   string
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
		DEBUG:                     false,
		IdleTimeout:               3 * time.Minute,  // Tempo para considerar o usuário ocioso
		ActivityCheckRate:         30 * time.Second, // Verificação menos frequente para economizar recursos
		ReduceOSSounds:            true,
		MinRandomQuestion:         45 * time.Minute,
		WellbeingQuestionsEnabled: true,
		MaxRandomQuestion:         90 * time.Minute,
		DatabaseFile:              "./focus_helper.db",
		LogFile:                   "./focus_helper.log",
		AlertLevels: []AlertLevel{
			{
				Enabled:   true,
				Level:     "LOW",
				Threshold: 45 * time.Minute,
				Actions: []ActionConfig{
					{Type: ActionSound, SoundFile: "alert_level_1.mp3"},
					{Type: ActionSpeak, Prompt: "Piloto-Alfa-Um, aqui é a Torre. Apenas um lembrete para verificar seus sistemas e fazer uma pequena pausa, se necessário."},
				},
			},
			{
				Enabled:    true,
				Level:      "MEDIUM",
				Threshold:  90 * time.Minute,
				Multiplier: 1.5,
				Actions: []ActionConfig{
					{Type: ActionSound, SoundFile: "autopilot.mp3"},
					{Type: ActionSpeak, Prompt: "Piloto-Alfa-Um, você está em um longo período de foco. Recomendamos uma pausa para hidratação e alongamento."},
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
					{Type: ActionSpeak, Prompt: "Piloto-Alfa-Um, detectamos sinais de hiperfoco. É crucial fazer uma pausa para manter a performance e o bem-estar."},
				},
			},
			{
				Enabled:    true,
				Level:      "CRITICAL",
				Threshold:  4 * time.Hour,
				Multiplier: 5.0,
				Actions: []ActionConfig{
					{Type: ActionSpeak, Prompt: "Mayday, Mayday, Mayday. Piloto-Alfa-Um, risco de burnout detectado. Desligue o piloto automático e faça uma pausa obrigatória imediatamente."}, // <-- Adicionado VoiceVolume
				},
			},
		},
	}
}

// func loadDebugConfig() Config {
// 	return Config{
// 		DEBUG:       true,
// 		PersonaName: "kitt",
// 		IAModel: IAModel{
// 			Type:  "ollama",
// 			Model: "llama3.2",
// 			URL:   "http://localhost:11434/api/generate"},
// 		IdleTimeout:               30 * time.Second,
// 		ActivityCheckRate:         5 * time.Second,
// 		MinRandomQuestion:         30 * time.Second,
// 		MaxRandomQuestion:         60 * time.Second,
// 		ReduceOSSounds:            true,
// 		WellbeingQuestionsEnabled: false,
// 		DatabaseFile:              "./focus_helper_debug.db",
// 		LogFile:                   "./focus_helper_debug.log",
// 		AlertLevels: []AlertLevel{
// 			{
// 				Enabled:    true,
// 				Level:      "LOW",
// 				Threshold:  10 * time.Second,
// 				Multiplier: 1.0,
// 				Actions: []ActionConfig{
// 					{Type: ActionSound, SoundFile: "alert_level_1.mp3"},
// 					{Type: ActionSpeak, Prompt: "Michael, detectei sinais iniciais de hiperfoco. Talvez seja prudente uma breve pausa."},
// 				},
// 			},
// 			{
// 				Enabled:    true,
// 				Level:      "MEDIUM",
// 				Threshold:  25 * time.Second,
// 				Multiplier: 1.5,
// 				Actions: []ActionConfig{
// 					{Type: ActionSound, SoundFile: "autopilot.mp3"},
// 					{Type: ActionSpeak, Prompt: "Michael, sua atenção está excessivamente concentrada. Sugiro deslocar-se até a cozinha e hidratar-se."},
// 				},
// 			},
// 			{
// 				Enabled:    true,
// 				Level:      "HIGH",
// 				Threshold:  45 * time.Second,
// 				Multiplier: 2.0,
// 				Actions: []ActionConfig{
// 					{Type: ActionSound, SoundFile: "alert_level_3.mp3"},
// 					{Type: ActionPopup, PopupTitle: "Alerta nível 3", PopupMessage: "Por favor, siga as instruções imediatamente."},
// 					{Type: ActionSpeak, Prompt: "Michael, sugiro que prossiga ao 'aeroporto' mais próximo — neste caso, a sala de descanso — e recupere o foco."},
// 				},
// 			},
// 			{
// 				Enabled:    true,
// 				Level:      "CRITICAL",
// 				Threshold:  60 * time.Second,
// 				Multiplier: 5.0,
// 				Actions: []ActionConfig{
// 					{Type: ActionSound, SoundFile: "autopilot.mp3"},
// 					{Type: ActionSpeak, Prompt: "Michael, ultrapassamos todos os limites seguros. Desative o piloto automático mental e retorne ao controle imediato."},
// 				},
// 			},
// 		},
// 	}
// }

func loadDebugConfig() Config {
	return Config{
		DEBUG:       true,
		PersonaName: "atc_tower",
		IAModel: IAModel{
			Type:  "ollama",
			Model: "llama3.2",
			URL:   "http://localhost:11434/api/generate"},
		IdleTimeout:               30 * time.Second,
		ActivityCheckRate:         5 * time.Second,
		MinRandomQuestion:         30 * time.Second,
		MaxRandomQuestion:         60 * time.Second,
		ReduceOSSounds:            true,
		WellbeingQuestionsEnabled: false,
		DatabaseFile:              "./focus_helper_debug.db",
		LogFile:                   "./focus_helper_debug.log",
		AlertLevels: []AlertLevel{
			{
				Enabled:   true,
				Level:     "LOW",
				Threshold: 45 * time.Minute,
				Actions: []ActionConfig{
					{Type: ActionSound, SoundFile: "alert_level_1.mp3"},
					{Type: ActionSpeak, Prompt: "Piloto-Alfa-Um, aqui é a Torre. Apenas um lembrete para verificar seus sistemas e fazer uma pequena pausa, se necessário."},
				},
			},
			{
				Enabled:    true,
				Level:      "MEDIUM",
				Threshold:  90 * time.Minute,
				Multiplier: 1.5,
				Actions: []ActionConfig{
					{Type: ActionSound, SoundFile: "autopilot.mp3"},
					{Type: ActionSpeak, Prompt: "Piloto-Alfa-Um, você está em um longo período de foco. Recomendamos uma pausa para hidratação e alongamento."},
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
					{Type: ActionSpeak, Prompt: "Piloto-Alfa-Um, detectamos sinais de hiperfoco. É crucial fazer uma pausa para manter a performance e o bem-estar."},
				},
			},
			{
				Enabled:    true,
				Level:      "CRITICAL",
				Threshold:  4 * time.Hour,
				Multiplier: 5.0,
				Actions: []ActionConfig{
					{Type: ActionSpeak, Prompt: "Mayday, Mayday, Mayday. Piloto-Alfa-Um, risco de burnout detectado. Desligue o piloto automático e faça uma pausa obrigatória imediatamente."}, // <-- Adicionado VoiceVolume
				},
			},
		},
	}
}
