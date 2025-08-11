// pkg/llm/ollama.go
package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"focus-helper/pkg/config"
)

type ollamaStreamResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

type OllamaAdapter struct {
	config config.IAModel
}

func NewOllamaAdapter(cfg config.IAModel) *OllamaAdapter {
	return &OllamaAdapter{config: cfg}
}

func (a *OllamaAdapter) Generate(prompt string) (string, error) {
	reqBody := map[string]any{
		"model":  a.config.Model,
		"prompt": prompt,
		"stream": false,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("ollama: erro ao serializar corpo da requisição: %w", err)
	}
	req, err := http.NewRequest("POST", "http://localhost:11434/api/generate", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("ollama: erro ao criar requisição: %w", err)
	}
	for key, value := range a.config.Headers {
		req.Header.Set(key, value)
	}
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama: failed to send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama: API returned non-200 status: %s", resp.Status)
	}
	var result struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("ollama: erro ao decodificar resposta: %w", err)
	}
	return result.Response, nil
}
