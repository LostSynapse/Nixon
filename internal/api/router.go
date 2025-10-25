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
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

	// Serve static files for the SPA
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "web", "dist"))
	FileServer(r, "/", filesDir)

	return r
}

// FileServer serves static files from a http.FileSystem.
// It falls back to serving index.html for any request that doesn't match a file.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		// Check if the file exists at the root of the static files directory
		f, err := root.Open(r.URL.Path)
		if os.IsNotExist(err) {
			// File does not exist, serve index.html
			http.ServeFile(w, r, filepath.Join("web", "dist", "index.html"))
			return
		} else if err != nil {
			// An error occurred, return a 500
			http.Error(w, http.StatusText(500), 500)
			return
		}
		f.Close() // We just checked for existence, close the file

		// File exists, serve it
		fs.ServeHTTP(w, r)
	})
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
