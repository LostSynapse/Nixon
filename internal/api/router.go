package api

import (
	"encoding/json"
	"net/http"
	"nixon/internal/common"
	"nixon/internal/config"
	"nixon/internal/control"
	"nixon/internal/pipewire"
	"nixon/internal/websocket" // FIXED: Added missing import
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"
)

// Router struct holds dependencies
type Router struct {
	ctrl *control.ControlManager
	hub  *websocket.Hub
}

// NewRouter creates a new router with dependencies
func NewRouter(ctrl *control.ControlManager, hub *websocket.Hub) *mux.Router {
	r := &Router{
		ctrl: ctrl,
		hub:  hub,
	}

	router := mux.NewRouter()

	// WebSocket handler
	// FIXED: wsHandler now defined and uses the hub
	router.HandleFunc("/ws", r.wsHandler)

	// API routes
	api := router.PathPrefix("/api").Subrouter()

	// Config
	api.HandleFunc("/config", r.getConfigHandler).Methods("GET")
	api.HandleFunc("/config", r.saveConfigHandler).Methods("POST")
	api.HandleFunc("/config/reload", r.reloadConfigHandler).Methods("POST")

	// Audio Devices
	api.HandleFunc("/devices", r.listAudioDevicesHandler).Methods("GET")
	api.HandleFunc("/devices/{name}/capabilities", r.getAudioCapabilitiesHandler).Methods("GET")

	// Status
	api.HandleFunc("/status", r.getStatusHandler).Methods("GET")

	// Recording
	api.HandleFunc("/recordings", r.startRecordingHandler).Methods("POST")
	api.HandleFunc("/recordings", r.stopRecordingHandler).Methods("DELETE")
	api.HandleFunc("/recordings", r.listRecordingsHandler).Methods("GET")
	api.HandleFunc("/recordings/{id}", r.getRecordingHandler).Methods("GET")
	api.HandleFunc("/recordings/{id}", r.updateRecordingHandler).Methods("PUT")
	api.HandleFunc("/recordings/{id}", r.deleteRecordingHandler).Methods("DELETE")

	// Streaming
	api.HandleFunc("/streams/srt", r.startSrtStreamHandler).Methods("POST")
	api.HandleFunc("/streams/srt", r.stopSrtStreamHandler).Methods("DELETE")
	api.HandleFunc("/streams/icecast", r.startIcecastStreamHandler).Methods("POST")
	api.HandleFunc("/streams/icecast", r.stopIcecastStreamHandler).Methods("DELETE")

	return router
}

// wsHandler handles websocket connections
// FIXED: This handler now uses the hub passed to the Router
func (rt *Router) wsHandler(w http.ResponseWriter, r *http.Request) {
	// GetHub() is internal, ServeWs is the handler function
	websocket.ServeWs(rt.hub, w, r)
}

// getConfigHandler returns the current configuration
func (rt *Router) getConfigHandler(w http.ResponseWriter, r *http.Request) {
	cfg := rt.ctrl.GetConfig()
	json.NewEncoder(w).Encode(cfg)
}

// saveConfigHandler saves the configuration
func (rt *Router) saveConfigHandler(w http.ResponseWriter, r *http.Request) {
	var cfg config.Config
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := rt.ctrl.SaveConfig(cfg); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast config update
	// FIXED: Broadcast requires a map
	cfgMap := configStructToMap(cfg)
	rt.hub.BroadcastUpdate(cfgMap)

	w.WriteHeader(http.StatusOK)
}

// reloadConfigHandler reloads the configuration from disk
func (rt *Router) reloadConfigHandler(w http.ResponseWriter, r *http.Request) {
	// FIXED: ReloadConfig returns an error, not the config
	if err := rt.ctrl.ReloadConfig(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the newly loaded config
	cfg := rt.ctrl.GetConfig()

	// Broadcast config update
	// FIXED: Broadcast requires a map
	cfgMap := configStructToMap(cfg)
	rt.hub.BroadcastUpdate(cfgMap)

	w.WriteHeader(http.StatusOK)
}

// listAudioDevicesHandler returns available audio devices
func (rt *Router) listAudioDevicesHandler(w http.ResponseWriter, r *http.Request) {
	devices, err := rt.ctrl.ListAudioDevices()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(devices)
}

// getAudioCapabilitiesHandler returns device capabilities
func (rt *Router) getAudioCapabilitiesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	caps, err := rt.ctrl.GetAudioCapabilities(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(caps)
}

// getStatusHandler returns the current audio status
func (rt *Router) getStatusHandler(w http.ResponseWriter, r *http.Request) {
	status := rt.ctrl.GetAudioStatus()
	json.NewEncoder(w).Encode(status)
}

// startRecordingHandler starts a new recording
func (rt *Router) startRecordingHandler(w http.ResponseWriter, r *http.Request) {
	rec, err := rt.ctrl.StartRecording()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(rec)
}

// stopRecordingHandler stops the current recording
func (rt *Router) stopRecordingHandler(w http.ResponseWriter, r *http.Request) {
	// FIXED: StopRecording returns one value (error)
	if err := rt.ctrl.StopRecording(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// listRecordingsHandler returns all recordings
func (rt *Router) listRecordingsHandler(w http.ResponseWriter, r *http.Request) {
	recordings, err := rt.ctrl.GetAllRecordings()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(recordings)
}

// getRecordingHandler returns a single recording
func (rt *Router) getRecordingHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// FIXED: Convert string ID to uint
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid recording ID", http.StatusBadRequest)
		return
	}

	rec, err := rt.ctrl.GetRecordingByID(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(rec)
}

// updateRecordingHandler updates a recording's metadata
func (rt *Router) updateRecordingHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// FIXED: Convert string ID to uint
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid recording ID", http.StatusBadRequest)
		return
	}

	var reqBody struct {
		Notes string `json:"notes"`
		Genre string `json:"genre"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// FIXED: UpdateRecording signature
	if err := rt.ctrl.UpdateRecording(uint(id), reqBody.Notes, reqBody.Genre); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// deleteRecordingHandler deletes a recording
func (rt *Router) deleteRecordingHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// FIXED: Convert string ID to uint
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid recording ID", http.StatusBadRequest)
		return
	}

	if err := rt.ctrl.DeleteRecording(uint(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// startSrtStreamHandler starts the SRT stream
func (rt *Router) startSrtStreamHandler(w http.ResponseWriter, r *http.Request) {
	if err := rt.ctrl.StartStream("srt"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// stopSrtStreamHandler stops the SRT stream
func (rt *Router) stopSrtStreamHandler(w http.ResponseWriter, r *http.Request) {
	if err := rt.ctrl.StopStream("srt"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// startIcecastStreamHandler starts the Icecast stream
func (rt *Router) startIcecastStreamHandler(w http.ResponseWriter, r *http.Request) {
	if err := rt.ctrl.StartStream("icecast"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// stopIcecastStreamHandler stops the Icecast stream
func (rt *Router) stopIcecastStreamHandler(w http.ResponseWriter, r *http.Request) {
	if err := rt.ctrl.StopStream("icecast"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// configStructToMap converts config struct to map for websocket broadcast
func configStructToMap(cfg config.Config) map[string]interface{} {
	// This is a robust way to convert a struct to the map required by websocket
	var cfgMap map[string]interface{}
	data, _ := json.Marshal(cfg)
	json.Unmarshal(data, &cfgMap)
	
	// Add message type for client-side routing
	cfgMap["type"] = "config_update"
	return cfgMap
}

