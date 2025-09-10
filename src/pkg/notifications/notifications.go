

package notifications

import (
	"log"

	"github.com/gen2brain/beeep"
	"github.com/sqweek/dialog"

)

const (
	displayAssetPath = "assets/displays"
)

type Notifier interface {
	Popup(title, message string) error

	Question(title, message string) (bool, error)

	Notify(title, message string) error
}

type DesktopNotifier struct{}

func NewDesktopNotifier() *DesktopNotifier {
	return &DesktopNotifier{}
}


func (n *DesktopNotifier) Popup(title, message string) error {
	log.Printf("NOTIFICATION (Popup): Title='%s', Message='%.30s...'", title, message)
	dialog.Message("%s", message).Title(title).Info()
	return nil
}

func (n *DesktopNotifier) Question(title, message string) (bool, error) {
	log.Printf("NOTIFICATION (Question): Title='%s', Message='%.30s...'", title, message)
	return dialog.Message(message).Title(title).YesNo(), nil
}

func (n *DesktopNotifier) Notify(title, message string) error {
	log.Printf("NOTIFICATION (Notify): Title='%s', Message='%.30s...'", title, message)
	return beeep.Alert(title, message, "")
}


























func getIntOption(options map[string]any, key string, defaultValue int) int {
	if val, ok := options[key].(int); ok {
		return val
	}
	if val, ok := options[key].(float64); ok {
		return int(val)
	}
	return defaultValue
}

func getStringOption(options map[string]any, key string, defaultValue string) string {
	if val, ok := options[key].(string); ok {
		return val
	}
	return defaultValue
}
