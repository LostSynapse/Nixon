package api

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"nixon/internal/config"
	"nixon/internal/control"
	"nixon/internal/slogger"
	"nixon/internal/websocket"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

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

// wsAuthMiddleware protects the WebSocket endpoint with a token.
func wsAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			respondWithError(w, http.StatusUnauthorized, errors.New("missing token"), "Authentication token is required")
			return
		}

		secret := config.AppConfig.Web.Secret
		if subtle.ConstantTimeCompare([]byte(token), []byte(secret)) != 1 {
			respondWithError(w, http.StatusForbidden, errors.New("invalid token"), "Invalid authentication token")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// apiRouter creates a new sub-router for all API endpoints. This isolates API logic.
func apiRouter(ctrl *control.Manager) http.Handler {
	r := chi.NewRouter()
	r.Get("/status", handleGetStatus(ctrl))
	r.Post("/stream/start", handleStreamStart(ctrl))
	r.Post("/stream/stop", handleStreamStop(ctrl))
	r.Post("/recording/start", handleRecordingStart(ctrl))
	r.Post("/recording/stop", handleRecordingStop(ctrl))
	r.Get("/recordings", handleGetRecordings(ctrl))
	r.Delete("/recording/{id}", handleDeleteRecording(ctrl))
	return r
}

// NewRouter creates the main router, mounting the API and frontend handlers.
func NewRouter(ctrl *control.Manager) *chi.Mux {
	r := chi.NewRouter()

	// --- Core Middleware ---
	r.Use(NewStructuredLogger(slogger.Log))
	r.Use(middleware.Recoverer)

	// --- Application Routes ---
	r.Mount("/api", apiRouter(ctrl))
	r.With(wsAuthMiddleware).Get("/ws", websocket.Handler)

	// --- Frontend Handling (Proxy for Dev, Static for Prod) ---
	webDevServerURL := config.AppConfig.Web.WebDevServerURL
	if webDevServerURL != "" {
		slogger.Log.Info("Development mode: Proxying frontend to Vite dev server", "url", webDevServerURL)
		remote, err := url.Parse(webDevServerURL)
		if err != nil {
			slogger.Log.Error("Invalid Vite server URL, cannot start proxy.", "err", err, "url", webDevServerURL)
		} else {
			proxy := httputil.NewSingleHostReverseProxy(remote)
			proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
				slogger.Log.Error("Vite dev server proxy error", "err", err, "path", r.URL.Path)
				w.WriteHeader(http.StatusBadGateway)
				w.Header().Set("Content-Type", "text/html")
				fmt.Fprintf(w, `
					<!doctype html><html><head><title>Frontend Dev Server Error</title></head><body>
					<h1>502 Bad Gateway</h1>
					<p>Could not connect to the Vite development server at <strong>%s</strong>.</p>
					<p>Please ensure the frontend dev server is running by executing <code>npm run dev --prefix web</code> in a separate terminal.</p>
					<hr>
					<p><em>Error: %v</em></p>
					</body></html>
				`, webDevServerURL, err)
			}

			r.Handle("/*", proxy)
		}
	} else {
		slogger.Log.Info("Production mode: Serving static frontend files from 'web/dist'")
		workDir, _ := os.Getwd()
		staticDir := filepath.Join(workDir, "web", "dist")
    	fs := http.FileServer(http.Dir(staticDir)) // ADDED: Create the file server
     	r.Get("/*", func(w http.ResponseWriter, r *http.Request) { // ADDED: Catch-all handler
     			// Check if the requested file exists in the static directory.
     			// Note: We need to clean the path to prevent directory traversal attacks with os.Stat
     			requestedPath := filepath.Join(staticDir, filepath.Clean(r.URL.Path))
     			if _, err := os.Stat(requestedPath); os.IsNotExist(err) {
     				// File does not exist, serve index.html for SPA routing.
     				http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
     				return
     			}
     			// File exists, let the file server handle it.
     			fs.ServeHTTP(w, r)
     		}) // ADDED: End of catch-all handler
	}
	return r
}

// --- API Handler Functions ---

func handleGetStatus(ctrl *control.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := ctrl.GetStatus()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	}
}

func handleStreamStart(ctrl *control.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Type string `json:"type" validate:"required,oneof=srt icecast"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			respondWithError(w, http.StatusBadRequest, err, "Invalid request body")
			return
		}
		if err := validate.Struct(body); err != nil {
			respondWithError(w, http.StatusBadRequest, err, "Validation failed: "+err.Error())
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
			Type string `json:"type" validate:"required,oneof=srt icecast"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			respondWithError(w, http.StatusBadRequest, err, "Invalid request body")
			return
		}
		if err := validate.Struct(body); err != nil {
			respondWithError(w, http.StatusBadRequest, err, "Validation failed: "+err.Error())
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
		if err := ctrl.StartRecording(); err != nil {
			respondWithError(w, http.StatusInternalServerError, err, "Failed to start recording")
			return
		}
		w.WriteHeader(http.StatusOK)
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
