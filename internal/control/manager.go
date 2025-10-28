package control

import (
	"nixon/internal/common"
	"nixon/internal/config"
	"nixon/internal/slogger"
	"nixon/internal/websocket"
	"sync"
	"time"
)

var (
	managerInstance *Manager
	once            sync.Once
)

// Manager orchestrates the audio processing, recording, and streaming.
type Manager struct {
	// pipewireManager *pipewire.Manager // We will re-integrate this later.

	status    common.AudioStatus
	statusMux sync.RWMutex
}

// GetManager initializes and returns the singleton Manager instance.
func GetManager() (*Manager, error) {
	once.Do(func() {
		managerInstance = &Manager{
			status: common.AudioStatus{
				State: common.StateStopped,
			},
		}
	})
	return managerInstance, nil
}

// GetStatus returns the current audio status in a thread-safe way.
func (m *Manager) GetStatus() common.AudioStatus {
	m.statusMux.RLock()
	defer m.statusMux.RUnlock()
	return m.status
}

// setStatus updates the current audio status and broadcasts it.
func (m *Manager) setStatus(newStatus common.AudioStatus) {
	m.statusMux.Lock()
	m.status = newStatus
	m.statusMux.Unlock()

	// Broadcast the new status to all connected WebSocket clients.
	websocket.BroadcastStatus(newStatus)
}

// StartAudio begins the main audio processing loop (VAD, etc.).
func (m *Manager) StartAudio() error {
	slogger.Log.Info("Control Manager: Starting audio processing...")

	// In a real implementation, this would start the GStreamer/Pipewire pipeline.
	// For now, we will just simulate the state change.

	go m.simulateVAD()
	return nil
}

// StopAudio stops the main audio processing loop.
func (m *Manager) StopAudio() error {
	slogger.Log.Info("Control Manager: Stopping audio processing.")
	// TODO: Add logic to stop the simulateVAD goroutine.
	return nil
}

// simulateVAD is a placeholder for the real VAD logic from GStreamer.
// It will periodically toggle the recording state to test the UI.
func (m *Manager) simulateVAD() {
	slogger.Log.Info("Starting VAD simulation...")
	isRecording := false
	ticker := time.NewTicker(15 * time.Second) // Toggle every 15 seconds
	defer ticker.Stop()

	for range ticker.C {
		if isRecording {
			slogger.Log.Debug("VAD SIM: Silence detected, stopping recording.")
			m.StopRecording()
			isRecording = false
		} else {
			slogger.Log.Debug("VAD SIM: Voice detected, starting recording.")
			m.StartRecording()
			isRecording = true
		}
	}
}

// StartRecording starts a new recording.
func (m *Manager) StartRecording() error {
	slogger.Log.Info("Control Manager: Starting recording...")
	cfg := config.AppConfig

	// Here, you would interact with the database to create a new recording entry.
	// For now, we just update the state.

	m.setStatus(common.AudioStatus{
		State:     common.StateRecording,
		IsAutoRec: cfg.AutoRec.Enabled, // Reflect config
	})
	return nil
}

// StopRecording stops the current recording.
func (m *Manager) StopRecording() error {
	slogger.Log.Info("Control Manager: Stopping recording...")
	m.setStatus(common.AudioStatus{
		State: common.StateStopped,
	})
	return nil
}

// --- Placeholder methods to satisfy the interface ---

func (m *Manager) StartStream(streamType string) error {
	slogger.Log.Info("Starting stream", "stream_type", streamType)
	return nil
}

func (m *Manager) StopStream(streamType string) error {
	slogger.Log.Info("Stopping stream", "stream_type", streamType)
	return nil
}

func (m *Manager) GetAudioDevices() ([]common.AudioDevice, error) {
	// Return dummy data for now
	return []common.AudioDevice{
		{DeviceName: "default", Description: "Default System Device"},
	}, nil
}
func (m *Manager) GetRecordings() ([]common.Recording, error) {
	slogger.Log.Debug("Manager: GetRecordings called (placeholder)")
	// In a real implementation, this would query the database.
	return []common.Recording{}, nil
}

func (m *Manager) DeleteRecording(id uint) error {
	slogger.Log.Debug("Manager: DeleteRecording called (placeholder)", "id", id)
	// In a real implementation, this would delete from the database.
	return nil
}
