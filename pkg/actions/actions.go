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
	"focus-helper/pkg/models"
	"focus-helper/pkg/persona"
	"focus-helper/pkg/state"
)

var (
	mu               sync.Mutex
	currentAppState  *state.AppState
	actionCancelFn   context.CancelFunc
	sequenceCancelFn context.CancelFunc
)

func Init(appState *state.AppState) {
	currentAppState = appState
}

func Execute(action models.ActionConfig) error {
	mu.Lock()
	currentAppState.EventChannel <- state.AppEvent{Type: "STOP_LISTENING"}
	ctx, cancel := context.WithCancel(context.Background())
	actionCancelFn = cancel
	defer func() {
		fmt.Println("[Executor] Sending START_LISTENING event.")
		currentAppState.EventChannel <- state.AppEvent{Type: "START_LISTENING"}
		mu.Unlock()
	}()
	log.Printf("EXECUTING ACTION: Type=%s", action.Type)
	switch action.Type {
	case models.ActionSound:
		return executeSound(ctx, action)
	case models.ActionStop:
		return StopCurrentActions()
	case models.ActionSpeakIA:
		return executeSpeakIAAction(ctx, action)
	case models.ActionSpeak:
		return executeSpeakAction(ctx, action)
	case models.ActionYoutubeAudio:
		return executeYouTubeAudio(ctx, action)
	case models.ActionPopup:
		_, err := currentAppState.Notifier.Question(action.PopupTitle, action.PopupMessage)
		return err
	default:
		return fmt.Errorf("wrong action: %s", action.Type)
	}
}

func executeSound(ctx context.Context, action models.ActionConfig) error {
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

func executeSpeakIAAction(ctx context.Context, action models.ActionConfig) error {
	done := make(chan error)
	go func() {
		currentPersona, err := persona.GetPersona(currentAppState.AppConfig.PersonaName, currentAppState.VarProcessor)
		if err != nil {
			done <- fmt.Errorf("failed to get persona: %w", err)
			return
		}
		taskPrompt, _ := currentPersona.GetPrompt(currentAppState.Language, action.Prompt)
		finalText, err := currentAppState.LLMAdapter.Generate(taskPrompt)
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

func executeSpeakAction(ctx context.Context, action models.ActionConfig) error {
	done := make(chan error)
	go func() {
		currentPersona, err := persona.GetPersona(currentAppState.AppConfig.PersonaName, currentAppState.VarProcessor)
		if err != nil {
			log.Printf("error on get person: %v", err)
			done <- err
		}
		finalText, err := currentPersona.GetText(currentAppState.Language, action.Text)
		if err != nil {
			log.Printf("error on processing audio: %v", err)
			done <- err
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

func executeYouTubeAudio(ctx context.Context, action models.ActionConfig) error {
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

func StopCurrentActions() error {
	mu.Lock()
	defer mu.Unlock()
	log.Println("Stopping all running actions...")
	if actionCancelFn != nil {
		actionCancelFn()
		actionCancelFn = nil
	}
	if sequenceCancelFn != nil {
		sequenceCancelFn()
		sequenceCancelFn = nil
	}
	audio.StopCurrentSound()
	return nil
}

func ExecuteSequence(actions []models.ActionConfig) {
	ctx, cancel := context.WithCancel(context.Background())
	mu.Lock()
	sequenceCancelFn = cancel
	mu.Unlock()
	defer func() {
		mu.Lock()
		sequenceCancelFn = nil
		mu.Unlock()
		log.Println("--- Action Sequence Finished ---")
	}()
	log.Println("--- Starting Action Sequence ---")
	for i, action := range actions {
		select {
		case <-ctx.Done():
			log.Println("Action sequence cancelled.")
			return
		default:
		}
		log.Printf("Executing action %d/%d of sequence...", i+1, len(actions))
		if err := Execute(action); err != nil {
			log.Printf("Stopping sequence due to action error: %v", err)
			return
		}
	}
}
