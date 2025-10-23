// internal/api/router.go
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm" // A6: For error checking
	"log"
	"net/http"
	"nixon/internal/config"
	"nixon/internal/db" // A6: Use GORM db package
	"nixon/internal/gstreamer"
	"nixon/internal/websocket"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	// A1: Using standard library http, no external router needed
)

// NewRouter creates and configures the main application router using net/http (A1)
func NewRouter() http.Handler {
	mux := http.NewServeMux()

	// --- API Routes ---
	mux.HandleFunc("GET /api/config", GetConfigHandler)
	mux.HandleFunc("POST /api/config", UpdateConfigHandler)
	mux.HandleFunc("GET /api/capabilities", GetAudioCapabilitiesHandler) // Matches GET /api/capabilities?device=...
	mux.HandleFunc("GET /api/status", GetStreamStatusHandler)
	mux.HandleFunc("POST /api/stream", SetStreamStateHandler) // Body: {"stream": "srt|icecast", "enabled": true|false}
	mux.HandleFunc("POST /api/record/start", StartRecordingHandler)
	mux.HandleFunc("POST /api/record/stop", StopRecordingHandler)

	// Recording CRUD Routes (A6 GORM) - Use Go 1.22+ PathValue style
	mux.HandleFunc("GET /api/recordings", GetRecordingsHandler)             // GET /api/recordings
	mux.HandleFunc("PUT /api/recordings/{id}", UpdateRecordingHandler)      // PUT /api/recordings/123
	mux.HandleFunc("POST /api/recordings/{id}/protect", ToggleProtectHandler) // POST /api/recordings/123/protect
	mux.HandleFunc("DELETE /api/recordings/{id}", DeleteRecordingHandler)   // DELETE /api/recordings/123

	// --- WebSocket Route ---
	mux.HandleFunc("/ws", websocket.HandleConnections)

	// --- Static File Server & SPA Handler ---
	staticDir := "./web" // Serve from ./web as per manual_setup.md
	fileServer := http.FileServer(http.Dir(staticDir))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "..") { http.Error(w, "Invalid path", http.StatusBadRequest); return }
		requestedPath := filepath.Join(staticDir, r.URL.Path)
		stat, err := os.Stat(requestedPath)

		if os.IsNotExist(err) || (err == nil && stat.IsDir()) {
			indexPath := filepath.Join(staticDir, "index.html")
			if _, indexErr := os.Stat(indexPath); indexErr != nil {
				http.Error(w, "index.html not found", http.StatusNotFound)
				log.Printf("CRITICAL ERROR: index.html not found in static directory '%s'", staticDir)
				return
			}
			http.ServeFile(w, r, indexPath)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Use StripPrefix for serving assets correctly
		http.StripPrefix("/", fileServer).ServeHTTP(w, r)
	})

	log.Println("Using net/http router (A1)")
	return mux
}

// --- Handler Implementations using net/http ---

func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			// Log error, but headers might already be sent
			log.Printf("Error encoding JSON response: %v", err)
		}
	}
}

func writeOK(w http.ResponseWriter) {
	writeJSONResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeError(w http.ResponseWriter, statusCode int, message string, err error) {
	log.Printf("API Error %d: %s (%v)", statusCode, message, err) // Log the underlying error
	http.Error(w, message, statusCode)                            // Send user-friendly message
}

// GetConfigHandler handles GET /api/config
func GetConfigHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.GetConfig()
	writeJSONResponse(w, http.StatusOK, cfg)
}

// UpdateConfigHandler handles POST /api/config
func UpdateConfigHandler(w http.ResponseWriter, r *http.Request) {
	var newConfig config.Config
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	defer r.Body.Close()

	if err := config.SetGlobalConfig(&newConfig); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update config in memory", err)
		return
	}
	if err := config.SaveGlobalConfig(); err != nil {
		// Log error, but proceed with OK response as config is updated in memory
		log.Printf("Warning: Failed to save config to disk after update: %v", err)
	}

	if manager := gstreamer.GetManager(); manager != nil { go manager.RestartPipeline() }
	go websocket.BroadcastUpdate(config.GetConfig()) // Broadcast the updated config
	writeOK(w)
}

// GetAudioCapabilitiesHandler handles GET /api/capabilities?device=...
func GetAudioCapabilitiesHandler(w http.ResponseWriter, r *http.Request) {
	device := r.URL.Query().Get("device")
	caps, err := gstreamer.GetAudioCapabilities(device)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get capabilities for '%s'", device), err)
		return
	}
	writeJSONResponse(w, http.StatusOK, caps)
}

