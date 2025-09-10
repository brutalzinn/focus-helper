

package persona

import (
	"fmt"
	"path/filepath"

	"focus-helper/src/pkg/config"
	"focus-helper/src/pkg/language"
	"focus-helper/src/pkg/variables"
)

var (
	VOICE_MODEL = filepath.Join(config.GetUserConfigPath(), "voices", "pt_BR-cadu-medium.onnx")
)

type DisplayContent struct {
	Type    string         `json:"type"`
	Value   string         `json:"value"`
	Options map[string]any `json:"options,omitempty"`
}

type Persona interface {
	GetName() string

	GetSystemPrompt(lm *language.LanguageManager) string

	GetConfirmWord(lm *language.LanguageManager) string

	GetPrompt(lm *language.LanguageManager, context string) (string, error)

	GetText(lm *language.LanguageManager, context string) (string, error)

	ProcessAudio(prompt string) error
}

type VisualPersona interface {
	GetDisplayWarn(context string) (*DisplayContent, error)
}

func GetPersona(name string, vp *variables.Processor) (Persona, error) {
	switch name {
	case "atc_tower":
		return NewATCPersona(vp), nil
	case "kitt":
		return NewKittPersona(vp), nil
	default:
		return nil, fmt.Errorf("persona '%s' not found", name)
	}
}
