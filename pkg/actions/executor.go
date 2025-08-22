// pkg/actions/executor.go
// REFACTORED: Sunday, August 10, 2025
package actions

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"focus-helper/pkg/audio"
	"focus-helper/pkg/common"
	"focus-helper/pkg/config"
	"focus-helper/pkg/language"
	"focus-helper/pkg/llm"
	"focus-helper/pkg/models"
	"focus-helper/pkg/notifications"
	"focus-helper/pkg/persona"
	"focus-helper/pkg/state"
	"focus-helper/pkg/variables"
)

type ExecutorDependencies struct {
	AppConfig    *models.Config
	Language     *language.LanguageManager
	VarProcessor *variables.Processor
	Notifier     notifications.Notifier
	LLMAdapter   llm.LLMAdapter
	AppState     *state.AppState
}

type Executor struct {
	deps     ExecutorDependencies
	cancelFn context.CancelFunc
	mu       sync.Mutex
}

func NewExecutor(deps ExecutorDependencies) *Executor {
	return &Executor{deps: deps}
}
func (e *Executor) Execute(action models.ActionConfig) error {
	e.mu.Lock()
	defer func() {
		fmt.Println("[Executor] Sending START_LISTENING event.")
		e.deps.AppState.EventChannel <- state.AppEvent{Type: "START_LISTENING"}
		e.mu.Unlock()
	}()
	e.deps.AppState.EventChannel <- state.AppEvent{Type: "STOP_LISTENING"}
	ctx, cancel := context.WithCancel(context.Background())
	e.cancelFn = cancel
	log.Printf("EXECUTING ACTION: Type=%s", action.Type)
	switch action.Type {
	case models.ActionSound:
		return e.executeSound(ctx, action)
	case models.ActionStop:
		return e.StopCurrentActions()
	case models.ActionSpeakIA:
		return e.executeSpeakIAAction(ctx, action)
	case models.ActionSpeak:
		return e.executeSpeakAction(ctx, action)
	case models.ActionYoutubeAudio:
		return e.executeYouTubeAudio(ctx, action)
	case models.ActionPopup:
		_, err := e.deps.Notifier.Question(action.PopupTitle, action.PopupMessage)
		return err
	default:
		return fmt.Errorf("wrong action: %s", action.Type)
	}
}

func (e *Executor) executeSound(ctx context.Context, action models.ActionConfig) error {
	done := make(chan error)
	go func() {
		done <- audio.PlaySound(audio.GetAssetPath(action.SoundFile), 1.0)
	}()
	select {
	case <-ctx.Done():
		log.Println("Sound action cancelled.")
		audio.StopCurrentSound()
		return fmt.Errorf("action cancelled")
	case err := <-done:
		return err
	}
}

func (e *Executor) executeSpeakIAAction(ctx context.Context, action models.ActionConfig) error {
	done := make(chan error)
	go func() {
		currentPersona, err := persona.GetPersona(e.deps.AppConfig.PersonaName, e.deps.VarProcessor)
		if err != nil {
			done <- fmt.Errorf("failed to get persona: %w", err)
			return
		}
		taskPrompt, _ := currentPersona.GetPrompt(e.deps.Language, action.Prompt)
		finalText, err := e.deps.LLMAdapter.Generate(taskPrompt)
		if err != nil {
			log.Printf("WARNING: LLM generation failed, falling back to basic prompt: %v", err)
		}
		err = currentPersona.ProcessAudio(finalText)
		if err != nil {
			log.Printf("error on processing audio: %v", err)
			done <- err
		}
		done <- nil
	}()
	select {
	case <-ctx.Done():
		log.Println("Sound action cancelled.")
		audio.StopCurrentSound()
		return fmt.Errorf("action cancelled")
	case err := <-done:
		return err
	}
}

func (e *Executor) executeSpeakAction(ctx context.Context, action models.ActionConfig) error {
	done := make(chan error)
	go func() {
		currentPersona, err := persona.GetPersona(e.deps.AppConfig.PersonaName, e.deps.VarProcessor)
		if err != nil {
			log.Printf("error on get person: %v", err)
			done <- err
		}
		finalText, err := currentPersona.GetText(e.deps.Language, action.Text)
		if err != nil {
			log.Printf("error on processing audio: %v", err)
			done <- err
		}
		err = currentPersona.ProcessAudio(finalText)
		if err != nil {
			log.Printf("error on processing audio: %v", err)
		}
		done <- nil

	}()
	select {
	case <-ctx.Done():
		log.Println("Sound action cancelled.")
		audio.StopCurrentSound()
		return fmt.Errorf("action cancelled")
	case err := <-done:
		return err
	}
}

func (e *Executor) executeYouTubeAudio(ctx context.Context, action models.ActionConfig) error {
	done := make(chan error)
	go func() {
		url := strings.TrimSpace(action.URL)
		if url == "" {
			done <- fmt.Errorf("failed to download YouTube audio: %v", url)
			return
		}
		videoID, err := common.GetYouTubeID(url)
		if err != nil {
			done <- err
			return
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
				done <- err
				return
			}
			log.Printf("Downloaded YouTube audio: %s", savePath)
		} else {
			log.Printf("Audio already exists, skipping download: %s", savePath)
		}
		err = audio.PlaySound(savePath, 1.0)
		if err != nil {
			log.Printf("error on processing audio: %v", err)
			done <- err
			return
		}
		done <- nil
	}()
	select {
	case <-ctx.Done():
		log.Println("Sound action cancelled.")
		audio.StopCurrentSound()
		return fmt.Errorf("action cancelled")
	case err := <-done:
		return err
	}
}

func (e *Executor) StopCurrentActions() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.cancelFn != nil {
		log.Println("Stopping all running actions...")
		e.cancelFn()
		e.cancelFn = nil
		audio.StopCurrentSound()
	}
	return nil
}