// GetStreamStatusHandler handles GET /api/status (A3)
func GetStreamStatusHandler(w http.ResponseWriter, r *http.Request) {
	manager := gstreamer.GetManager()
	if manager == nil { writeError(w, http.StatusServiceUnavailable, "GStreamer manager not initialized", nil); return }
	writeJSONResponse(w, http.StatusOK, manager.GetStatus())
}

// SetStreamStateHandler handles POST /api/stream
func SetStreamStateHandler(w http.ResponseWriter, r *http.Request) {
	var req struct { Stream string `json:"stream"`; Enabled bool `json:"enabled"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { writeError(w, http.StatusBadRequest, "Invalid request body", err); return }
	defer r.Body.Close()
	manager := gstreamer.GetManager(); if manager == nil { writeError(w, http.StatusServiceUnavailable, "GStreamer manager not initialized", nil); return }

	var err error
	switch req.Stream {
	case "srt": err = manager.ToggleSrtStream(req.Enabled)
	case "icecast": err = manager.ToggleIcecastStream(req.Enabled)
	default: writeError(w, http.StatusBadRequest, "Invalid stream type ('srt' or 'icecast')", nil); return
	}
	if err != nil { writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to toggle stream %s", req.Stream), err); return }
	writeOK(w)
}

// StartRecordingHandler handles POST /api/record/start
func StartRecordingHandler(w http.ResponseWriter, r *http.Request) {
	manager := gstreamer.GetManager(); if manager == nil { writeError(w, http.StatusServiceUnavailable, "GStreamer manager not initialized", nil); return }
	if err := manager.StartRecording(); err != nil { writeError(w, http.StatusInternalServerError, "Failed to start recording", err); return }
	writeOK(w)
}

// StopRecordingHandler handles POST /api/record/stop
func StopRecordingHandler(w http.ResponseWriter, r *http.Request) {
	manager := gstreamer.GetManager(); if manager == nil { writeError(w, http.StatusServiceUnavailable, "GStreamer manager not initialized", nil); return }
	if err := manager.StopRecording(); err != nil {
		// Log error but generally return OK as state is updated anyway
		log.Printf("Non-critical error stopping recording: %v", err)
	}
	writeOK(w)
}

// --- Recording Handlers using GORM (A6) ---

// GetRecordingsHandler handles GET /api/recordings
func GetRecordingsHandler(w http.ResponseWriter, r *http.Request) {
	recordings, err := db.GetRecordings()
	if err != nil { writeError(w, http.StatusInternalServerError, "Failed to retrieve recordings", err); return }
	writeJSONResponse(w, http.StatusOK, recordings)
}

// Helper to get ID from path using standard lib (A1 adaptation, requires Go 1.22+)
// Ensure your main.go setup uses http.ServeMux correctly for PathValue.
func getIDFromPathValue(r *http.Request) (uint, error) {
	idStr := r.PathValue("id")
	if idStr == "" { return 0, errors.New("missing recording ID in path pattern") }
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil { return 0, fmt.Errorf("invalid recording ID '%s' in path: %w", idStr, err) }
	return uint(idUint64), nil
}


// UpdateRecordingHandler handles PUT /api/recordings/{id}
func UpdateRecordingHandler(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPathValue(r)
	if err != nil { writeError(w, http.StatusBadRequest, err.Error(), err); return }

	var reqBody struct { Name string `json:"name"`; Notes *string `json:"notes"`; Genre *string `json:"genre"` }
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil { writeError(w, http.StatusBadRequest, "Invalid request body", err); return }
	defer r.Body.Close()

	if err := db.UpdateRecording(id, reqBody.Name, reqBody.Notes, reqBody.Genre); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { writeError(w, http.StatusNotFound, fmt.Sprintf("Recording %d not found", id), err) } else { writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update recording %d", id), err) }
		return
	}
	writeOK(w)
}

// ToggleProtectHandler handles POST /api/recordings/{id}/protect
func ToggleProtectHandler(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPathValue(r)
	if err != nil { writeError(w, http.StatusBadRequest, err.Error(), err); return }

	if err := db.ToggleProtect(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { writeError(w, http.StatusNotFound, fmt.Sprintf("Recording %d not found", id), err) } else { writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to toggle protection for recording %d", id), err) }
		return
	}
	writeOK(w)
}

// DeleteRecordingHandler handles DELETE /api/recordings/{id}
func DeleteRecordingHandler(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPathValue(r)
	if err != nil { writeError(w, http.StatusBadRequest, err.Error(), err); return }

	if err := db.DeleteRecording(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(err.Error(),"not found") { writeError(w, http.StatusNotFound, fmt.Sprintf("Recording %d not found", id), err) } else if strings.Contains(err.Error(), "protected") { writeError(w, http.StatusForbidden, err.Error(), err) } else { writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete recording %d", id), err) }
		return
	}
	writeOK(w)
}

