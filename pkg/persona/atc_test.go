package persona

import (
	"testing"

	"focus-helper/pkg/language"
	"focus-helper/pkg/variables"
)

func TestATCPersona_GetPrompt(t *testing.T) {
	varProcessor := variables.NewProcessor()
	atcPersona := NewATCPersona(varProcessor)
	lm, err := language.NewManager("../../language", "atc_tower", "pt-br")
	if err != nil {
		t.Fatalf("Failed to create language manager for test: %v", err)
	}
	prompt, err := atcPersona.GetPrompt(lm, "Pista 05 interditada")
	if err != nil {
		t.Fatalf("GetPrompt returned an unexpected error: %v", err)
	}
	expected := "Torre de controle para Alex, alerta de prioridade. Pista 05 interditada. Confirme o recebimento."
	if prompt != expected {
		t.Errorf("Expected prompt '%s', but got '%s'", expected, prompt)
	}
}
