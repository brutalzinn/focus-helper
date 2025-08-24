package audio

import "sync"

var currentAudio struct {
	mu       sync.Mutex
	stopChan chan any
}

func PlaySound(filePath string, volume float64) error {
	stopChan := make(chan any)
	currentAudio.mu.Lock()
	if currentAudio.stopChan != nil {
		close(currentAudio.stopChan)
	}
	currentAudio.stopChan = stopChan
	currentAudio.mu.Unlock()
	err := playFile(filePath, volume, stopChan, false)
	currentAudio.mu.Lock()
	currentAudio.stopChan = nil
	currentAudio.mu.Unlock()
	return err
}

func StopCurrentSound() {
	currentAudio.mu.Lock()
	defer currentAudio.mu.Unlock()
	if currentAudio.stopChan != nil {
		close(currentAudio.stopChan)
		currentAudio.stopChan = nil
	}
}
