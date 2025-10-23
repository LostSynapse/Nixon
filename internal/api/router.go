// internal/api/router.go
package api

import (
	"net/http"
	"nixon/internal/config"
	"nixon/internal/websocket"
	"path/filepath"

	"github.com/gorilla/mux"
)

// NewRouter creates and configures the main application router
func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	// API routes (must be defined BEFORE the static handlers)
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/config", GetConfig).Methods("GET")
	api.HandleFunc("/config", UpdateConfig).Methods("POST")
	api.HandleFunc("/capabilities", GetAudioCapabilitiesHandler).Methods("GET")
	api.HandleFunc("/status", GetStreamStatus).Methods("GET")
	api.HandleFunc("/stream", SetStreamState).Methods("POST")
	api.HandleFunc("/record/start", StartRecordingHandler).Methods("POST")
	api.HandleFunc("/record/stop", StopRecordingHandler).Methods("POST")

	// WebSocket route
	router.HandleFunc("/ws", websocket.HandleConnections)

	// --- Static File Server (Replicating old Gin logic) ---

	// 1. Serve the /assets directory
	// This handles /assets/index-DtCWpx-N.js, etc.
	assetDir := "./web/assets/"
	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir(assetDir))))

	// 2. Serve the /recordings directory
	// Note: We get the directory from the config, which is loaded *before* the router is created.
	recDir := config.GetConfig().AutoRecord.Directory
	router.PathPrefix("/recordings/").Handler(http.StripPrefix("/recordings/", http.FileServer(http.Dir(recDir))))

	// 3. Serve specific files from the /web root
	// This handles the favicon
	router.Path("/nixon_logo.svg").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/nixon_logo.svg")
	})

	// 4. "NoRoute" catch-all: Serve index.html for all other routes
	// This is the most critical part for React's client-side router.
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("./web", "index.html"))
	})

	return router
}

