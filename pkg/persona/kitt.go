package persona

import (
	"bytes"
	"fmt"
	"focus-helper/pkg/audio"
	"focus-helper/pkg/commands"
	"focus-helper/pkg/config"
	"focus-helper/pkg/language"
	"focus-helper/pkg/variables"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type KittPersona struct {
	proc *variables.Processor
}

func NewKittPersona(vp *variables.Processor) *KittPersona {
	return &KittPersona{proc: vp}
}

func (k *KittPersona) GetName() string {
	return "kitt"
}

func (k *KittPersona) GetSystemPrompt(lm *language.LanguageManager) string {
	return lm.Get("system_prompt")
}

func (k *KittPersona) GetConfirmWord(lm *language.LanguageManager) string {
	return lm.Get("confirm_word")
}

func (k *KittPersona) GetPrompt(lm *language.LanguageManager, context string) (string, error) {
	templateWithContext := fmt.Sprintf("%s %s", k.GetSystemPrompt(lm), context)
	finalPrompt := k.proc.Process(templateWithContext, k.GetName())
	return finalPrompt, nil
}

func (k *KittPersona) GetText(lm *language.LanguageManager, context string) (string, error) {
	finalPrompt := k.proc.Process(context, k.GetName())
	return finalPrompt, nil
}

func (k *KittPersona) ProcessAudio(text string) error {
	timestamp := time.Now().UnixNano()
	originalFilePath := filepath.Join(config.GetUserConfigPath(), config.TEMP_AUDIO_DIR, fmt.Sprintf("%d_%s.wav", timestamp, k.GetName()))
	finalFilePath := filepath.Join(config.GetUserConfigPath(), config.TEMP_AUDIO_DIR, fmt.Sprintf("%d_%s_temp_filter.wav", timestamp, k.GetName()))
	defer func() {
		_ = os.Remove(originalFilePath)
		_ = os.Remove(finalFilePath)
	}()
	piperCmd := exec.Command("piper", "--model", VOICE_MODEL, "--output_file", originalFilePath)
	piperCmd.Stdin = bytes.NewBufferString(text)
	if err := commands.RunCommand(piperCmd); err != nil {
		return fmt.Errorf("piper TTS failed: %w", err)
	}
	if err := audio.ApplyRadioFilter(originalFilePath, finalFilePath); err != nil {
		return err
	}
	audio.PlaySound(finalFilePath, 1.0)
	return nil
}

func (k *KittPersona) GetDisplayWarn(context string) (*DisplayContent, error) {
	return &DisplayContent{
		Type:    "html_dialog",
		Value:   "kitt/index.html",
		Options: map[string]any{"width": 400, "height": 180, "title": "K.I.T.T. Voice Module"},
	}, nil
}
