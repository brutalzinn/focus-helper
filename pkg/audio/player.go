package audio

import (
	"runtime"
	"sync"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
)

var audioMutex sync.Mutex
var audioInitialized bool

func InitSpeaker() {
	sampleRate := beep.SampleRate(44100)
	err := speaker.Init(sampleRate, sampleRate.N(time.Second/10))
	if err != nil {
		audioInitialized = false
	} else {
		audioInitialized = true
	}
}

func IsReady() bool {
	return audioInitialized
}

func PlaySound(filePath string, volume float64) error {
	audioMutex.Lock()
	defer audioMutex.Unlock()
	// if !IsReady() {
	// 	return nil
	// }
	if volume <= 0 {
		volume = 1.0
	}
	err := playPrioritySound(filePath, volume)
	return err
}

func playPrioritySound(filename string, volume float64) error {
	switch runtime.GOOS {
	case "linux":
		return playSoundIsolatedLinux(filename, volume)
	case "darwin", "windows":
		return playSoundAmplified(filename, volume)
	default:
		return playFile(filename, 1.0)
	}
}
