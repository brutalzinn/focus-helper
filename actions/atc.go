package actions

import (
	"log"

	"github.com/brutalzinn/focus-helper/audio"
	"github.com/brutalzinn/focus-helper/config"
	"github.com/brutalzinn/focus-helper/integrations"
)

type ATCAction struct {
	LlamaPrompt      string
	BackgroundFile   string
	BackgroundVolume float64
	VoiceVolume      float64
	Multiplier       float64
}

func (a *ATCAction) Execute(alert config.AlertLevel) error {
	log.Println("  -> Executando ATCAction")
	// Gerar o texto com Llama

	prompt := integrations.NewATCPromptManager()
	finalPrompt := prompt.FormatPromptWithLevel(alert.Level, a.LlamaPrompt)
	alertText, err := integrations.GenerateTextWithLlama(config.AppConfig.Llama.Model, finalPrompt)
	if err != nil {
		log.Printf("Erro ao gerar texto ATC com Llama, usando fallback: %v", err)
		alertText = "Alfa-Um, aqui é a Torre. Ação imediata requerida."
	}

	return audio.PlayRadioSimulation(
		alertText,
		a.VoiceVolume,
		a.BackgroundVolume,
		a.BackgroundFile,
	)
}
