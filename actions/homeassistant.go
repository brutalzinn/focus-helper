package actions

import (
	"log"

	"github.com/brutalzinn/focus-helper/config"
	"github.com/brutalzinn/focus-helper/integrations"
)

type HomeAssistantAction struct {
	WebhookURL string
	Data       string
}

func (a *HomeAssistantAction) Execute(alert config.AlertLevel) error {
	log.Println("  -> Executando HomeAssistantAction")
	return integrations.TriggerHomeAssistant(a.WebhookURL, a.Data)
}
