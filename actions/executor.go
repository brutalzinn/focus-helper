package actions

import (
	"log"
	"time"

	"github.com/brutalzinn/focus-helper/config"
)

func Execute(alert config.AlertLevel, hyperfocusState *config.HyperfocusState) {
	log.Printf("Executando ações para o nível de alerta: %s", alert.Level)
	repetitions := int(alert.Multiplier)
	if repetitions <= 0 {
		repetitions = 1
	}
	log.Printf("Nível de agressividade: %d repetições para ações de áudio/ATC.", repetitions)
	for i := 0; i < repetitions; i++ {
		if repetitions > 1 {
			log.Printf("--> Executando ciclo de ações %d de %d", i+1, repetitions)
		}
		for _, actionCfg := range alert.Actions {
			isAudioAction := actionCfg.Type == config.ActionATC
			if !isAudioAction && i > 0 {
				continue
			}
			action, err := NewActionFromConfig(alert, actionCfg)
			if err != nil {
				log.Printf("Erro ao criar ação: %v", err)
				continue
			}
			go action.Execute(alert)
		}
		if repetitions > 1 && i < repetitions-1 {
			time.Sleep(5 * time.Second)
		}
	}
}
