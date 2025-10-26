package api

import (
	"encoding/json"
	"nixon/internal/logger"
	"net/http"
	"nixon/internal/control"
	"nixon/internal/websocket"
	"os"
	"path/filepath"
	"strconv"
	"github.com/rs/zerolog"
	"time"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)
// NewStructuredLogger creates a new middleware for structured logging with Zerolog.
func NewStructuredLogger(logger zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()

			defer func() {
				logger.Info().
					Str("method", r.Method).
					Stringer("url", r.URL).
					Int("status", ww.Status()).
					Int("bytes", ww.BytesWritten()).
					Dur("latency_ms", time.Since(start)).
					Msg("Request completed")
			}()

			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}

// NewRouter creates a new router with all the application's routes.
func NewRouter(ctrl *control.Manager) *chi.Mux {
	r := chi.NewRouter()

	r.Use(NewStructuredLogger(logger.Log))
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
	staticDir := http.Dir(filepath.Join(workDir, "web", "dist"))
	fileServer := http.FileServer(staticDir)

	// Serve static assets
	r.Handle("/assets/*", fileServer)
	r.Handle("/nixon_logo.svg", fileServer)

	// For all other requests, serve the SPA's entry point
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(workDir, "web", "dist", "index.html"))
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
			logger.Log.Error().Err(err).Uint64("id", id).Msg("Error deleting recording")
			http.Error(w, "Failed to delete recording", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
