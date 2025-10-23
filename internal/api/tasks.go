// internal/api/tasks.go
package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"nixon/internal/config"
	"nixon/internal/gstreamer" // A3: For updating state in manager
	"os"
	"strings"
	"time"

	"golang.org/x/sys/unix" // Keep for Statfs
)

// StartStatusUpdater initializes background tasks (Disk Usage, Icecast Listeners). (A3 adapted)
func StartStatusUpdater() {
	log.Println("Starting background tasks (Disk Usage, Icecast Listeners)...")
	go monitorDiskUsage()
	go monitorIcecastListeners()
	// No WebSocket polling started here anymore (A3)
}

// monitorDiskUsage periodically checks disk usage and updates the GStreamer manager state (A3).
func monitorDiskUsage() {
	// Initial update immediately
	updateDiskUsage()

	ticker := time.NewTicker(30 * time.Second) // Check less frequently?
	defer ticker.Stop()
	for range ticker.C {
		updateDiskUsage()
	}
}

// updateDiskUsage performs the disk usage check.
func updateDiskUsage() {
	var stat unix.Statfs_t
	// Check usage of the configured recordings directory for relevance
	cfg := config.GetConfig()
	targetDir := cfg.AutoRecord.Directory
	if targetDir == "" { targetDir = "." } // Fallback to CWD if empty

	if err := unix.Statfs(targetDir, &stat); err != nil {
		// Log error but maybe don't stop the monitor
		log.Printf("Error getting disk stats for '%s': %v", targetDir, err)
		// Optionally update state to an error indicator (-1?)
		// gstreamer.GetManager().SetDiskUsage(-1) // Example error state
		return
	}

	// Calculate usage (handle potential division by zero)
	totalBlocks := float64(stat.Blocks)
	freeBlocks := float64(stat.Bfree)
	blockSize := float64(stat.Bsize)
	percent := 0

	if totalBlocks > 0 && blockSize > 0 {
		totalBytes := totalBlocks * blockSize
		freeBytes := freeBlocks * blockSize
		usedBytes := totalBytes - freeBytes
		percent = int((usedBytes / totalBytes) * 100)
	} else {
		log.Printf("Warning: Invalid disk stats reported for '%s' (blocks=%d, bsize=%d)", targetDir, stat.Blocks, stat.Bsize)
		// Set error state?
		// gstreamer.GetManager().SetDiskUsage(-1)
		return
	}


	// A3: Update state within GStreamer Manager
	if manager := gstreamer.GetManager(); manager != nil {
		manager.SetDiskUsage(percent)
		// Logging is now handled within SetDiskUsage if changed
	}
}


// monitorIcecastListeners periodically checks listener counts (A3 adapted).
func monitorIcecastListeners() {
	// Structure definitions moved inside for locality
	type IcecastSourceStats struct { Listeners float64 `json:"listeners"` ListenerPeak float64 `json:"listener_peak"` PeakListeners float64 `json:"peak_listeners"` ListenURL string `json:"listenurl"` }
	type IcecastStats struct { Icestats struct { Source interface{} `json:"source"` } `json:"icestats"` }

	// Initial update immediately
	updateIcecastListeners()

	ticker := time.NewTicker(10 * time.Second) // Check frequency
	defer ticker.Stop()

	for range ticker.C {
		updateIcecastListeners()
	}
}

// updateIcecastListeners performs the listener count check.
func updateIcecastListeners() {
    // Structure definitions for parsing JSON
    type IcecastSourceStats struct { Listeners float64 `json:"listeners"` ListenerPeak float64 `json:"listener_peak"` PeakListeners float64 `json:"peak_listeners"` ListenURL string `json:"listenurl"` }
    type IcecastStats struct { Icestats struct { Source interface{} `json:"source"` } `json:"icestats"` }

	// Create HTTP client with timeout for robustness
	httpClient := &http.Client{Timeout: 5 * time.Second}

	// Get latest config and pipeline status safely
	cfg := config.GetConfig()
	manager := gstreamer.GetManager()
	if manager == nil { return } // Skip if manager not ready
	status := manager.GetStatus() // Get current status map

	// Extract Icecast status from the map
	isIcecastEnabled := cfg.IcecastSettings.IcecastEnabled // Check config directly
	isIcecastStreaming, _ := status["is_streaming_icecast"].(bool) // Check runtime state

	// If disabled in config or not currently streaming, set listeners to 0 and return
	if !isIcecastEnabled || !isIcecastStreaming {
		manager.SetListeners(0, -1) // Use -1 to indicate reset or keep existing peak? Let SetListeners decide.
		return
	}

	// Construct status URL
	host := cfg.IcecastSettings.IcecastHost
	port := cfg.IcecastSettings.IcecastPort
	if host == "" || port == 0 {
		log.Println("Warning: Icecast host/port not configured, cannot check listeners.")
		manager.SetListeners(0, -1) // Set to 0 if misconfigured
		return
	}
	statusURL := fmt.Sprintf("http://%s:%d/status-json.xsl", host, port)
	// TODO: Add config option for Icecast-KH /stats URL if needed

	resp, err := httpClient.Get(statusURL)
	if err != nil {
		log.Printf("Error fetching Icecast status from %s: %v", statusURL, err)
		// Don't reset listeners on temporary fetch error, keep last known state
		return
	}
	defer resp.Body.Close() // Ensure body is closed

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error fetching Icecast status: Received status %d from %s", resp.StatusCode, statusURL)
		// Don't reset listeners on temporary server error
		return
	}

	var stats IcecastStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		log.Printf("Error decoding Icecast status JSON from %s: %v", statusURL, err)
		return
	}

	// Process source(s) to find our mount point
	currentListeners := 0
	sourcePeak := 0 // Peak reported by the source for *this interval*
	foundSource := false
	// Ensure mount point starts with '/'
	mountWithSlash := cfg.IcecastSettings.IcecastMount
	if !strings.HasPrefix(mountWithSlash, "/") { mountWithSlash = "/" + mountWithSlash }

	processSource := func(sourceData map[string]interface{}) (int, int, bool) {
		listenURL, _ := sourceData["listenurl"].(string)
		// Check if URL ends with the mount point OR exactly matches if mount is just "/"
		if listenURL != "" && (strings.HasSuffix(listenURL, mountWithSlash) || mountWithSlash == "/") {
			listeners := 0.0; peak := 0.0
			if l, ok := sourceData["listeners"].(float64); ok { listeners = l }
			// Check both potential peak field names
			if p, ok := sourceData["listener_peak"].(float64); ok { peak = p }
			if pk, ok := sourceData["peak_listeners"].(float64); ok && peak == 0 { peak = pk }
			return int(listeners), int(peak), true
		}
		return 0, 0, false
	}

	switch src := stats.Icestats.Source.(type) {
	case map[string]interface{}:
		l, p, found := processSource(src)
		if found { currentListeners, sourcePeak, foundSource = l, p, true }
	case []interface{}:
		for _, item := range src {
			if sourceMap, ok := item.(map[string]interface{}); ok {
				l, p, found := processSource(sourceMap)
				if found { currentListeners, sourcePeak, foundSource = l, p, true; break }
			}
		}
	}

	if !foundSource {
		// Only log if we expect the stream to be running
		// log.Printf("Warning: Mount point '%s' not found in Icecast status from %s", mountWithSlash, statusURL)
		currentListeners = 0 // Assume 0 if mount not found while streaming
		sourcePeak = 0
	}

	// A3: Update state in GStreamer Manager
	manager.SetListeners(currentListeners, sourcePeak)
	// Logging is handled by SetListeners if changed
}

