package integrations

import (
	"bytes"
	"log"
	"net/http"
	"time"
)

func TriggerHomeAssistant(webhookURL string, payload string) {
	if webhookURL == "" || webhookURL == "http://SEU_HOME_ASSISTANT_IP:8123/api/webhook/SEU_WEBHOOK_ID" {
		log.Println("URL do Home Assistant não configurada. Pulando webhook.")
		return
	}
	req, _ := http.NewRequest("POST", webhookURL, bytes.NewBuffer([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Erro ao enviar webhook para Home Assistant: %v", err)
		return
	}
	defer resp.Body.Close()
	log.Printf("Webhook enviado para Home Assistant. Status: %s", resp.Status)
}
