// pkg/actions/executor.go
// REFACTORED: Sunday, August 10, 2025
package actions

import (
	"fmt"
	"log"
	"sync"

	"focus-helper/pkg/audio"
	"focus-helper/pkg/config"
	"focus-helper/pkg/llm"
	"focus-helper/pkg/models"
	"focus-helper/pkg/notifications"
	"focus-helper/pkg/persona"
	"focus-helper/pkg/state"
	"focus-helper/pkg/variables"
)

// ExecutorDependencies holds all the managers and configs the executor needs.
type ExecutorDependencies struct {
	AppConfig    *models.Config
	VarProcessor *variables.Processor
	Notifier     notifications.Notifier
	LLMAdapter   llm.LLMAdapter
	AppState     *state.AppState
}

// Executor is responsible for executing all types of actions.
type Executor struct {
	deps ExecutorDependencies
}

var mutex sync.Mutex

// NewExecutor creates a new action executor.
func NewExecutor(deps ExecutorDependencies) *Executor {
	return &Executor{deps: deps}
}

// Execute takes an action config and performs the corresponding action.
func (e *Executor) Execute(action models.ActionConfig) error {
	log.Printf("EXECUTING ACTION: Type=%s", action.Type)
	mutex.Lock()
	defer mutex.Unlock()
	switch action.Type {
	case config.ActionSound:
		return audio.PlaySound(audio.GetAssetPath(action.SoundFile), 1.0)

	case config.ActionPopup:
		_, err := e.deps.Notifier.Question(action.PopupTitle, action.PopupMessage)
		return err

	case config.ActionSpeakIA:
		return e.executeSpeakIAAction(action)
	case config.ActionSpeak:
		return e.executeSpeakAction(action)

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
