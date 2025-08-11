package audio

import (
	"fmt"
	"focus-helper/pkg/commands"
	"focus-helper/pkg/config"
	"os"
	"os/exec"
	"path/filepath"
)

func ApplyRadioFilter(inputFile, outputFile string) error {
	cmd := exec.Command("sox", inputFile, "-r", "22050", "-c", "1", outputFile, "highpass", "300", "lowpass", "3000", "compand", "0.3,1", "6:-70,-60,-20", "-5", "-90", "0.2", "gain", "-n")
	return commands.RunCommand(cmd)
}

func MixWithBackground(originalFilePath, backgroundAudioPath, finalOutputFile string, backgroundVolume float64, originalVolume float64) error {
	tempBackgroundCropped := filepath.Join(config.GetUserConfigPath(), config.TEMP_AUDIO_DIR, "temp_background_cropped.wav")
	defer os.Remove(tempBackgroundCropped)
	duration, err := getAudioDuration(originalFilePath)
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
	if err := commands.RunCommand(ffmpegCmd); err != nil {
		return fmt.Errorf("erro ao cortar áudio estático: %w", err)
	}
	soxMixCmd := exec.Command("sox",
		"-m",
		"-v", fmt.Sprintf("%.2f", originalVolume), originalFilePath,
		"-v", fmt.Sprintf("%.2f", backgroundVolume), tempBackgroundCropped,
		finalOutputFile,
		"trim", "0", fmt.Sprintf("%.3f", duration.Seconds()),
	)
	if err := commands.RunCommand(soxMixCmd); err != nil {
		return fmt.Errorf("erro ao mixar áudio com sox: %w", err)
	}
	return nil
}
