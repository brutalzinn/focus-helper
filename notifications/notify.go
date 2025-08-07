package notifications

import (
	"github.com/gen2brain/beeep"
	"github.com/sqweek/dialog"
)

// ShowPopup exibe um diálogo modal no centro da tela.
func ShowPopup(title, message string) {
	dialog.Message("%s", message).Title(title).Info()
}

// ShowQuestionPopup exibe um diálogo de pergunta Sim/Não.
func ShowQuestionPopup(title, question string) bool {
	return dialog.Message("%s", question).Title(title).YesNo()
}

// ShowDesktopNotification envia uma notificação padrão de sistema.
func ShowDesktopNotification(title, message string) {
	beeep.Alert(title, message, "") // O último argumento é o ícone, opcional.
}
