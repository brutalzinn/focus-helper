package audio

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
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

func PlaySound(filename string, volume float64) {
	audioMutex.Lock()
	defer audioMutex.Unlock()
	if !IsReady() {
		return
	}
	if err := playFile(filename, 1.0); err != nil {
		log.Printf("Error playing final audio: %v", err)
	}
}

func PlayRadioSimulation(message string, staticVolume float64, voiceVolume float64) error {
	audioMutex.Lock()
	defer audioMutex.Unlock()
	if !IsReady() {
		log.Println("Sistema de áudio não inicializado, pulando simulação de rádio.")
		return nil
	}
	modelDir := "voices"
	modelName := "pt_BR-cadu-medium"
	assetsDir := "assets"

	modelPath := filepath.Join(modelDir, modelName+".onnx")
	configPath := filepath.Join(modelDir, modelName+".json")
	staticAudio := filepath.Join(assetsDir, "radio_static.wav")

	tempVoiceRaw := filepath.Join(assetsDir, "temp_voice_raw.wav")
	tempVoiceFiltered := filepath.Join(assetsDir, "temp_voice_filtered.wav")
	tempStaticCropped := filepath.Join(assetsDir, "temp_static_cropped.wav")

	finalOutput := filepath.Join(assetsDir, "final_radio_output.wav")

	piperCmd := exec.Command("piper",
		"--model", modelPath,
		"--config", configPath,
		"--output_file", tempVoiceRaw,
	)
	piperCmd.Stdin = bytes.NewBufferString(message)
	if err := runCommand(piperCmd); err != nil {
		return fmt.Errorf("erro ao executar piper: %w", err)
	}

	soxFilterCmd := exec.Command("sox",
		tempVoiceRaw,
		"-r", "22050",
		"-c", "1",
		tempVoiceFiltered,
		"highpass", "300",
		"lowpass", "3000",
		"compand", "0.3,1", "6:-70,-60,-20", "-5", "-90", "0.2",
		"gain", "-n",
	)
	if err := runCommand(soxFilterCmd); err != nil {
		return fmt.Errorf("erro ao aplicar efeitos de rádio com sox: %w", err)
	}

	duration, err := getAudioDuration(tempVoiceFiltered)
	if err != nil {
		return err
	}

	ffmpegCmd := exec.Command("ffmpeg", "-y",
		"-i", staticAudio,
		"-t", fmt.Sprintf("%.3f", duration.Seconds()),
		"-ar", "22050",
		"-ac", "1",
		tempStaticCropped,
	)
	if err := runCommand(ffmpegCmd); err != nil {
		return fmt.Errorf("erro ao cortar áudio estático: %w", err)
	}

	soxMixCmd := exec.Command("sox",
		"-m",
		"-v", fmt.Sprintf("%.2f", voiceVolume), tempVoiceFiltered,
		"-v", fmt.Sprintf("%.2f", staticVolume), tempStaticCropped,
		finalOutput,
	)
	if err := runCommand(soxMixCmd); err != nil {
		return fmt.Errorf("erro ao mixar áudio com sox: %w", err)
	}

	if err := playFile(finalOutput, 1.0); err != nil {
		log.Printf("Error playing final audio: %v", err)
	}

	_ = os.Remove(tempVoiceRaw)
	_ = os.Remove(tempVoiceFiltered)
	_ = os.Remove(tempStaticCropped)
	_ = os.Remove(finalOutput)
	return nil
}

func runCommand(cmd *exec.Cmd) error {
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Printf("Erro no comando: %s\nOutput: %s", cmd.String(), stderr.String())
	}
	return err
}

func playFile(filename string, volume float64) error {
	playCmd := exec.Command("play", filename, "vol", fmt.Sprintf("%.2f", volume))
	return runCommand(playCmd)
}

func getAudioDuration(filePath string) (time.Duration, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filePath,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("erro ao executar ffprobe: %w", err)
	}

	durationStr := strings.TrimSpace(string(output))
	durationFloat, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("erro ao converter duração para float: %w", err)
	}

	return time.Duration(durationFloat * float64(time.Second)), nil
}
