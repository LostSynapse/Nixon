package api

import (
	"encoding/json"
	"nixon/internal/slogger"
	"net/http"
	"nixon/internal/control"
	"nixon/internal/websocket"
	"os"
	"path/filepath"
	"strconv"
	"log/slog"
	"time"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)
// NewStructuredLogger creates a new middleware for structured logging with slog.
func NewStructuredLogger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()

			defer func() {
				logger.Info("Request completed",
					"method", r.Method,
					"url", r.URL.String(),
					"status", ww.Status(),
					"bytes", ww.BytesWritten(),
					"latency", time.Since(start),
				)
			}()

			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}

// respondWithError logs the detailed error and sends a standardized JSON error to the client.
func respondWithError(w http.ResponseWriter, status int, err error, message string) {
	slogger.Log.Error(message, "err", err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}


// NewRouter creates a new router with all the application's routes.
func NewRouter(ctrl *control.Manager) *chi.Mux {
	r := chi.NewRouter()

	r.Use(NewStructuredLogger(slogger.Log))
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
			respondWithError(w, http.StatusBadRequest, err, "Invalid request body")
			return
		}

		if err := ctrl.StartStream(body.Type); err != nil {
			respondWithError(w, http.StatusInternalServerError, err, "Failed to start stream")
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
			respondWithError(w, http.StatusBadRequest, err, "Invalid request body")
			return
		}
		if err := ctrl.StopStream(body.Type); err != nil {
			respondWithError(w, http.StatusInternalServerError, err, "Failed to stop stream")
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func handleRecordingStart(ctrl *control.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := ctrl.StartRecording()
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err, "Failed to start recording")
			return
		}
		json.NewEncoder(w).Encode(map[string]uint{"id": id})
	}
}

func handleRecordingStop(ctrl *control.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := ctrl.StopRecording(); err != nil {
			respondWithError(w, http.StatusInternalServerError, err, "Failed to stop recording")
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func handleGetRecordings(ctrl *control.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		recordings, err := ctrl.GetRecordings()
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err, "Failed to get recordings")
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
			respondWithError(w, http.StatusBadRequest, err, "Invalid recording ID")
			return
		}
		if err := ctrl.DeleteRecording(uint(id)); err != nil {
			respondWithError(w, http.StatusInternalServerError, err, "Failed to delete recording")

			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
