// focus-helper/pkg/language/manager.go
// REFACTORED: Sunday, August 10, 2025
package language

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type LanguageManager struct {
	translations map[string]any
}

func NewManager(langDir, personaName, langCode string) (*LanguageManager, error) {
	m := &LanguageManager{
		translations: make(map[string]any),
	}
	if err := m.loadTranslations(filepath.Join(langDir, "general", langCode+".json")); err != nil {
		log.Printf("LANGUAGE: Warning — could not load general translations: %v", err)
	}
	if err := m.loadTranslations(filepath.Join(langDir, personaName, langCode+".json")); err != nil {
		return nil, fmt.Errorf("could not load persona translations: %w", err)
	}
	log.Printf("LANGUAGE: Successfully loaded '%s' for persona '%s'", langCode, personaName)
	return m, nil
}

func (m *LanguageManager) loadTranslations(filePath string) error {
	log.Printf("LANGUAGE: Loading translation file: %s", filePath)
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	tmp := make(map[string]interface{})
	if err := json.Unmarshal(fileBytes, &tmp); err != nil {
		return fmt.Errorf("failed to parse JSON from '%s': %w", filePath, err)
	}
	for k, v := range tmp {
		m.translations[k] = v
	}
	return nil
}

func (m *LanguageManager) Get(key string) string {
	parts := strings.Split(key, ".")
	var current any = m.translations
	for _, p := range parts {
		if mMap, ok := current.(map[string]interface{}); ok {
			if val, exists := mMap[p]; exists {
				current = val
			} else {
				return fmt.Sprintf("!!MISSING_KEY:%s!!", key)
			}
		} else {
			return fmt.Sprintf("!!MISSING_KEY:%s!!", key)
		}
	}

	if str, ok := current.(string); ok {
		return str
	}

	return fmt.Sprintf("!!INVALID_TYPE:%s!!", key)
}
