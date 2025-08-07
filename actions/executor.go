package actions

import (
	"fmt"
	"log"

	"github.com/brutalzinn/focus-helper/audio"
	"github.com/brutalzinn/focus-helper/config"
	"github.com/brutalzinn/focus-helper/integrations"
	"github.com/brutalzinn/focus-helper/notifications"
)

func Execute(level config.AlertLevel, cfg config.Config) {
	log.Printf("Executando ações para o nível com limite de %v", level.Threshold)
	for _, action := range level.Actions {
		actionVolume := action.Volume
		if actionVolume == 0 {
			actionVolume = 1.0
		}

		switch action.Type {
		case config.ActionPopup:
			log.Printf("  -> Executando Ação POPUP: %s", action.PopupTitle)
			notifications.ShowPopup(action.PopupTitle, action.PopupMessage)

		case config.ActionSound:
			log.Printf("  -> Executando Ação SOUND: tocando %s com volume %.2f", action.SoundFile, actionVolume)
			notifications.ShowDesktopNotification("Alerta de Foco", "É hora de uma pausa. Ouça o aviso sonoro.")
			audio.PlaySound(action.SoundFile, actionVolume)

		case config.ActionATC:
			log.Println("  -> Executando Ação ATC_VOICE")
			executeATC(action.LlamaPrompt, cfg.LlamaModel, level.StaticVolume, actionVolume)
		}
	}
	if level.TriggerHomeAssistant && cfg.HomeAssistantEnabled {
		go integrations.TriggerHomeAssistant(cfg.HomeAssistantWebhookURL, `{"message": "Alerta de hiperfoco acionado."}`)
	}
}

func executeATC(prompt, model string, staticVolume float64, voiceVolume float64) {
	log.Println("Ação ATC_VOICE: Iniciando protocolo.")
	alertText, err := integrations.GenerateTextWithLlama(model, prompt)
	if err != nil {
		log.Printf("Erro ao gerar texto ATC com Llama, usando fallback: %v", err)
		alertText = "Alfa-Um, aqui é a Torre. Desvie de rota imediatamente."
	}
	err = audio.PlayRadioSimulation(alertText, staticVolume, voiceVolume)
	if err != nil {
		fmt.Println("Error:", err)
	}
	if !audio.IsReady() {
		log.Println("Sistema de áudio não inicializado. Usando fallback de pop-up para ATC.")
		notifications.ShowPopup("Torre de Comando para Alfa-Um", alertText)
		return
	}
	go notifications.ShowDesktopNotification("Torre de Comando para Alfa-Um", "Nova transmissão recebida. Verifique seu áudio.")
}
