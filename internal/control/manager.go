package control

import (
	"fmt"
	"nixon/internal/slogger"
	"nixon/internal/common"
	"nixon/internal/config"
	"nixon/internal/pipewire"
	"nixon/internal/websocket"
	"sync"
	"time"
)

var (
	manager      *Manager
	managerMutex = &sync.Mutex{}
)

// Manager orchestrates the streaming and recording processes.
type Manager struct {
	pwManager *pipewire.Manager
	// More fields would be added here, e.g., for managing recordings, stream outputs, etc.
}

// GetManager initializes and returns the singleton Manager instance.
func GetManager() (*Manager, error) {
	managerMutex.Lock()
	defer managerMutex.Unlock()

	if manager == nil {
		cfg := config.AppConfig // This line was changed from config.GetConfig()
		pwManager, err := pipewire.NewManager(cfg.Pipewire.Socket)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Pipewire manager: %w", err)
		}

		manager = &Manager{
			pwManager: pwManager,
		}
	}

	return manager, nil
}

// StartBackgroundTasks starts long-running processes for the application.
func (m *Manager) StartBackgroundTasks() {
	slogger.Log.Info("Starting control manager background tasks...")
	go m.monitorVAD()
}

// monitorVAD checks for voice activity and broadcasts updates.
func (m *Manager) monitorVAD() {
	slogger.Log.Info("Starting VAD monitoring...")
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// In a real implementation, we would check VAD status here.
		// For now, we'll simulate broadcasting a status update.
		websocket.Broadcast("{\"type\": \"vad_status\", \"active\": false}")
	}
}

// StartStream starts a stream.
func (m *Manager) StartStream(streamType string) error {
	// Logic to start a stream, e.g., SRT or WebRTC
	slogger.Log.Info("Starting stream", "stream_type", streamType)
	return nil
}

// StopStream stops a stream.
func (m *Manager) StopStream(streamType string) error {
	slogger.Log.Info("Stopping stream", "stream_type", streamType)
	return nil
}

// StartRecording starts a recording.
func (m *Manager) StartRecording() (uint, error) {
	// Logic to start a recording
	slogger.Log.Info("Starting recording")

	// Placeholder: Return a dummy recording ID and no error.
	// In a real implementation, this would interact with the database.
	return 1, nil
}

// StopRecording stops the current recording.
func (m *Manager) StopRecording() error {
	slogger.Log.Info("Stopping recording")
	return nil
}

// GetAudioDevices lists available audio devices.
func (m *Manager) GetAudioDevices() ([]string, error) {
	// This would interact with the audio backend (PipeWire/GStreamer)
	return []string{"Default", "Microphone 1", "Line In"}, nil
}

// GetRecordings retrieves all recordings.
func (m *Manager) GetRecordings() ([]common.Recording, error) {
	// Placeholder implementation to satisfy the router.
	slogger.Log.Debug("Manager: Getting recordings")
	return []common.Recording{}, nil
}

// DeleteRecording removes a recording.
func (m *Manager) DeleteRecording(id uint) error {
	// Placeholder implementation to satisfy the router.
	slogger.Log.Info("Manager: Deleting recording", "recording_id", id)
	return nil
}