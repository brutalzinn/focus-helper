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
	const playbackGain = "+1"

	switch runtime.GOOS {
	case "linux":
		if !adjustSystemVolume {
			cmd = exec.Command("play", "-q", filename, "vol", fmt.Sprintf("%.2f", volume))
			break
		}

		sink, err := runCommandWithOutput(exec.Command("pactl", "get-default-sink"))
		if err != nil {
			log.Println("Could not get default sink, playing normally:", err)
			cmd = exec.Command("play", "-q", filename, "vol", fmt.Sprintf("%.2f", volume), "gain", playbackGain)
			break
		}

		out, err := runCommandWithOutput(exec.Command("pactl", "get-sink-volume", sink))
		if err != nil {
			log.Println("Could not get current sink volume, playing normally:", err)
			cmd = exec.Command("play", "-q", filename, "vol", fmt.Sprintf("%.2f", volume), "gain", playbackGain)
			break
		}
		originalVolume := strings.Split(strings.Split(out, "/")[1], "%")[0]
		originalVolume = strings.TrimSpace(originalVolume) + "%"

		log.Println("Lowering system volume for playback...")
		err = commands.RunCommand(exec.Command("pactl", "set-sink-volume", sink, "30%"))
		if err != nil {
			log.Printf("Failed to lower system volume: %v. Playing without ducking.", err)
			cmd = exec.Command("play", "-q", filename, "vol", fmt.Sprintf("%.2f", volume), "gain", playbackGain)
			break
		}

		defer func() {
			log.Printf("Restoring system volume to: %s", originalVolume)
			_ = commands.RunCommand(exec.Command("pactl", "set-sink-volume", sink, originalVolume))
		}()
		cmd = exec.Command("play", "-q", filename, "vol", fmt.Sprintf("%.2f", volume), "gain", playbackGain)

	case "darwin":
		if !adjustSystemVolume {
			cmd = exec.Command("afplay", "-v", fmt.Sprintf("%.2f", volume), filename)
			break
		}
		originalVolume, err := runCommandWithOutput(exec.Command("osascript", "-e", "output volume of (get volume settings)"))
		if err != nil {
			log.Println("Could not get current volume, playing normally:", err)
			cmd = exec.Command("afplay", "-v", fmt.Sprintf("%.2f", volume), filename)
			break
		}

		log.Println("Lowering system volume for playback...")
		err = commands.RunCommand(exec.Command("osascript", "-e", "set volume output volume 25"))
		if err != nil {
			log.Printf("Failed to lower system volume: %v. Playing without ducking.", err)
			cmd = exec.Command("afplay", "-v", fmt.Sprintf("%.2f", volume), filename)
			break
		}

		defer func() {
			log.Printf("Restoring system volume to: %s", originalVolume)
			// Use your RunCommand in the defer block
			_ = commands.RunCommand(exec.Command("osascript", "-e", fmt.Sprintf("set volume output volume %s", originalVolume)))
		}()

		cmd = exec.Command("afplay", "-v", fmt.Sprintf("%.2f", volume), filename)

	case "windows":
		if adjustSystemVolume {
			log.Println("Audio ducking on Windows is best-effort. We will boost this app's volume only.")
		}
		cmd = exec.Command("play", "-q", filename, "vol", fmt.Sprintf("%.2f", volume), "gain", playbackGain)

	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	// The main playback command cannot use `RunCommand` because it uses `cmd.Run()`,
	// which is blocking. We need `cmd.Start()` to manage it in a separate goroutine
	// and allow the `stopChan` to interrupt it.
	if err := cmd.Start(); err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-stopChan:
		log.Println("Playback stopped by external signal.")
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		return nil
	case err := <-done:
		if err != nil {
			return fmt.Errorf("playback command failed: %w", err)
		}
		return nil
	}
}

func GetAssetPath(filename string) string {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return filepath.Join("/app", "assets", filename)
	}
	return filepath.Join(config.GetUserConfigPath(), "assets", filename)
}

func runCommandWithOutput(cmd *exec.Cmd) (string, error) {
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running command %v: %w, output: %s", cmd.Args, err, string(output))
	}
	return strings.TrimSpace(string(output)), nil
}
