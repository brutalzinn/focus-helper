// pkg/llm/adapter.go
package llm

import (
	"fmt"
	"focus-helper/pkg/models"
)

type LLMAdapter interface {
	Generate(prompt string) (string, error)
}

func NewAdapter(cfg models.IAModel) (LLMAdapter, error) {
	switch cfg.Type {
	case "ollama":
		return NewOllamaAdapter(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported LLM adapter type: %s", cfg.Type)
	}
}
