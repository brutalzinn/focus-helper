package audio

import (
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
)

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

	// if !IsReady() {
	// 	return nil
	// }
	if volume <= 0 {
		volume = 1.0
	}
	err := playSoundAmplified(filePath, volume)
	return err
}
