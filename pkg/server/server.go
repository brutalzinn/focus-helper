// pkg/server/server.go
package server

import (
	"focus-helper/pkg/config"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func StartServer() {
	tempAudioDir := filepath.Join(config.GetUserConfigPath(), config.TEMP_AUDIO_DIR)
	if err := os.MkdirAll(tempAudioDir, 0755); err != nil {
		log.Fatalf("Failed to create temp audio directory: %v", err)
	}

	assetsDir := filepath.Join(config.GetUserConfigPath(), config.ASSETS_DIR)
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		log.Fatalf("Failed to create assets directory: %v", err)
	}

	http.Handle("/temp_audio/", http.StripPrefix("/temp_audio/", http.FileServer(http.Dir(tempAudioDir))))
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsDir))))

	log.Printf("Running audio server at http://localhost:%s", config.SERVER_PORT)
	if err := http.ListenAndServe(":"+config.SERVER_PORT, nil); err != nil {
		log.Fatalf("Fail to start audio server: %v", err)
	}
}
