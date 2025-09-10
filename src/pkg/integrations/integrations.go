package integrations

import (
	"bytes"
	"log"
	"net/http"
	"time"
)

func TriggerWebhook(webhookURL string, payload []byte) error {
	if webhookURL == "" {
		log.Println("URL do Home Assistant não configurada. Pulando webhook.")
		return nil
	}
	req, _ := http.NewRequest("POST", webhookURL, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Erro ao enviar webhook para Home Assistant: %v", err)
		return err
	}
	defer resp.Body.Close()
	log.Printf("Webhook enviado para Home Assistant. Status: %s", resp.Status)
	return nil
}
