package actions

import (
	"fmt"

	"github.com/brutalzinn/focus-helper/config"
)

func NewActionFromConfig(level config.AlertLevel, actionCfg config.ActionConfig) (Action, error) {

	switch actionCfg.Type {
	case config.ActionSound:
		volumeMultiplier := level.Multiplier
		if volumeMultiplier <= 0 {
			volumeMultiplier = 1.0
		}
		return &SoundAction{
			FilePath:   actionCfg.SoundFile,
			Multiplier: volumeMultiplier,
		}, nil

	case config.ActionPopup:
		return &PopupAction{
			Title:   actionCfg.PopupTitle,
			Message: actionCfg.PopupMessage,
		}, nil

	case config.ActionATC:
		volumeMultiplier := level.Multiplier
		if volumeMultiplier <= 0 {
			volumeMultiplier = 1.0
		}
		return &ATCAction{
			LlamaPrompt:      actionCfg.LlamaPrompt,
			BackgroundFile:   actionCfg.BackgroundFile,
			BackgroundVolume: actionCfg.BackgroundVolume,
			VoiceVolume:      actionCfg.VoiceVolume,
			Multiplier:       volumeMultiplier,
		}, nil
	case config.ActionHomeAssistant:
		return &HomeAssistantAction{
			WebhookURL: actionCfg.HomeAssistant.WebhookURL,
			Data:       "",
		}, nil
	default:
		return nil, fmt.Errorf("tipo de ação desconhecido: %s", actionCfg.Type)
	}
}
