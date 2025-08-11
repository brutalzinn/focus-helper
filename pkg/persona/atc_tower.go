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
	"focus-helper/pkg/config"
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
	templateWithContext := fmt.Sprintf("%s %s", a.GetSystemPrompt(lm), context)
	finalPrompt := a.proc.Process(templateWithContext, a.GetName())
	return finalPrompt, nil
}

func (a *ATCPersona) GetText(lm *language.LanguageManager, context string) (string, error) {
	finalPrompt := a.proc.Process(context, a.GetName())
	return finalPrompt, nil
}

func (a *ATCPersona) ProcessAudio(text string) error {
	timestamp := time.Now().UnixNano()
	backgroundAudioPath := audio.GetAssetPath("radio_static.wav")
	originalFilePath := filepath.Join(config.TEMP_AUDIO_DIR, fmt.Sprintf("%d_%s.wav", timestamp, a.GetName()))
	tempFiltered := filepath.Join(config.TEMP_AUDIO_DIR, fmt.Sprintf("%d_%s_temp_filter.wav", timestamp, a.GetName()))
	finalFilePath := filepath.Join(config.TEMP_AUDIO_DIR, fmt.Sprintf("%d_%s_final.wav", timestamp, a.GetName()))
	defer func() {
		_ = os.Remove(originalFilePath)
		_ = os.Remove(tempFiltered)
		_ = os.Remove(finalFilePath)
	}()
	piperCmd := exec.Command("piper", "--model", VOICE_MODEL, "--output_file", originalFilePath)
	piperCmd.Stdin = bytes.NewBufferString(text)
	if err := commands.RunCommand(piperCmd); err != nil {
		return fmt.Errorf("piper TTS failed: %w", err)
	}
	if err := audio.ApplyRadioFilter(originalFilePath, tempFiltered); err != nil {
		return err
	}
	audio.MixWithBackground(tempFiltered, backgroundAudioPath, finalFilePath, 0.4, 1.0)
	audio.PlaySound(finalFilePath, 1.0)
	return nil
}
