package logging

import (
	"focus-helper/pkg/models"
	"io"
	"log"
	"os"
)

func SetupLogger(appConfig *models.Config) {
	if appConfig.LogFile == "" {
		appConfig.LogFile = "focus_helper.log"
	}

	f, err := os.OpenFile(appConfig.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	multiWriter := io.MultiWriter(os.Stdout, f)
	log.SetOutput(multiWriter)
}
