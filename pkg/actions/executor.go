// pkg/actions/executor.go
// REFACTORED: Sunday, August 10, 2025
package actions

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"focus-helper/pkg/audio"
	"focus-helper/pkg/common"
	"focus-helper/pkg/config"
	"focus-helper/pkg/llm"
	"focus-helper/pkg/models"
	"focus-helper/pkg/notifications"
	"focus-helper/pkg/persona"
	"focus-helper/pkg/state"
	"focus-helper/pkg/variables"
)

type ExecutorDependencies struct {
	AppConfig    *models.Config
	VarProcessor *variables.Processor
	Notifier     notifications.Notifier
	LLMAdapter   llm.LLMAdapter
	AppState     *state.AppState
}

type Executor struct {
	deps ExecutorDependencies
}

func NewExecutor(deps ExecutorDependencies) *Executor {
	return &Executor{deps: deps}
}

// Execute takes an action config and performs the corresponding action.
func (e *Executor) Execute(action models.ActionConfig) error {

	log.Printf("EXECUTING ACTION: Type=%s", action.Type)

	switch action.Type {
	case models.ActionSound:
		return audio.PlaySound(audio.GetAssetPath(action.SoundFile), 1.0)

	case models.ActionPopup:
		_, err := e.deps.Notifier.Question(action.PopupTitle, action.PopupMessage)
		return err

	case models.ActionSpeakIA:
		return e.executeSpeakIAAction(action)
	case models.ActionSpeak:
		return e.executeSpeakAction(action)
	case models.ActionYoutubeAudio:
		return e.executeYouTubeAudio(action)

	default:
		return fmt.Errorf("ação desconhecida: %s", action.Type)
	}
}

func (e *Executor) executeSpeakIAAction(action models.ActionConfig) error {

	currentPersona, err := persona.GetPersona(e.deps.AppConfig.PersonaName, e.deps.VarProcessor)
	if err != nil {
		return fmt.Errorf("failed to get persona: %w", err)
	}
	taskPrompt, _ := currentPersona.GetPrompt(e.deps.AppState.Language, action.Prompt)
	finalText, err := e.deps.LLMAdapter.Generate(taskPrompt)
	if err != nil {
		log.Printf("WARNING: LLM generation failed, falling back to basic prompt: %v", err)
	}
	err = currentPersona.ProcessAudio(finalText)
	if err != nil {
		log.Printf("error on processing audio: %v", err)
	}
	// if visualPersona, ok := currentPersona.(persona.VisualPersona); ok {
	// 	displayContent, _ := visualPersona.GetDisplayWarn(finalText)
	// 	if displayContent != nil && displayContent.Type == "html_dialog" {
	// 		log.Printf("VISUAL TRIGGER: Opening dialog '%s' with URL '%s'", displayContent.Value)
	// 		notifications.OpenWebViewDialog(displayContent)
	// 		return nil
	// 	}
	// }
	return nil
}

func (e *Executor) executeSpeakAction(action models.ActionConfig) error {
	currentPersona, err := persona.GetPersona(e.deps.AppConfig.PersonaName, e.deps.VarProcessor)
	if err != nil {
		return fmt.Errorf("failed to get persona: %w", err)
	}
	finalText, err := currentPersona.GetText(e.deps.AppState.Language, action.Text)
	if err != nil {
		return fmt.Errorf("failed to get text persona: %w", err)
	}
	err = currentPersona.ProcessAudio(finalText)
	if err != nil {
		log.Printf("error on processing audio: %v", err)
	}
	return nil
}

func (e *Executor) executeYouTubeAudio(action models.ActionConfig) error {
	url := strings.TrimSpace(action.URL)
	if url == "" {
		return fmt.Errorf("no YouTube URL provided")
	}
	videoID, err := common.GetYouTubeID(url)
	if err != nil {
		return err
	}
	savePath := filepath.Join(config.GetUserConfigPath(), config.ASSETS_DIR, videoID+".mp3")
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		args := []string{"-x", "--audio-format", "mp3", "-o", savePath, url}
		if action.StartAt != "" || action.EndAt != "" {
			start := action.StartAt
			end := action.EndAt
			args = append(args, "--postprocessor-args", fmt.Sprintf("-ss %s -to %s", start, end))
		}
		cmd := exec.Command("yt-dlp", args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("yt-dlp error: %s", string(out))
			return fmt.Errorf("failed to download YouTube audio: %w", err)
		}
		log.Printf("Downloaded YouTube audio: %s", savePath)
	} else {
		log.Printf("Audio already exists, skipping download: %s", savePath)
	}
	err = audio.PlaySound(savePath, 1.0)
	if err != nil {
		return fmt.Errorf("failed to play audio: %w", err)
	}
	return nil
}
