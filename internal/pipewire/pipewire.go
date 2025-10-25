package pipewire

import (
	"fmt"
	"log"
	"nixon/internal/common"
	"nixon/internal/config"
	"nixon/internal/db"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AudioManager handles all audio processing via PipeWire
type AudioManager struct {
	conf       config.Config
	status     common.AudioStatus
	statusLock sync.RWMutex
	currentRec *common.Recording
}

var (
	manager *AudioManager
	once    sync.Once
)

// GetManager returns the singleton instance of the AudioManager
func GetManager(conf config.Config) *AudioManager {
	once.Do(func() {
		manager = &AudioManager{
			conf: conf,
			status: common.AudioStatus{
				State:          common.StateStopped,
				ActiveStreams:  make(map[string]bool),
				VADStatus:      false,
				MasterPeak:     -100.0,
				LastVADEvent:   time.Now(),
				IsAutoRec:      false,
				CurrentRecFile: "",
			},
		}
		// TODO: Init PipeWire connection, start VAD monitor, etc.
		go manager.connectPipeWire()
	})
	return manager
}

// connectPipeWire initializes the connection and monitoring.
// (This is a stub for the real PipeWire DBus/client logic)
func (m *AudioManager) connectPipeWire() {
	log.Println("PipeWire Manager: Initializing...")
	// STUB: This is where you would connect to PipeWire's DBus API
	// and set up listeners for VAD (e.g., monitoring a Level node).
	// For simulation, we'll just log.
	log.Println("PipeWire Manager: Connection stub initialized.")

	// Example VAD monitoring loop (simulation)
	/*
		go func() {
			for {
				time.Sleep(100 * time.Millisecond)
				// STUB: m.updateVADStatus(peak)
			}
		}()
	*/
}

// ReloadConfig updates the audio manager's config.
func (m *AudioManager) ReloadConfig() {
	log.Println("PipeWire Manager: Reloading configuration...")
	m.statusLock.Lock()
	m.conf = config.GetConfig()
	m.statusLock.Unlock()
	// TODO: Update PipeWire nodes, VAD settings, etc.
}

// GetAudioStatus returns the current audio status.
func (m *AudioManager) GetAudioStatus() common.AudioStatus {
	m.statusLock.RLock()
	defer m.statusLock.RUnlock()
	// Create a copy to avoid race conditions on the map
	statusCopy := m.status
	statusCopy.ActiveStreams = make(map[string]bool)
	for k, v := range m.status.ActiveStreams {
		statusCopy.ActiveStreams[k] = v
	}
	return statusCopy
}

// StartRecording starts a manual recording.
func (m *AudioManager) StartRecording() (string, error) {
	m.statusLock.Lock()
	defer m.statusLock.Unlock()

	if m.status.State == common.StateRecording {
		return "", fmt.Errorf("recording is already in progress")
	}

	m.status.State = common.StateRecording
	m.status.IsAutoRec = false

	filename := m.generateFilename()
	m.status.CurrentRecFile = filename
	log.Printf("PipeWire Manager: Starting manual recording: %s", filename)

	// --- Database Entry ---
	rec, err := db.AddRecording(filename, time.Now())
	if err != nil {
		log.Printf("Error adding recording to DB: %v", err)
		// Don't stop the recording, but log the error
	}
	m.currentRec = rec
	// ---

	// TODO: Tell PipeWire Session Manager (WirePlumber) to start
	// a node chain to record to this file.
	// e.g., dbus.Call("org.pipewire.WirePlumber", "/CreateRecording", ...)

	// Broadcast update (via callback or internal reference)
	// This is handled by api.InitTasks which polls GetAudioStatus

	return filename, nil
}

// StopRecording stops a manual recording.
func (m *AudioManager) StopRecording() error {
	m.statusLock.Lock()
	defer m.statusLock.Unlock()

	if m.status.State == common.StateStopped {
		return fmt.Errorf("no recording is in progress")
	}

	log.Printf("PipeWire Manager: Stopping recording: %s", m.status.CurrentRecFile)
	m.status.State = common.StateStopped
	m.status.CurrentRecFile = ""

	// --- Finalize DB Entry ---
	if m.currentRec != nil {
		go m.finalizeRecording(m.currentRec) // Run in goroutine to not block
		m.currentRec = nil
	}
	// ---

	// TODO: Tell PipeWire Session Manager (WirePlumber) to stop
	// the recording node chain.

	return nil
}

// --- NEWLY ADDED METHODS ---

// StartStream starts a specific stream.
// (This is a stub)
func (m *AudioManager) StartStream(streamName string) error {
	m.statusLock.Lock()
	defer m.statusLock.Unlock()

	log.Printf("PipeWire Manager: Starting stream '%s'", streamName)
	// TODO: Logic to find the correct stream config (SRT/Icecast)
	// and load/configure the corresponding PipeWire module via DBus.

	m.status.ActiveStreams[streamName] = true

	// Broadcast update (handled by polling)
	return nil
}

// StopStream stops a specific stream.
// (This is a stub)
func (m *AudioManager) StopStream(streamName string) error {
	m.statusLock.Lock()
	defer m.statusLock.Unlock()

	log.Printf("PipeWire Manager: Stopping stream '%s'", streamName)
	// TODO: Logic to unload the PipeWire module via DBus.

	delete(m.status.ActiveStreams, streamName)

	// Broadcast update (handled by polling)
	return nil
}

// --- Hardware Stubs ---

// ListAudioDevices returns a stub list of devices.
func (m *AudioManager) ListAudioDevices() ([]common.AudioDevice, error) {
	// STUB: This should query PipeWire/WirePlumber via DBus
	return []common.AudioDevice{
		{DeviceName: "default-source", Driver: "PipeWire", Description: "Default Source"},
		{DeviceName: "default-sink", Driver: "PipeWire", Description: "Default Sink"},
	}, nil
}

// GetAudioCapabilities returns a stub list of capabilities.
func (m *AudioManager) GetAudioCapabilities(deviceName string) (common.AudioCapabilities, error) {
	// STUB: This should query the specific device node via DBus
	return common.AudioCapabilities{
		Formats:     []string{"S16LE", "S24LE", "S32LE", "F32LE"},
		SampleRates: []int{44100, 48000, 96000},
		Channels:    []int{1, 2, 8, 16},
	}, nil
}

// --- Internal Helpers ---

// finalizeRecording updates the DB entry with end time and duration.
func (m *AudioManager) finalizeRecording(rec *common.Recording) {
	log.Printf("Finalizing recording ID %d in DB...", rec.ID)
	endTime := time.Now()
	duration := endTime.Sub(rec.StartTime)

	// TODO: Get actual file path and calculate size
	// rec.FileSize = ...

	err := db.UpdateRecording(rec.ID, rec.Notes, rec.Genre, endTime, duration)
	if err != nil {
		log.Printf("Error finalizing recording %d in DB: %v", rec.ID, err)
	}
	log.Printf("Finalized recording ID %d.", rec.ID)
}

// generateFilename creates a new filename based on the config pattern.
func (m *AudioManager) generateFilename() string {
	// Use directory from Recording settings
	dir := m.conf.Recording.Directory
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("Warning: Could not create recording directory '%s': %v", dir, err)
		dir = "." // Fallback to current directory
	}

	// STUB: This needs a real pattern replacer (YYYY, MM, etc.)
	// For now, use a simple timestamp.
	pattern := m.conf.Recording.FilePattern
	format := m.conf.Recording.FileFormat
	if format == "" {
		format = "flac" // Default
	}

	// Basic replacement
	filename := fmt.Sprintf("rec_%d.%s", time.Now().Unix(), format)
	log.Printf("Using pattern: %s (STUBBED)", pattern)

	return filepath.Join(dir, filename)
}

