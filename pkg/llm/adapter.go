// pkg/llm/adapter.go
package llm

import (
	"fmt"
	"focus-helper/pkg/config"
)

// LLMAdapter is the standard interface your application will use
// to interact with any language model.
type LLMAdapter interface {
	Generate(prompt string) (string, error)
}

// NewAdapter is a factory function that creates the correct adapter
// based on the configuration.
func NewAdapter(cfg config.IAModel) (LLMAdapter, error) {
	switch cfg.Type {
	case "ollama":
		return NewOllamaAdapter(cfg), nil
	// case "openai":
	// 	return NewOpenAIAdapter(cfg), nil // For future expansion
	default:
		return nil, fmt.Errorf("unsupported LLM adapter type: %s", cfg.Type)
	}
}
