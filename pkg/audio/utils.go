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

func playFile(filename string, volume float64) error {
	playCmd := exec.Command("play", "-q", filename, "vol", fmt.Sprintf("%.2f", volume))
	return commands.RunCommand(playCmd)
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

func playSoundAmplified(filename string, volume float64) error {
	var lowerVolumeCmd, restoreVolumeCmd *exec.Cmd
	var originalVolume string
	var err error

	switch runtime.GOOS {
	case "darwin":
		originalVolume, err = getSystemVolumeMac()
		if err != nil {
			originalVolume = "75"
		}
		lowerVolumeCmd = exec.Command("osascript", "-e", "set volume output volume 20")
		restoreVolumeCmd = exec.Command("osascript", "-e", "set volume output volume "+originalVolume)

	case "windows":
		originalVolume = "80%"                                                 // for logging
		lowerVolumeCmd = exec.Command("nircmd.exe", "setsysvolume", "13107")   // ~20%
		restoreVolumeCmd = exec.Command("nircmd.exe", "setsysvolume", "52428") // ~80%

	case "linux":
		sink, err := commands.GetDefaultSinkName()
		if err != nil {
			log.Println("Could not get default sink, skipping volume ducking.")
			return playFile(filename, volume)
		}
		originalVolume, err = getSystemVolumeLinux()
		if err != nil {
			log.Println("Could not get current volume, defaulting to 100%")
			originalVolume = "100%"
		}
		lowerVolumeCmd = exec.Command("pactl", "set-sink-volume", sink, "20%")
		restoreVolumeCmd = exec.Command("pactl", "set-sink-volume", sink, originalVolume)

	default:
		return fmt.Errorf("unsupported OS for this method: %s", runtime.GOOS)
	}

	if err := commands.RunCommand(lowerVolumeCmd); err != nil {
		log.Println("Could not lower system volume, playing normally.")
		return playFile(filename, 1.0)
	}
	defer func() {
		// log.Printf("Restoring system volume to: %s", originalVolume)
		commands.RunCommand(restoreVolumeCmd)
	}()
	return playFile(filename, volume)
}

func playSoundIsolatedLinux(filename string, volume float64) error {
	originalVolume, err := getSystemVolumeLinux()
	if err != nil {
		originalVolume = "100%"
		log.Println("Could not get current volume, defaulting to 100%")
	}

	if err := commands.RunCommand(exec.Command("pactl", "set-sink-volume", "@DEFAULT_SINK@", "20%")); err != nil {
		log.Println("Failed to lower system volume:", err)
	}

	defer func() {
		time.Sleep(100 * time.Millisecond)
		if err := commands.RunCommand(exec.Command("pactl", "set-sink-volume", "@DEFAULT_SINK@", originalVolume)); err != nil {
			log.Println("Failed to restore system volume:", err)
		}
	}()

	return playSoundAmplified(filename, volume)
}

func GetAssetPath(filename string) string {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return filepath.Join("/app", "assets", filename)
	}
	return filepath.Join(config.GetUserConfigPath(), "assets", filename)
}
