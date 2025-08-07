package actions

import (
	"log"

	"github.com/brutalzinn/focus-helper/config"
	"github.com/brutalzinn/focus-helper/notifications"
)

type PopupAction struct {
	Title   string
	Message string
}

func (a *PopupAction) Execute(alert config.AlertLevel) error {
	log.Printf("  -> Executando PopupAction: %s", a.Title)
	notifications.ShowPopup(a.Title, a.Message)
	return nil
}
