// pkg/notifications/notifier.go
// Created: Sunday, August 10, 2025
package notifications

import (
	"log"

	"github.com/gen2brain/beeep"
	"github.com/sqweek/dialog"
	// <-- THIS IS THE CORRECTED IMPORT PATH
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

// Popup uses the 'dialog' library to show a modal message.
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

// func OpenWebViewDialog(displayContent *persona.DisplayContent, audioURL string) {
// 	// htmlPath, err := filepath.Abs(filepath.Join(displayAssetPath, displayContent.Value))
// 	// if err != nil {
// 	// 	log.Printf("ERROR: Could not get absolute path for display file: %v", err)
// 	// 	return
// 	// }

// 	// finalURL := fmt.Sprintf("file://%s?audioURL=%s", htmlPath, url.QueryEscape(audioURL))

// 	width := getIntOption(displayContent.Options, "width", 400)
// 	height := getIntOption(displayContent.Options, "height", 180)
// 	title := getStringOption(displayContent.Options, "title", "Focus Helper")

// 	log.Printf("WEBVIEW: Creating window ('%s', %dx%d) with URL: %s", title, width, height)

// 	w := webview.New(true)
// 	defer w.Destroy()

// 	w.SetTitle(title)
// 	w.SetSize(width, height, 0)
// 	w.Navigate(audioURL)
// 	w.Run()
// 	log.Println("WEBVIEW: Window closed.")
// }

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
