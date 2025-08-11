package commands

import (
	"bytes"
	"log"
	"os/exec"
	"strings"
)

func RunCommand(cmd *exec.Cmd) error {
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Printf("Erro no comando: %s\nOutput: %s", cmd.String(), stderr.String())
	}
	return err
}

func GetDefaultSinkName() (string, error) {
	cmd := exec.Command("pactl", "get-default-sink")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
