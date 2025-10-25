package api

import (
	"fmt" // Corrected: Added missing import
	"log"
	"nixon/internal/common"
	"nixon/internal/config"
	"nixon/internal/pipewire"
	"nixon/internal/websocket"
	"syscall"
	"time"
)

var (
	ctrl *control.ControlManager // This is likely needed but not in scope, assuming it's part of a larger context
)

// InitTasks starts background tasks for monitoring
// FIXED: InitTasks now accepts the config to pass to the audio manager
func InitTasks(cfg config.Config) {
	log.Println("Initializing background tasks...")
	// FIXED: Pass config to GetManager
	audioManager := pipewire.GetManager(cfg)

	// Start VAD monitoring
	go monitorVAD(audioManager)

	// Start disk usage monitoring
	go monitorDiskUsage(cfg)
}

// monitorVAD polls the audio manager for VAD status
func monitorVAD(audioManager *pipewire.AudioManager) {
	ticker := time.NewTicker(200 * time.Millisecond) // Poll rate
	defer ticker.Stop()

	for range ticker.C {
		status := audioManager.GetAudioStatus()
		// Broadcast VAD status
		websocket.GetHub().BroadcastUpdate(map[string]interface{}{
			"type":       "vad_status",
			"vadStatus":  status.VADStatus,
			"masterPeak": status.MasterPeak,
		})
	}
}

// monitorDiskUsage polls the recording directory for disk usage
func monitorDiskUsage(cfg config.Config) {
	ticker := time.NewTicker(30 * time.Second) // Poll rate
	defer ticker.Stop()

	// FIXED: Path corrected from cfg.AutoRecord.Directory to cfg.Recording.Directory
	path := cfg.Recording.Directory

	for {
		var stat syscall.Statfs_t
		err := syscall.Statfs(path, &stat)
		if err != nil {
			log.Printf("Error getting disk stats for %s: %v", path, err)
			// Wait before retrying
			<-ticker.C
			continue
		}

		// Calculate disk usage
		total := stat.Blocks * uint64(stat.Bsize)
		free := stat.Bfree * uint64(stat.Bsize)
		used := total - free
		percentUsed := (float64(used) / float64(total)) * 100

		diskUsage := common.DiskUsage{
			Total:       total,
			Free:        free,
			Used:        used,
			PercentUsed: percentUsed,
		}

		// Broadcast disk usage
		websocket.GetHub().BroadcastUpdate(map[string]interface{}{
			"type":      "disk_usage",
			"diskUsage": diskUsage,
		})

		<-ticker.C
	}
}

// monitorIcecastListeners polls the Icecast server for listener count
func monitorIcecastListeners(cfg config.Config) {
	if !cfg.IcecastSettings.IcecastEnabled {
		return
	}

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	// TODO: Implement Icecast XML status polling
	// http://user:pass@host:port/admin/stats
	// For now, broadcast a dummy count
	for range ticker.C {
		status := common.IcecastStatus{
			CurrentListeners: 0, // Dummy data
			PeakListeners:    0, // Dummy data
		}

		websocket.GetHub().BroadcastUpdate(map[string]interface{}{
			"type":          "icecast_status",
			"icecastStatus": status,
		})
	}
}

