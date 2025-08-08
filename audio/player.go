package audio

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
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
	staticAudio := getAssetPath("assets", filename)
	if err := playPrioritySound(staticAudio, volume); err != nil {
		log.Printf("Error playing final audio with ducking: %v", err)
		return nil
	}
	return nil
}

func PlayRadioSimulation(message string, backgroundVolume, voiceVolume, playbackMultiplier float64, backgroundSound string) error {
	audioMutex.Lock()
	if !IsReady() {
		log.Println("Sistema de áudio não inicializado, pulando simulação de rádio.")
		return nil
	}

	modelPath := getAssetPath("voices", "pt_BR-cadu-medium.onnx")
	configPath := getAssetPath("voices", "pt_BR-cadu-medium.onnx.json")

	tempVoiceRaw := getAssetPath("assets", "temp_voice_raw.wav")
	tempVoiceFiltered := getAssetPath("assets", "temp_voice_filtered.wav")
	tempBackgroundCropped := getAssetPath("assets", "temp_background_cropped.wav")
	finalOutput := getAssetPath("assets", "final_radio_output.wav")
	backgroundAudioPath := getAssetPath("assets", backgroundSound)

	defer func() {
		audioMutex.Unlock()
		log.Println("Limpando arquivos de áudio temporários...")
		_ = os.Remove(tempVoiceRaw)
		_ = os.Remove(tempVoiceFiltered)
		_ = os.Remove(tempBackgroundCropped)
		_ = os.Remove(finalOutput)
	}()

	piperCmd := exec.Command("piper", "--model", modelPath, "--config", configPath, "--output_file", tempVoiceRaw)
	piperCmd.Stdin = bytes.NewBufferString(message)
	if err := runCommand(piperCmd); err != nil {
		return fmt.Errorf("erro ao executar piper: %w", err)
	}

	soxFilterCmd := exec.Command("sox", tempVoiceRaw, "-r", "22050", "-c", "1", tempVoiceFiltered, "highpass", "300", "lowpass", "3000", "compand", "0.3,1", "6:-70,-60,-20", "-5", "-90", "0.2", "gain", "-n")
	if err := runCommand(soxFilterCmd); err != nil {
		return fmt.Errorf("erro ao aplicar efeitos de rádio com sox: %w", err)
	}

	if backgroundSound == "" {
		log.Println("Nenhum som de fundo especificado, tocando apenas a voz ATC.")
		return playPrioritySound(tempVoiceFiltered, playbackMultiplier)
	}

	duration, err := getAudioDuration(tempVoiceFiltered)
	if err != nil {
		return err
	}

	ffmpegCmd := exec.Command("ffmpeg", "-y",
		"-stream_loop", "-1",
		"-i", backgroundAudioPath,
		"-t", fmt.Sprintf("%.3f", duration.Seconds()),
		"-ar", "22050",
		"-ac", "1",
		tempBackgroundCropped,
	)
	log.Printf("Executando comando FFmpeg para criar fundo com loop: %s", ffmpegCmd.String())
	if err := runCommand(ffmpegCmd); err != nil {
		return fmt.Errorf("erro ao cortar áudio estático: %w", err)
	}

	soxMixCmd := exec.Command("sox",
		"-m",
		"-v", fmt.Sprintf("%.2f", voiceVolume), tempVoiceFiltered,
		"-v", fmt.Sprintf("%.2f", backgroundVolume), tempBackgroundCropped,
		finalOutput,
		"trim", "0", fmt.Sprintf("%.3f", duration.Seconds()),
	)
	if err := runCommand(soxMixCmd); err != nil {
		return fmt.Errorf("erro ao mixar áudio com sox: %w", err)
	}

	if err := playPrioritySound(finalOutput, playbackMultiplier); err != nil {
		log.Printf("Error playing final audio with ducking: %v", err)
	}
	return nil
}

func playPrioritySound(filename string, multiplier float64) error {
	switch runtime.GOOS {
	case "linux":
		log.Println("Using Linux 'virtual sink' method for priority audio.")
		return playSoundIsolatedLinux(filename, multiplier)

	case "darwin", "windows":
		log.Printf("Using '%s' 'amplify and lower' method for priority audio.", runtime.GOOS)
		return playSoundAmplified(filename, multiplier)

	default:
		log.Printf("Priority audio not supported on %s. Playing normally.", runtime.GOOS)
		return playFile(filename, 1.0)
	}
}
