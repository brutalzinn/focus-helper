package actions

import (
	"log"

	"github.com/brutalzinn/focus-helper/audio"
	"github.com/brutalzinn/focus-helper/config"
)

type SoundAction struct {
	FilePath   string
	Multiplier float64
}

func (a *SoundAction) Execute(alert config.AlertLevel) error {
	log.Printf("  -> Executando SoundAction: %s", a.FilePath)
	return audio.PlaySound(a.FilePath, a.Multiplier)
}
