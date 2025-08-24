// pkg/utils/utils.go
package utils

import (
	"fmt"
	"focus-helper/pkg/config"
	"log"
	"os"
	"path/filepath"
	"time"
)

func FormatDuration(d time.Duration, hUnit, mUnit, sUnit string) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	if h > 0 {
		return fmt.Sprintf("%d%s %d%s %d%s", h, hUnit, m, mUnit, s, sUnit)
	}
	if m > 0 {
		return fmt.Sprintf("%d%s %d%s", m, mUnit, s, sUnit)
	}
	return fmt.Sprintf("%d%s", s, sUnit)
}

func ClearTempAudioOnExit() {
	tempAudioDir := filepath.Join(config.GetUserConfigPath(), config.TEMP_AUDIO_DIR)
	if _, err := os.Stat(tempAudioDir); os.IsNotExist(err) {
		return
	}

	err := os.RemoveAll(tempAudioDir)
	if err != nil {
		log.Printf("Error clearing temp_audio: %v", err)
	} else {
		fmt.Println("All files inside temp_audio have been cleared.")
	}
}
