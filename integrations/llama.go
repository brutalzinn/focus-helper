package integrations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

func GenerateTextWithLlama(model, prompt string) (string, error) {
	endpoint := "http://localhost:11434/api/generate"
	requestData := OllamaRequest{Model: model, Prompt: prompt, Stream: false}
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("erro ao converter para JSON: %w", err)
	}
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("erro ao criar HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	log.Println("Enviando prompt para Ollama...")
	client := &http.Client{Timeout: 30 * time.Second}
	httpResponse, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro ao enviar request para Ollama: %w", err)
	}
	defer httpResponse.Body.Close()
	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return "", fmt.Errorf("erro ao ler resposta do Ollama: %w", err)
	}
	if httpResponse.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama retornou status não-200: %s - Body: %s", httpResponse.Status, string(responseBody))
	}
	var ollamaResp OllamaResponse
	err = json.Unmarshal(responseBody, &ollamaResp)
	if err != nil {
		return "", fmt.Errorf("erro ao decodificar JSON do Ollama: %w", err)
	}
	log.Println("Resposta recebida do Ollama.")
	return ollamaResp.Response, nil
}
