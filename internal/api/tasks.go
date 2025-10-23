// internal/api/tasks.go
package api

import (
	"nixon/internal/config"
	"nixon/internal/gstreamer"
	"nixon/internal/websocket"
	"time"
)

// AppStatus holds the current state of the application
type AppStatus struct {
	Config   *config.Config         `json:"config"`
	Gst      map[string]interface{} `json:"gst"`
	Clients  int                    `json:"clients"`
	Uptime   string                 `json:"uptime"` // This might be better calculated client-side
	ServerOS string                 `json:"server_os"`
}

var startTime = time.Now()

// GetFullStatus builds and returns the complete application status
func GetFullStatus() AppStatus {
	gstManager := gstreamer.GetManager()
	var gstStatus map[string]interface{}
	if gstManager != nil {
		gstStatus = gstManager.GetStatus()
	} else {
		gstStatus = make(map[string]interface{})
		gstStatus["error"] = "GStreamer manager not initialized"
	}

	return AppStatus{
		Config:   config.GetConfig(),
		Gst:      gstStatus,
		Clients:  websocket.GetClientCount(),
		Uptime:   time.Since(startTime).Round(time.Second).String(),
		ServerOS: "linux/arm64", // Placeholder, could be runtime detected
	}
}

// StartStatusUpdater periodically gets the status and broadcasts it
func StartStatusUpdater() {
	ticker := time.NewTicker(1 * time.Second) // Update every second
	defer ticker.Stop()

	for range ticker.C {
		status := GetFullStatus()
		websocket.BroadcastStatus(status)
	}
}

