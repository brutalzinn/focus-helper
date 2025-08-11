package audio

import (
	"fmt"
	"focus-helper/pkg/commands"
	"os/exec"
)

func ApplyRadioFilter(inputFile, outputFile string) error {
	cmd := exec.Command("sox", inputFile, "-r", "22050", "-c", "1", outputFile, "highpass", "300", "lowpass", "3000", "compand", "0.3,1", "6:-70,-60,-20", "-5", "-90", "0.2", "gain", "-n")
	return commands.RunCommand(cmd)
}

func MixWithBackground(originalFilePath, backgroundAudioPath, finalOutputFile string, backgroundVolume float64) error {
	convertedOriginal := originalFilePath + "_converted.wav"
	convertedBackground := backgroundAudioPath + "_converted.wav"
	cmd1 := exec.Command("sox", originalFilePath, "-r", "44100", "-c", "2", convertedOriginal)
	if err := commands.RunCommand(cmd1); err != nil {
		return fmt.Errorf("erro convertendo original: %w", err)
	}
	cmd2 := exec.Command("sox", backgroundAudioPath, "-r", "44100", "-c", "2", convertedBackground)
	if err := commands.RunCommand(cmd2); err != nil {
		return fmt.Errorf("erro convertendo background: %w", err)
	}
	cmdMix := exec.Command(
		"sox", "-m",
		"-v", "1.0", convertedOriginal,
		"-v", fmt.Sprintf("%.2f", backgroundVolume),
		 convertedBackground,
		finalOutputFile,
	)
	if err := commands.RunCommand(cmdMix); err != nil {
		return fmt.Errorf("erro no mix: %w", err)
	}
	return nil
}

// func PlayRadioSimulation(message string, voiceVolume, backgroundVolume float64, backgroundSound string) error {
// 	audioMutex.Lock()
// 	if !IsReady() {
// 		log.Println("Sistema de áudio não inicializado, pulando simulação de rádio.")
// 		return nil
// 	}

// 	if voiceVolume <= 0 {
// 		voiceVolume = 1.0
// 	}

// 	modelPath := getAssetPath("voices", "pt_BR-cadu-medium.onnx")
// 	configPath := getAssetPath("voices", "pt_BR-cadu-medium.onnx.json")

// 	tempVoiceRaw := getAssetPath("assets", "temp_voice_raw.wav")
// 	tempVoiceFiltered := getAssetPath("assets", "temp_voice_filtered.wav")
// 	tempBackgroundCropped := getAssetPath("assets", "temp_background_cropped.wav")
// 	finalOutput := getAssetPath("assets", "final_radio_output.wav")
// 	backgroundAudioPath := getAssetPath("assets", backgroundSound)

// 	defer func() {
// 		audioMutex.Unlock()
// 		log.Println("Limpando arquivos de áudio temporários...")
// 		_ = os.Remove(tempVoiceRaw)
// 		_ = os.Remove(tempVoiceFiltered)
// 		_ = os.Remove(tempBackgroundCropped)
// 		_ = os.Remove(finalOutput)
// 	}()

// 	piperCmd := exec.Command("piper", "--model", modelPath, "--config", configPath, "--output_file", tempVoiceRaw)
// 	piperCmd.Stdin = bytes.NewBufferString(message)
// 	if err := runCommand(piperCmd); err != nil {
// 		return fmt.Errorf("erro ao executar piper: %w", err)
// 	}

// 	soxFilterCmd := exec.Command("sox", tempVoiceRaw, "-r", "22050", "-c", "1", tempVoiceFiltered, "highpass", "300", "lowpass", "3000", "compand", "0.3,1", "6:-70,-60,-20", "-5", "-90", "0.2", "gain", "-n")
// 	if err := runCommand(soxFilterCmd); err != nil {
// 		return fmt.Errorf("erro ao aplicar efeitos de rádio com sox: %w", err)
// 	}

// 	if backgroundSound == "" {
// 		log.Println("Nenhum som de fundo especificado, tocando apenas a voz ATC.")
// 		return playPrioritySound(tempVoiceFiltered, voiceVolume)
// 	}

// 	if backgroundVolume <= 0 {
// 		backgroundVolume = 0.5
// 	}

// 	duration, err := getAudioDuration(tempVoiceFiltered)
// 	if err != nil {
// 		return err
// 	}

// 	ffmpegCmd := exec.Command("ffmpeg", "-y",
// 		"-stream_loop", "-1",
// 		"-i", backgroundAudioPath,
// 		"-t", fmt.Sprintf("%.3f", duration.Seconds()),
// 		"-ar", "22050",
// 		"-ac", "1",
// 		tempBackgroundCropped,
// 	)
// 	log.Printf("Executando comando FFmpeg para criar fundo com loop: %s", ffmpegCmd.String())
// 	if err := runCommand(ffmpegCmd); err != nil {
// 		return fmt.Errorf("erro ao cortar áudio estático: %w", err)
// 	}

// 	soxMixCmd := exec.Command("sox",
// 		"-m",
// 		"-v", fmt.Sprintf("%.2f", voiceVolume), tempVoiceFiltered,
// 		"-v", fmt.Sprintf("%.2f", backgroundVolume), tempBackgroundCropped,
// 		finalOutput,
// 		"trim", "0", fmt.Sprintf("%.3f", duration.Seconds()),
// 	)
// 	if err := runCommand(soxMixCmd); err != nil {
// 		return fmt.Errorf("erro ao mixar áudio com sox: %w", err)
// 	}

// 	if err := playPrioritySound(finalOutput, 1.0); err != nil {
// 		log.Printf("Error playing final audio with ducking: %v", err)
// 	}
// 	return nil
// }
