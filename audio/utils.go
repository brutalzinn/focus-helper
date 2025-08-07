package audio

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
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
	playCmd := exec.Command("play", "-q", filename, "vol", fmt.Sprintf("%.2f", volume))
	return runCommand(playCmd)
}

func getAudioDuration(filePath string) (time.Duration, error) {
	// This command asks ffprobe for the duration in seconds.
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

	// The output is a string like "12.345600\n", so we trim it.
	durationStr := strings.TrimSpace(string(output))

	// Parse the string into a float64.
	durationFloat, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse duration '%s': %w", durationStr, err)
	}

	// Convert the seconds into a time.Duration.
	return time.Duration(durationFloat * float64(time.Second)), nil
}

func playFileOnSink(filename string, multiplier float64, sink string) error {
	if multiplier <= 0 {
		multiplier = 1.0
	}
	playCmd := exec.Command("play", "-q", filename, "gain", "-n", "vol", fmt.Sprintf("%.2f", multiplier))
	playCmd.Env = append(os.Environ(), "PULSE_SINK="+sink)
	return runCommand(playCmd)
}

func playSoundAmplified(filename string, multiplier float64) error {
	var lowerVolumeCmd, restoreVolumeCmd *exec.Cmd
	var originalVolume string
	var err error

	if runtime.GOOS == "darwin" {
		originalVolume, err = getSystemVolumeMac()
		if err != nil {
			originalVolume = "75"
		} // Default restore volume for mac
		lowerVolumeCmd = exec.Command("osascript", "-e", "set volume output volume 20")
		restoreVolumeCmd = exec.Command("osascript", "-e", "set volume output volume "+originalVolume)

	} else if runtime.GOOS == "windows" {
		originalVolume = "80%"                                                 // For logging purposes, as NirCmd can't get volume easily
		lowerVolumeCmd = exec.Command("nircmd.exe", "setsysvolume", "13107")   // ~20%
		restoreVolumeCmd = exec.Command("nircmd.exe", "setsysvolume", "52428") // ~80%

	} else {
		return fmt.Errorf("unsupported OS for this method: %s", runtime.GOOS)
	}

	if err := runCommand(lowerVolumeCmd); err != nil {
		log.Println("Could not lower system volume, playing normally.")
		return playFile(filename, 1.0)
	}
	defer func() {
		log.Printf("Restoring system volume to: %s", originalVolume)
		runCommand(restoreVolumeCmd)
	}()

	log.Printf("Playing amplified sound with multiplier %.2f", multiplier)
	return playFile(filename, multiplier)
}

func playSoundIsolatedLinux(filename string, multiplier float64) error {
	sinkName := "focus_priority_sink"
	loadSinkCmd := exec.Command("pactl", "load-module", "module-null-sink", "sink_name="+sinkName)
	if err := runCommand(loadSinkCmd); err != nil {
		log.Println("Could not create virtual sink, playing with amplification as fallback.")
		return playSoundAmplified(filename, multiplier)
	}
	defer func() {
		log.Printf("Unloading virtual sink: %s", sinkName)
		time.Sleep(100 * time.Millisecond)
		unloadSinkCmd := exec.Command("pactl", "unload-module", "module-null-sink")
		runCommand(unloadSinkCmd)
	}()
	originalVolume, err := getSystemVolumeLinux()
	if err != nil {
		originalVolume = "100%"
	}
	lowerVolumeCmd := exec.Command("pactl", "set-sink-volume", "@DEFAULT_SINK@", "20%")
	restoreVolumeCmd := exec.Command("pactl", "set-sink-volume", "@DEFAULT_SINK@", originalVolume)
	runCommand(lowerVolumeCmd)
	defer runCommand(restoreVolumeCmd)
	log.Printf("Playing sound on isolated sink '%s' with multiplier %.2f", sinkName, multiplier)
	return playFileOnSink(filename, multiplier, sinkName)
}
