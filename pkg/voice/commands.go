package voice

import (
	"log"
	"regexp"
	"strings"
)

var accentReplacements = map[rune]rune{
	'á': 'a', 'à': 'a', 'ã': 'a', 'â': 'a',
	'Á': 'a', 'À': 'a', 'Ã': 'a', 'Â': 'a',
	'é': 'e', 'è': 'e', 'ê': 'e', 'É': 'e', 'È': 'e', 'Ê': 'e',
	'í': 'i', 'ì': 'i', 'î': 'i', 'Í': 'i', 'Ì': 'i', 'Î': 'i',
	'ó': 'o', 'ò': 'o', 'õ': 'o', 'ô': 'o', 'Ó': 'o', 'Ò': 'o', 'Õ': 'o', 'Ô': 'o',
	'ú': 'u', 'ù': 'u', 'û': 'u', 'Ú': 'u', 'Ù': 'u', 'Û': 'u',
	'ç': 'c', 'Ç': 'c',
}

func (l *Listener) RegisterWakeUpWord(callback func(*CommandContext), phrases []string) {
	var normalized []string
	for _, p := range phrases {
		normalized = append(normalized, normalizeText(p))
	}
	l.commands = append(l.commands, Command{
		Phrases:      normalized,
		Callback:     callback,
		IsActivation: true,
	})
	log.Printf("Registered voice activation for phrases: %v", normalized)
}

func (l *Listener) RegisterCommand(callback func(*CommandContext), phrases []string) {
	var normalized []string
	for _, p := range phrases {
		normalized = append(normalized, normalizeText(p))
	}
	l.commands = append(l.commands, Command{
		Phrases:      normalized,
		Callback:     callback,
		IsActivation: false,
	})
	log.Printf("Registered voice command for phrases: %v", normalized)
}

func removeAccents(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if repl, ok := accentReplacements[r]; ok {
			sb.WriteRune(repl)
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func normalizeText(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = removeAccents(s)
	re := regexp.MustCompile(`[^a-z0-9\s]+`)
	s = re.ReplaceAllString(s, "")
	s = strings.Join(strings.Fields(s), " ")
	return s
}
