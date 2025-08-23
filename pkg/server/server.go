// pkg/server/server.go
package server

import (
	"context"
	"focus-helper/pkg/config"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func StartServer(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	mux := http.NewServeMux()
	tempAudioDir := filepath.Join(config.GetUserConfigPath(), config.TEMP_AUDIO_DIR)
	if err := os.MkdirAll(tempAudioDir, 0755); err != nil {
		log.Printf("Failed to create temp audio directory: %v", err)
		return
	}
	assetsDir := filepath.Join(config.GetUserConfigPath(), config.ASSETS_DIR)
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		log.Printf("Failed to create assets directory: %v", err)
		return
	}
	mux.Handle("/temp_audio/", http.StripPrefix("/temp_audio/", http.FileServer(http.Dir(tempAudioDir))))
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsDir))))
	server := &http.Server{
		Addr:    ":" + config.SERVER_PORT,
		Handler: mux,
	}
	go func() {
		log.Printf("Running audio server at http://localhost:%s", config.SERVER_PORT)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe error: %v", err)
		}
	}()
	<-ctx.Done()
	log.Println("Shutting down the audio server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Audio server shutdown failed: %v", err)
	}

	log.Println("Audio server shut down gracefully.")
}
