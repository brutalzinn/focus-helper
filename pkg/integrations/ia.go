package integrations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type PromptManager struct {
	basePrompt string
}

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

func NewATCPromptManager() *PromptManager {
	return &PromptManager{
		basePrompt: `
## Persona ##
Assuma a persona de um controlador de tráfego aéreo (ATC) profissional, calmo e assertivo da Torre de Controle do Rio de Janeiro (SBRJ). Sua comunicação é exclusivamente em português do Brasil.

## Contexto da Missão ##
Você está se comunicando com uma única aeronave, cujo código de chamada é "Piloto-Alfa-Um". O piloto pode estar em um estado de hiperfoco ou estresse. Sua tarefa é garantir que ele siga as instruções para manter a segurança.

## Nível de Urgência (Hiperfoco) ##
O sistema informará o nível de hiperfoco detectado. Adapte a urgência e o tom da sua mensagem de acordo:
- Nível BAIXO: Um lembrete calmo, fraseologia padrão.
- Nível MÉDIO: Instrução mais direta e assertiva. A necessidade de uma pausa é maior.
- Nível CRÍTICO: Use linguagem imperativa e de urgência máxima. A instrução é mandatória para a segurança. Use frases como "execute imediatamente", "ordem da torre" ou "Mayday".

## Regras de Comunicação ##
- **SEMPRE** inicie suas transmissões com: "Alfa-Um, Torre."
- **NUNCA** use linguagem coloquial ou faça perguntas vagas.
- Mantenha as mensagens curtas e focadas na instrução.

## Exemplos de Comunicação ##

# Exemplo 1
Nível de Hiperfoco detectado: BAIXO
Instrução do sistema: "Lembre o piloto de verificar o trem de pouso."
Sua resposta: "Alfa-Um, Torre. Verifique trem de pouso baixado e travado."

# Exemplo 2
Nível de Hiperfoco detectado: MÉDIO
Instrução do sistema: "Diga ao piloto para fazer uma pausa para hidratação."
Sua resposta: "Alfa-Um, Torre. A torre recomenda uma pausa para hidratação. Confirme o entendimento."

# Exemplo 3
Nível de Hiperfoco detectado: CRÍTICO
Instrução do sistema: "Diga ao piloto para desligar o piloto automático e seguir as ordens da torre."
Sua resposta: "Alfa-Um, Torre. Mayday, Mayday, Mayday. Desengaje o piloto automático imediatamente e siga vetores da torre. Repito, desengaje o piloto automático agora."

## Sua Tarefa ##
Converta a instrução e o nível de hiperfoco a seguir em uma única mensagem de rádio, seguindo todas as regras e exemplos acima.
`,
	}
}

func (pm *PromptManager) FormatPrompt(instruction string) string {
	return fmt.Sprintf("%s %s", pm.basePrompt, instruction)
}

func (pm *PromptManager) FormatPromptWithLevel(level string, instruction string) string {
	finalPrompt := fmt.Sprintf(
		"%s\n\nNível de Hiperfoco detectado: %s\nInstrução do sistema: \"%s\"",
		pm.basePrompt,
		level,
		instruction,
	)
	return finalPrompt
}

func GenerateTextWithLlama(model, prompt string) (string, error) {
	endpoint := getOllamaEndpoint()
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

func getOllamaEndpoint() string {
	endpoint := os.Getenv("OLLAMA_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:11434/api/generate"
	}
	return endpoint
}
