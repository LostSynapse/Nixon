// internal/api/handlers.go
package api

import (
	"encoding/json"
	// "log" // Removed: imported and not used
	"net/http"
	"nixon/internal/config"
	"nixon/internal/gstreamer"
)

// HandleGetConfig retrieves the current configuration
func HandleGetConfig(w http.ResponseWriter, r *http.Request) {
	conf := config.GetConfig()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conf)
}

// HandleSetConfig updates and saves the configuration
func HandleSetConfig(w http.ResponseWriter, r *http.Request) {
	var newConfig config.Config
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update the global config
	config.SetGlobalConfig(&newConfig)

	// Save the config to disk
	// Corrected: Call the exported SaveConfig function
	if err := config.SaveConfig(&newConfig); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: Need to signal GStreamer manager to restart/reconfigure
	// For now, client will need to trigger a restart if audio settings changed.

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(newConfig)
}

// SetStreamStateRequest defines the body for the stream state endpoint
type SetStreamStateRequest struct {
	Recording bool `json:"is_recording"`
	SRT       bool `json:"is_streaming_srt"`
	Icecast   bool `json:"is_streaming_icecast"`
}

// HandleSetStreamState controls the dynamic GStreamer branches
func HandleSetStreamState(w http.ResponseWriter, r *http.Request) {
	var req SetStreamStateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Corrected: Call the exported GetManager function
	pm := gstreamer.GetManager()
	if pm == nil {
		http.Error(w, "PipelineManager not initialized", http.StatusInternalServerError)
		return
	}

	// Corrected: Call the appropriate Start/Stop/Toggle methods
	if req.Recording {
		go pm.StartRecording() // Use 'go' for non-blocking call
	} else {
		go pm.StopRecording()
	}

	go pm.ToggleSrtStream(req.SRT)
	go pm.ToggleIcecastStream(req.Icecast)

	w.WriteHeader(http.StatusOK)
}

// HandleGetStatus retrieves the current status from GStreamer
func HandleGetStatus(w http.ResponseWriter, r *http.Request) {
	// Corrected: Call the exported GetManager function
	pm := gstreamer.GetManager()
	if pm == nil {
		http.Error(w, "PipelineManager not initialized", http.StatusInternalServerError)
		return
	}

	// Corrected: Call the GetStatus() method which returns the full map
	status := pm.GetStatus()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// HandleGetAudioCapabilities retrieves hardware capabilities
func HandleGetAudioCapabilities(w http.ResponseWriter, r *http.Request) {
	device := r.URL.Query().Get("device")
	if device == "" {
		http.Error(w, "Missing 'device' query parameter", http.StatusBadRequest)
		return
	}

	caps, err := gstreamer.GetAudioCapabilities(device)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(caps)
}

