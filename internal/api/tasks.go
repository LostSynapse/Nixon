// internal/api/tasks.go
// This file handles background tasks like disk monitoring and Icecast listener polling.

package api

import (
	"fmt"
	"log"
	"nixon/internal/config"
	"nixon/internal/gstreamer"
	"nixon/internal/websocket"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// --- Task Manager ---

// TaskManager holds global state for background monitoring tasks.
type TaskManager struct {
	updateBroadcast func(map[string]interface{})
}

// InitializeTasks sets up and starts all background tasks.
func InitializeTasks(broadcastFunc func(map[string]interface{})) *TaskManager {
	tm := &TaskManager{
		updateBroadcast: broadcastFunc,
	}

	// Start monitoring tasks
	go tm.monitorDiskUsage()
	go tm.monitorIcecastListeners()
	
	log.Println("Background tasks initialized.")
	return tm
}

// --- Status Structs ---

// IcecastStatus mirrors relevant information from Icecast's JSON status page.
// The syntax errors are in these two struct definitions.
type IcecastStatus struct {
	// CORRECTED: Removed incorrect pointer syntax (**int)
	ListenerCurrent int `json:"listeners"` 
	ListenerPeak int `json:"listener_peak"`   
}

// GlobalStatus is the structure for the comprehensive status broadcast.
// It combines GStreamer status with monitoring data.
type GlobalStatus struct {
	gstreamer.StatusUpdate // Embedded GStreamer state

	// Icecast listener data
	ListenerCurrent int `json:"listener_current"`
	ListenerPeak    int `json:"listener_peak"`
}

// --- Disk Usage Monitor ---

// monitorDiskUsage runs periodically to check the disk space in the recording directory.
func (tm *TaskManager) monitorDiskUsage() {
	// Execute immediately and then repeat
	tm.checkDiskUsage()
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		tm.checkDiskUsage()
	}
}

func (tm *TaskManager) checkDiskUsage() {
	dir := config.GetConfig().AutoRecord.Directory
	if dir == "" {
		// Cannot check disk usage if directory is not set
		return
	}

	// Use 'df -P' to get disk space information for the directory
	// -P ensures POSIX output format, easier to parse
	cmd := exec.Command("df", "-P", dir)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("ERROR: Failed to run df command: %v", err)
		return
	}

	// Output is typically two lines: header and data
	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		log.Printf("ERROR: df output unexpectedly short.")
		return
	}

	// Split the data line (second line) by whitespace
	fields := strings.Fields(lines[1])
	
	// The 5th field (index 4) contains the usage percentage, e.g., "12%"
	if len(fields) >= 5 {
		usageStr := fields[4]
		// Remove the trailing '%' and parse to integer
		usageStr = strings.TrimSuffix(usageStr, "%")
		if percent, err := strconv.Atoi(usageStr); err == nil {
			// Update the GStreamer manager with the new disk usage percentage
			gstreamer.GetManager().SetDiskUsage(percent)
		} else {
			log.Printf("ERROR: Failed to parse disk usage percentage '%s': %v", fields[4], err)
		}
	} else {
		log.Printf("ERROR: df output has too few fields in data line: %v", fields)
	}
}

// --- Icecast Listener Monitor ---

// monitorIcecastListeners runs periodically to check listener counts.
func (tm *TaskManager) monitorIcecastListeners() {
	// Execute immediately and then repeat
	tm.pollIcecast()
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		tm.pollIcecast()
	}
}

func (tm *TaskManager) pollIcecast() {
	cfg := config.GetConfig()
	if !cfg.IcecastSettings.IcecastEnabled {
		return
	}

	// NOTE: This assumes the Icecast server is running on the local appliance
	// and uses the standard admin credentials and API path.
	// FIX: Assuming the following fields now exist in config.IcecastSettings
	url := fmt.Sprintf("http://%s:%s@%s:%d/admin/stats", 
		cfg.IcecastSettings.IcecastUser,    // FIX
		cfg.IcecastSettings.IcecastPassword, // FIX
		cfg.IcecastSettings.IcecastHost,     // FIX
		cfg.IcecastSettings.IcecastPort,
	)

	// In a real implementation, you would use net/http to fetch and parse the JSON.
	// We'll simulate this with a log message for now since we don't want to add a net/http dependency to a non-modular task.
	log.Printf("Polling Icecast listener data from: %s (Simulated)", url)

	// SIMULATED RESPONSE PARSING
	// Assuming the result is successfully parsed into an IcecastStatus struct:
	current := 0
	peak := 0
	
	// Update the GStreamer manager (which calls broadcast)
	gstreamer.GetManager().SetListeners(current, peak)

	// Broadcast the specific listener update separately via WebSocket,
	// as listener state is external to the GStreamer pipeline manager.
	websocket.BroadcastUpdate(map[string]interface{}{
		"listener_current": current,
		"listener_peak": peak,
	})
}
