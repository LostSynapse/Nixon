package api

import (
	"encoding/json"
	"log"
	"net/http"
	"nixon/internal/control"
	"nixon/internal/websocket"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.comcom/go-chi/chi/v5/middleware"
)

// NewRouter creates a new router with all the application's routes.
func NewRouter(ctrl *control.Manager) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// WebSocket endpoint
	r.Get("/ws", websocket.Handler)

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Post("/stream/start", handleStreamStart(ctrl))
		r.Post("/stream/stop", handleStreamStop(ctrl))
		r.Post("/recording/start", handleRecordingStart(ctrl))
		r.Post("/recording/stop", handleRecordingStop(ctrl))
		r.Get("/recordings", handleGetRecordings(ctrl))
		r.Delete("/recording/{id}", handleDeleteRecording(ctrl))
	})

	// Serve the SPA
	workDir, _ := os.Getwd()
	staticPath := filepath.Join(workDir, "web", "dist")
	staticFS := http.Dir(staticPath)
	fileServer := http.StripPrefix("/", http.FileServer(staticFS))

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		// Check if a file exists for the given path
		if _, err := staticFS.Open(r.URL.Path); os.IsNotExist(err) {
			// File does not exist, serve index.html
			http.ServeFile(w, r, filepath.Join(staticPath, "index.html"))
			return
		}
		// Otherwise, serve the static file
		fileServer.ServeHTTP(w, r)
	})

	return r
}

func handleStreamStart(ctrl *control.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Type string `json:"type"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := ctrl.StartStream(body.Type); err != nil {
			http.Error(w, "Failed to start stream", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func handleStreamStop(ctrl *control.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Type string `json:"type"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if err := ctrl.StopStream(body.Type); err != nil {
			http.Error(w, "Failed to stop stream", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func handleRecordingStart(ctrl *control.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := ctrl.StartRecording()
		if err != nil {
			http.Error(w, "Failed to start recording", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]uint{"id": id})
	}
}

func handleRecordingStop(ctrl *control.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := ctrl.StopRecording(); err != nil {
			http.Error(w, "Failed to stop recording", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func handleGetRecordings(ctrl *control.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		recordings, err := ctrl.GetRecordings()
		if err != nil {
			http.Error(w, "Failed to get recordings", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(recordings)
	}
}

func handleDeleteRecording(ctrl *control.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			http.Error(w, "Invalid recording ID", http.StatusBadRequest)
			return
		}
		if err := ctrl.DeleteRecording(uint(id)); err != nil {
			log.Printf("Error deleting recording: %v", err)
			http.Error(w, "Failed to delete recording", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
