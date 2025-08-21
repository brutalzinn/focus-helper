package voice

import (
	"log"
	"strings"
)

func (l *Listener) RegisterCommand(phrase string, callback func(string)) {
	normalizedPhrase := strings.ToLower(phrase)
	l.commands[normalizedPhrase] = Command{
		Phrase:   phrase,
		Callback: callback,
	}
	log.Printf("Registered voice command for phrase: '%s'", phrase)
}
