package persona

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"focus-helper/pkg/audio"
	"focus-helper/pkg/commands"
	"focus-helper/pkg/language"
	"focus-helper/pkg/variables"
)

type ATCPersona struct {
	proc *variables.Processor
}

func NewATCPersona(vp *variables.Processor) *ATCPersona {
	return &ATCPersona{proc: vp}
}

func (a *ATCPersona) GetName() string {
	return "atc_tower"
}

func (a *ATCPersona) GetSystemPrompt(lm *language.LanguageManager) string {
	return lm.Get("system_prompt")
}

func (a *ATCPersona) GetConfirmWord(lm *language.LanguageManager) string {
	return lm.Get("confirm_word")
}

func (a *ATCPersona) GetPrompt(lm *language.LanguageManager, context string) (string, error) {
	rawTemplate := lm.Get("alert_prompt")
	templateWithContext := fmt.Sprintf(rawTemplate, context)
	finalPrompt := a.proc.Process(templateWithContext, a.GetName())
	return finalPrompt, nil
}

func (a *ATCPersona) ProcessAudio(prompt, filePath string) error {
	timestamp := time.Now().UnixNano()
	normalVoiceFileName := fmt.Sprintf("%s_%d.wav", a.GetName(), timestamp)
	finalFilePath := filepath.Join("temp_audio", normalVoiceFileName)
	defer os.Remove(finalFilePath)
	piperCmd := exec.Command("piper", "--model", VOICE_MODEL, "--output_file", finalFilePath)
	piperCmd.Stdin = bytes.NewBufferString(prompt)
	if err := commands.RunCommand(piperCmd); err != nil {
		return fmt.Errorf("piper TTS failed: %w", err)
	}
	// tempFiltered := fmt.Sprintf("%s_temp.wav", filePath)
	// if err := audio.ApplyRadioFilter(filePath, tempFiltered); err != nil {
	// 	return err
	// }
	// defer os.Remove(tempFiltered)
	// audio.MixWithBackground(tempFiltered, "assets/radio_static.wav", filePath, 0.4)

	audio.PlaySound(finalFilePath, 1.0)

	return nil
}
