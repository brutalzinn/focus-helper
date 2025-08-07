package config

import "time"

type ActionType string

const (
	ActionPopup ActionType = "POPUP"
	ActionSound ActionType = "SOUND"
	ActionATC   ActionType = "ATC_VOICE"
)

type ActionConfig struct {
	Type         ActionType `json:"type"`
	Volume       float64    `json:"volume,omitempty"`
	LlamaPrompt  string     `json:"llama_prompt,omitempty"`
	SoundFile    string     `json:"sound_file,omitempty"`
	PopupTitle   string     `json:"popup_title,omitempty"`
	PopupMessage string     `json:"popup_message,omitempty"`
}

type AlertLevel struct {
	Enabled              bool
	Threshold            time.Duration
	TriggerHomeAssistant bool
	Actions              []ActionConfig
	StaticVolume         float64
}

type Config struct {
	DEBUG                     bool
	IdleTimeout               time.Duration
	ActivityCheckRate         time.Duration
	MinRandomQuestion         time.Duration
	MaxRandomQuestion         time.Duration
	DatabaseFile              string
	LogFile                   string
	LlamaModel                string
	HomeAssistantEnabled      bool
	WellbeingQuestionsEnabled bool
	HomeAssistantWebhookURL   string
	AlertLevels               []AlertLevel
}

func LoadConfig(debugMode bool) Config {
	if debugMode {
		return loadDebugConfig()
	}
	return loadProdConfig()
}

func loadProdConfig() Config {
	return Config{
		DEBUG:                   false,
		IdleTimeout:             2 * time.Minute,
		ActivityCheckRate:       10 * time.Second,
		MinRandomQuestion:       45 * time.Minute,
		MaxRandomQuestion:       90 * time.Minute,
		DatabaseFile:            "./focus_helper.db",
		LogFile:                 "./focus_helper.log",
		LlamaModel:              "llama3.2:latest",
		HomeAssistantEnabled:    false,
		HomeAssistantWebhookURL: "http://localhost:8123/api/webhook/SEU_WEBHOOK_ID",
		AlertLevels: []AlertLevel{
			{
				Enabled:              true,
				Threshold:            10 * time.Second,
				TriggerHomeAssistant: false,
				Actions: []ActionConfig{
					{Type: ActionSound, SoundFile: "./assets/alert_level_1.mp3"},
					{Type: ActionATC, LlamaPrompt: "Você é uma torre de controle (ATC). O usuário é o 'Piloto-Alfa-Um' e o radar detectou provável Windshear. Gere uma mensagem de rádio curta e direta, usando fraseologia ATC, para instruir o piloto a abortar a decolagem."},
				},
			},
			{
				Enabled:              true,
				Threshold:            25 * time.Second,
				StaticVolume:         0.2,
				TriggerHomeAssistant: false,
				Actions: []ActionConfig{
					{Type: ActionSound, SoundFile: "./assets/autopilot.mp3"},
					{Type: ActionSound, SoundFile: "./assets/alert_level_2.mp3"},
					{Type: ActionATC, LlamaPrompt: "Você é uma torre de controle (ATC). O usuário é o 'Piloto-Alfa-Um' e está em rota próxima a um hiperfoco. Gere uma mensagem de rádio curta e direta, usando fraseologia ATC, para instruir o piloto a um provável pouso de emergência."},
				},
			},
			{
				Enabled:              true,
				Threshold:            45 * time.Second,
				StaticVolume:         0.4,
				TriggerHomeAssistant: false,
				Actions: []ActionConfig{
					{Type: ActionSound, SoundFile: "./assets/alert_level_3.mp3"},
					{Type: ActionPopup, PopupTitle: "Teste de alerta nível 3", PopupMessage: "Siga as instruções vetoriais da torre de comando imediatamente!"},
					{Type: ActionATC, LlamaPrompt: "Você é uma torre de controle (ATC). O usuário é o 'Piloto-Alfa-Um' e está perdendo o controle. Siga as instruções vetoriais na tela."},
				},
			},
		},
	}
}

func loadDebugConfig() Config {
	return Config{
		DEBUG:                   true,
		IdleTimeout:             30 * time.Second,
		ActivityCheckRate:       5 * time.Second,
		MinRandomQuestion:       30 * time.Second,
		MaxRandomQuestion:       60 * time.Second,
		DatabaseFile:            "./focus_helper_debug.db",
		LogFile:                 "./focus_helper_debug.log",
		LlamaModel:              "llama3.2:latest",
		HomeAssistantEnabled:    false,
		HomeAssistantWebhookURL: "http://SEU_HOME_ASSISTANT_IP:8123/api/webhook/SEU_WEBHOOK_ID_DEBUG",
		AlertLevels: []AlertLevel{
			{
				Enabled:              true,
				Threshold:            10 * time.Second,
				TriggerHomeAssistant: false,
				Actions: []ActionConfig{
					{Type: ActionSound, SoundFile: "./assets/alert_level_1.mp3"},
					{Type: ActionATC, LlamaPrompt: "Você é uma torre de controle (ATC). O usuário é o 'Piloto-Alfa-Um' e o radar detectou aviso de Windshear. Gere uma mensagem de rádio curta e direta, usando fraseologia ATC, para instruir o piloto a abortar a decolagem."},
				},
			},
			{
				Enabled:              true,
				Threshold:            25 * time.Second,
				TriggerHomeAssistant: false,
				Actions: []ActionConfig{
					{Type: ActionSound, SoundFile: "./assets/alert_level_2.mp3"},
					{Type: ActionATC, LlamaPrompt: "Você é uma torre de controle (ATC). O usuário é o 'Piloto-Alfa-Um' e está em rota próxima a um hiperfoco. Gere uma mensagem de rádio curta e direta, usando fraseologia ATC, para instruir o piloto a um provável pouso de emergência."},
				},
			},
			{
				Enabled:              true,
				Threshold:            45 * time.Second,
				TriggerHomeAssistant: false,
				Actions: []ActionConfig{
					{Type: ActionSound, SoundFile: "./assets/alert_level_3.mp3"},
					{Type: ActionPopup, PopupTitle: "Teste de alerta nível 3", PopupMessage: "Siga as instruções vetoriais da torre de comando imediatamente!"},
					{Type: ActionATC, LlamaPrompt: "Você é uma torre de controle (ATC). O usuário é o 'Piloto-Alfa-Um' e está perdendo o controle. Siga as instruções vetoriais na tela."},
				},
			},
		},
	}
}
