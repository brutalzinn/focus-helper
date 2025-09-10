
package server

import (
	"context"
	"encoding/json"
	"focus-helper/src/pkg/config"
	"focus-helper/src/pkg/models"
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
	viewsDir := filepath.Join(config.GetUserConfigPath(), config.VIEWS_DIR)
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		log.Printf("Failed to create assets directory: %v", err)
		return
	}
	mux.Handle("/views/", http.StripPrefix("/views/", http.FileServer(http.Dir(viewsDir))))
	mux.Handle("/temp_audio/", http.StripPrefix("/temp_audio/", http.FileServer(http.Dir(tempAudioDir))))
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsDir))))
	mux.HandleFunc("/api/profiles", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			profilePath := filepath.Join(config.GetUserConfigPath(), config.PROFILES_FILE_NAME)
			profiles, err := config.LoadProfiles(profilePath)
			if err != nil {
				http.Error(w, "Failed to load profiles: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(profiles)

		case http.MethodPost:
			var profiles []models.Config
			if err := json.NewDecoder(r.Body).Decode(&profiles); err != nil {
				http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
				return
			}
			if err := config.SaveProfiles(profiles); err != nil {
				http.Error(w, "Failed to save profiles: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

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
