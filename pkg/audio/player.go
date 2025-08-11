package audio

import (
	"log"
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
		log.Printf("AVISO: Não foi possível inicializar o sistema de áudio: %v", err)
		audioInitialized = false
	} else {
		audioInitialized = true
	}
}

func IsReady() bool {
	return audioInitialized
}

func PlaySound(filename string, volume float64) error {
	audioMutex.Lock()
	defer audioMutex.Unlock()
	if !IsReady() {
		return nil
	}
	if volume <= 0 {
		log.Println("PlayRadioSimulation Volume deve ser maior que zero, usando volume padrão de 1.0")
		volume = 1.0
	}
	staticAudio := getAssetPath("assets", filename)
	if err := playPrioritySound(staticAudio, volume); err != nil {
		log.Printf("Error playing final audio with ducking: %v", err)
		return nil
	}
	return nil
}

func playPrioritySound(filename string, volume float64) error {
	switch runtime.GOOS {
	case "linux":
		log.Println("Using Linux 'virtual sink' method for priority audio.")
		return playSoundIsolatedLinux(filename, volume)

	case "darwin", "windows":
		log.Printf("Using '%s' 'amplify and lower' method for priority audio.", runtime.GOOS)
		return playSoundAmplified(filename, volume)

	default:
		log.Printf("Priority audio not supported on %s. Playing normally.", runtime.GOOS)
		return playFile(filename, 1.0)
	}
}
