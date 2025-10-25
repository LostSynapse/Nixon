package control

import (
	"log"
	"nixon/internal/common"
	"nixon/internal/config"
	"nixon/internal/db"
	"nixon/internal/pipewire"
	"sync"
	"time"
)

// ControlManager implements the ControlHandler interface.
// It acts as the intermediary between the API/plugins and the core services.
type ControlManager struct {
	conf         config.Config
	audioManager *pipewire.AudioManager
	managerLock  sync.RWMutex
}

var (
	manager *ControlManager
	once    sync.Once
)

// GetManager returns the singleton instance of the ControlManager.
func GetManager(conf config.Config, audioManager *pipewire.AudioManager) *ControlManager {
	once.Do(func() {
		manager = &ControlManager{
			conf:         conf,
			audioManager: audioManager,
		}
	})
	return manager
}

// --- Interface Implementation ---

// StartRecording tells the audio manager to start recording.
func (m *ControlManager) StartRecording() (string, error) {
	log.Println("Control: Received StartRecording command")
	return m.audioManager.StartRecording()
}

// StopRecording tells the audio manager to stop recording.
func (m *ControlManager) StopRecording() error {
	log.Println("Control: Received StopRecording command")
	return m.audioManager.StopRecording()
}

// GetAudioStatus retrieves the current status from the audio manager.
func (m *ControlManager) GetAudioStatus() common.AudioStatus {
	return m.audioManager.GetAudioStatus()
}

// ReloadConfig reloads configuration and propagates it to all managers.
func (m *ControlManager) ReloadConfig() {
	log.Println("Control: Received ReloadConfig command")
	// 1. Reload config from disk
	config.LoadConfig()
	// 2. Get the reloaded config
	m.managerLock.Lock()
	m.conf = config.GetConfig()
	m.managerLock.Unlock()

	// 3. Notify subsystems
	m.audioManager.ReloadConfig()
	log.Println("Control: Configuration reloaded and propagated to audio manager.")
}

// ListAudioDevices retrieves available audio devices.
func (m *ControlManager) ListAudioDevices() ([]common.AudioDevice, error) {
	return m.audioManager.ListAudioDevices()
}

// GetAudioCapabilities retrieves capabilities for a specific device.
func (mC *ControlManager) GetAudioCapabilities(deviceName string) (common.AudioCapabilities, error) {
	return mC.audioManager.GetAudioCapabilities(deviceName)
}

// --- Config Methods ---

// GetConfig returns the current in-memory configuration.
func (m *ControlManager) GetConfig() config.Config {
	m.managerLock.RLock()
	defer m.managerLock.RUnlock()
	return m.conf
}

// SaveConfig saves the configuration to disk.
func (m *ControlManager) SaveConfig(cfg config.Config) error {
	// Update in-memory config
	m.managerLock.Lock()
	m.conf = cfg
	m.managerLock.Unlock()

	// Save to disk
	// FIXED: Pass the config struct to the save function
	return config.SaveGlobalConfig(cfg)
}

// --- Stream Methods ---

// StartStream tells the audio manager to start a specific stream.
func (m *ControlManager) StartStream(streamName string) error {
	log.Printf("Control: Received StartStream command for: %s", streamName)
	return m.audioManager.StartStream(streamName)
}

// StopStream tells the audio manager to stop a specific stream.
func (m *ControlManager) StopStream(streamName string) error {
	log.Printf("Control: Received StopStream command for: %s", streamName)
	return m.audioManager.StopStream(streamName)
}

// --- Recording Management Methods ---

// GetAllRecordings retrieves all recordings from the database.
func (m *ControlManager) GetAllRecordings() ([]common.Recording, error) {
	return db.GetAllRecordings()
}

// GetRecordingByID retrieves a single recording by its ID.
func (m *ControlManager) GetRecordingByID(id uint) (*common.Recording, error) {
	return db.GetRecordingByID(id)
}

// UpdateRecording updates a recording's details in the database.
func (m *ControlManager) UpdateRecording(id uint, notes string, genre string, endTime time.Time, duration time.Duration) error {
	return db.UpdateRecording(id, notes, genre, endTime, duration)
}

// DeleteRecording deletes a recording from the database.
func (m *ControlManager) DeleteRecording(id uint) error {
	// TODO: Delete the actual file from disk
	/*
		rec, err := db.GetRecordingByID(id)
		if err == nil && rec != nil {
			// Delete file
		}
	*/

	// Delete from database
	return db.DeleteRecording(id)
}

