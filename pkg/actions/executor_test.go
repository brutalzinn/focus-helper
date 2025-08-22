// pkg/actions/executor_test.go
package actions

import (
	"testing"

	"focus-helper/pkg/models"
)

func TestExecutor_ExecuteSound(t *testing.T) {
	execDeps := ExecutorDependencies{
		AppConfig:  &models.Config{},
		Notifier:   nil,
		LLMAdapter: nil,
	}

	exec := NewExecutor(execDeps)

	action := models.ActionConfig{
		Type:      models.ActionSound,
		SoundFile: "kitt_scanner.mp3",
	}

	if err := exec.Execute(action); err != nil {
		t.Fatalf("Execute Sound failed: %v", err)
	}
}

func TestExecutor_ExecuteYouTubeAudio(t *testing.T) {
	execDeps := ExecutorDependencies{
		AppConfig:  &models.Config{},
		Notifier:   nil,
		LLMAdapter: nil,
	}

	exec := NewExecutor(execDeps)

	action := models.ActionConfig{
		Type:    models.ActionYoutubeAudio,
		URL:     "https://www.youtube.com/watch?v=-VVML0k5xsk",
		StartAt: "0:00",
		EndAt:   "0:03",
	}

	if err := exec.Execute(action); err != nil {
		t.Fatalf("Execute YouTubeAudio failed: %v", err)
	}

}
