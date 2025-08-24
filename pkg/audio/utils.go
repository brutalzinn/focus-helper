package audio

import (
	"fmt"
	"focus-helper/pkg/commands"
	"focus-helper/pkg/config"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func getSystemVolumeLinux() (string, error) {
	cmdStr := "pactl list sinks | grep 'Volume:' | head -n1 | cut -d'/' -f2 | tr -d ' %'"
	cmd := exec.Command("bash", "-c", cmdStr)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get linux volume: %w", err)
	}

	volume := strings.TrimSpace(string(output))
	if _, err := strconv.Atoi(volume); err != nil {
		return "", fmt.Errorf("could not parse volume: %s", volume)
	}

	return volume + "%", nil
}

func getSystemVolumeMac() (string, error) {
	cmd := exec.Command("osascript", "-e", "output volume of (get volume settings)")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get macos volume: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func getAudioDuration(filePath string) (time.Duration, error) {
	cmd := exec.Command("ffprobe",
		"-i", filePath,
		"-show_entries", "format=duration",
		"-v", "quiet",
		"-of", "csv=p=0",
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("error running ffprobe on %s: %w", filePath, err)
	}
	durationStr := strings.TrimSpace(string(output))
	durationFloat, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse duration '%s': %w", durationStr, err)
	}
	return time.Duration(durationFloat * float64(time.Second)), nil
}
func playFile(filename string, volume float64, stopChan chan any, adjustSystemVolume bool) error {
	var cmd *exec.Cmd
	var lowerVolumeCmd, restoreVolumeCmd *exec.Cmd
	var originalVolume string

	switch runtime.GOOS {
	case "linux":
		sink, err := commands.GetDefaultSinkName()
		if err != nil || !adjustSystemVolume {
			log.Println("Could not get default sink or volume adjustment disabled, playing normally")
			cmd = exec.Command("play", "-q", filename, "vol", fmt.Sprintf("%.2f", volume))
			break
		}
		originalVolume, err = getSystemVolumeLinux()
		if err != nil {
			originalVolume = "100%"
		}
		lowerVolumeCmd = exec.Command("pactl", "set-sink-volume", sink, "20%")
		restoreVolumeCmd = exec.Command("pactl", "set-sink-volume", sink, originalVolume)
		_ = commands.RunCommand(lowerVolumeCmd)
		defer func() {
			log.Printf("Restoring system volume to: %s", originalVolume)
			_ = commands.RunCommand(restoreVolumeCmd)
		}()
		cmd = exec.Command("play", "-q", filename, "vol", fmt.Sprintf("%.2f", volume))
	case "darwin":
		cmd = exec.Command("afplay", filename)
	case "windows":
		cmd = exec.Command("powershell", "-c", fmt.Sprintf(`(New-Object Media.SoundPlayer "%s").PlaySync()`, filename))
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case <-stopChan:
		log.Println("Playback stopped by StopCurrentSound")
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		return nil
	case err := <-done:
		return err
	}
}

func GetAssetPath(filename string) string {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return filepath.Join("/app", "assets", filename)
	}
	return filepath.Join(config.GetUserConfigPath(), "assets", filename)
}
