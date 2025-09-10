
package utils

import (
	"fmt"
	"focus-helper/src/pkg/config"
	"log"
	"math/rand"
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

func GenerateRandomOdd(min, max int) int {
	if min > max {
		min, max = max, min
	}

	for {
		num := rand.Intn(max-min+1) + min
		if num%2 != 0 {
			return num
		}
	}
}

func GenerateRandomSquareIntervalVaried(minSec, maxSec int) (int, int) {
	intervalLengths := []int{10, 15, 20, 30}
	intervalLength := intervalLengths[rand.Intn(len(intervalLengths))]
	if maxSec-minSec < intervalLength {
		return minSec, maxSec
	}
	maxStart := maxSec - intervalLength
	start := rand.Intn(maxStart-minSec+1) + minSec
	start = (start / intervalLength) * intervalLength
	end := start + intervalLength
	if end > maxSec {
		end = maxSec
		start = end - intervalLength
	}
	return start, end
}

func FormatTime(seconds int) string {
	minutes := seconds / 60
	secs := seconds % 60
	return fmt.Sprintf("%d:%02d", minutes, secs)
}
