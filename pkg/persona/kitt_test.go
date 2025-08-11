package persona

import (
	"focus-helper/pkg/variables"
	"reflect"
	"testing"
)

func setupKittTest(t *testing.T) *KittPersona {
	varProcessor := variables.NewProcessor()
	return NewKittPersona(varProcessor)
}

func TestKittPersona_GetName(t *testing.T) {
	kitt := setupKittTest(t)
	if name := kitt.GetName(); name != "kitt" {
		t.Errorf("Expected name 'kitt', got '%s'", name)
	}
}
func TestKittPersona_GetDisplayWarn(t *testing.T) {
	kitt := setupKittTest(t)
	displayContent, err := kitt.GetDisplayWarn("")
	if err != nil {
		t.Fatalf("GetDisplayWarn returned an unexpected error: %v", err)
	}

	expected := &DisplayContent{
		Type:    "html_dialog",
		Value:   "kitt/index.html",
		Options: map[string]any{"width": 400, "height": 180, "title": "K.I.T.T. Voice Module"},
	}

	if !reflect.DeepEqual(displayContent, expected) {
		t.Errorf("DisplayContent did not match expected.\nGot: %+v\nExpected: %+v", displayContent, expected)
	}
}
