package api

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	"net/http/httputil" // ADDED: For reverse proxy
    "net/url"           // ADDED: For parsing URLs
	"nixon/internal/config"
	"nixon/internal/control"
	"nixon/internal/slogger"
	"nixon/internal/websocket"
	"os"
	"path/filepath"
	"strconv"
	"time"
	
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
		// Get the token from the ?token= query parameter.
		token := r.URL.Query().Get("token")
		if token == "" {
			respondWithError(w, http.StatusUnauthorized, errors.New("missing token"), "Authentication token is required")
			return
		}

		// Get the secret from the application configuration.
		secret := config.AppConfig.Web.Secret

		// Use subtle.ConstantTimeCompare to prevent timing attacks.
		if subtle.ConstantTimeCompare([]byte(token), []byte(secret)) != 1 {
			respondWithError(w, http.StatusForbidden, errors.New("invalid token"), "Invalid authentication token")
			return
		}

		// If the token is valid, proceed to the actual WebSocket handler.
		next.ServeHTTP(w, r)
	})
}

// NewRouter creates a new router with all the application's routes.
func NewRouter(ctrl *control.Manager, webDevServerURL string) *chi.Mux { // MODIFIED: Added webDevServerURL param

	r := chi.NewRouter()

	r.Use(NewStructuredLogger(slogger.Log))
	r.Use(middleware.Recoverer)

	// WebSocket endpoint
	r.With(wsAuthMiddleware).Get("/ws", websocket.Handler)

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Post("/stream/start", handleStreamStart(ctrl))
		r.Post("/stream/stop", handleStreamStop(ctrl))
		r.Post("/recording/start", handleRecordingStart(ctrl))
		r.Post("/recording/stop", handleRecordingStop(ctrl))
		r.Get("/recordings", handleGetRecordings(ctrl))
		r.Delete("/recording/{id}", handleDeleteRecording(ctrl))
		r.Get("/status", handleGetStatus(ctrl))
	})

    // --- Serve the SPA / Frontend Development Proxy ---
    if webDevServerURL != "" {
        slogger.Log.Info("Development mode: Proxying frontend to Vite dev server", "url", webDevServerURL)
        remote, err := url.Parse(webDevServerURL)
        if err != nil {
            slogger.Log.Error("Failed to parse web development server URL", "err", err, "url", webDevServerURL)
            // Fallback to static serving if URL is invalid, although it might not work in dev.
            // In a real scenario, this might be a fatal error or clearer error page.
            workDir, _ := os.Getwd()
            http.ServeFile(nil, nil, filepath.Join(workDir, "web", "index.html")) // Serve raw index.html as fallback
            return nil // Or handle error appropriately
        }

        proxy := httputil.NewSingleHostReverseProxy(remote)
        r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Update the request to point to the proxy target
            // Optionally, add X-Forwarded-For header
            r.Header.Set("X-Forwarded-For", r.RemoteAddr)
            proxy.ServeHTTP(w, r)
        }))
    } else {
        slogger.Log.Info("Production mode: Serving static frontend files from 'web/dist'")
        workDir, _ := os.Getwd()
        staticFilesPath := http.Dir(filepath.Join(workDir, "web", "dist"))
        staticFileServer := http.FileServer(staticFilesPath)

        // Serve static files (like JS, CSS, images) from the /assets directory
        r.Handle("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir(filepath.Join(workDir, "web", "dist", "assets")))))

        // Serve other specific root files if needed (like favicon, logo)
        r.Handle("/nixon_logo.svg", staticFileServer) // Assumes nixon_logo.svg is in web/dist or accessible from staticFilesPath

        // For all other requests that are not API calls, serve the SPA's entry point.
        r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
            http.ServeFile(w, r, filepath.Join(workDir, "web", "dist", "index.html"))
        })
    }

	return r
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

func handleGetStatus(ctrl *control.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := ctrl.GetStatus()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	}
}
