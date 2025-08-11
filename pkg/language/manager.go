// focus-helper/pkg/language/manager.go
// REFACTORED: Sunday, August 10, 2025
package language

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type LanguageManager struct {
	translations map[string]string
}

func NewManager(langDir, personaName, langCode string) (*LanguageManager, error) {
	m := &LanguageManager{
		translations: make(map[string]string),
	}

	filePath := filepath.Join(langDir, personaName, langCode+".json")
	log.Printf("LANGUAGE: Loading specific translation file: %s", filePath)

	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not read language file '%s': %w", filePath, err)
	}

	if err := json.Unmarshal(fileBytes, &m.translations); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from '%s': %w", filePath, err)
	}

	log.Printf("LANGUAGE: Successfully loaded '%s' for persona '%s'", langCode, personaName)
	return m, nil
}

func (m *LanguageManager) Get(key string) string {
	if value, ok := m.translations[key]; ok {
		return value
	}
	return fmt.Sprintf("!!MISSING_KEY:%s!!", key)
}
