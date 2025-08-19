package models

import (
	"encoding/json"
	"fmt"
	"time"
)

type ActionType string

type Duration struct {
	time.Duration
}

type ActionConfig struct {
	Type         ActionType `json:"type"`
	RandomChance float64    `json:"random_chance,omitempty"`
	SoundFile    string     `json:"sound_file,omitempty"`
	Volume       float64    `json:"volume,omitempty"`
	Text         string     `json:"text,omitempty"`

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
	Enabled   bool           `json:"enabled"`
	Level     string         `json:"level"`
	Threshold Duration       `json:"threshold"`
	Actions   []ActionConfig `json:"actions"`
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
	WarnedThresholds map[time.Duration]bool
}

type Config struct {
	ProfileName               string            `json:"name"`
	HyperfocusAssociations    map[string]string `json:"hyperfocus_associations"`
	DEBUG                     bool              `json:"debug"`
	TimeLocation              string            `json:"time_location"`
	Language                  string            `json:"language"`
	HyperfocusMinDuration     Duration          `json:"hyperfocus_min_duration"`
	Username                  string            `json:"username"`
	PersonaName               string            `json:"persona_name"`
	IAModel                   IAModel           `json:"iamodel"`
	IdleTimeout               Duration          `json:"idle_timeout"`
	ActivityCheckRate         Duration          `json:"activity_check_rate"`
	MinRandomQuestion         Duration          `json:"min_random_question"`
	MaxRandomQuestion         Duration          `json:"max_random_question"`
	ReduceOSSounds            bool              `json:"reduce_os_sounds"`
	IADetectorEnabled         bool              `json:"ia_detector_enabled"`
	WellbeingQuestionsEnabled bool              `json:"wellbeing_questions_enabled"`
	DatabaseFile              string            `json:"database_file"`
	LogFile                   string            `json:"log_file"`
	Misc                      MiscConfig        `json:"misc"`
	AlertLevels               []AlertLevel      `json:"alert_levels"`
	MaydayListenerEnabled     bool              `json:"mayday_listener_enabled"`
	MaydayActivationWord      string            `json:"mayday_activation_word"`
	VADThreshold              float64           `json:"vad_threshold"`
	VADSilenceTimeout         Duration          `json:"vad_silence_timeout"`
	WhisperModelPath          string            `json:"whisper_model_path"`
}

func (d *Duration) UnmarshalJSON(b []byte) (err error) {
	if b[0] == '"' {
		sd := string(b[1 : len(b)-1])
		d.Duration, err = time.ParseDuration(sd)
		return
	}
	var id int64
	id, err = json.Number(string(b)).Int64()
	d.Duration = time.Duration(id)
	return
}

func (d Duration) MarshalJSON() (b []byte, err error) {
	return []byte(fmt.Sprintf(`"%s"`, d.String())), nil
}
