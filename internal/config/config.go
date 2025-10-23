// internal/config/config.go
package config

import (
	"bytes" // Added import for bytes.NewReader
	"encoding/json"
	"fmt"
	"io" // Import io for Reader interface used by atomic
	"log"
	"os"
	"sync"

	"github.com/natefinch/atomic" // Using natefinch/atomic
)

const (
	configFile = "./config.json"
	// RecordingsDir = "./recordings" // Defined in struct now
)

// --- Struct Definitions ---
// (AudioSettings, IcecastSettings, SrtSettings, AutoRecordSettings, NetworkSettings remain the same)

type AudioSettings struct {
	Device         string  `json:"device"`
	SampleRate     int     `json:"sample_rate"`
	BitDepth       int     `json:"bit_depth"`
	MasterChannels []int   `json:"master_channels"` // For selecting channels from multichannel device
	Bitrate        int     `json:"bitrate"`         // Kept for compatibility/legacy? Review if needed. GStreamer uses specific encoder bitrates now.
	Channels       int     `json:"channels"`        // Kept for compatibility/legacy? Review if needed. GStreamer uses master_channels length now.
}

type IcecastSettings struct {
	IcecastEnabled    bool   `json:"icecast_enabled"`
	IcecastHost       string `json:"icecast_host"`
	IcecastPort       int    `json:"icecast_port"`
	IcecastMount      string `json:"icecast_mount"`
	IcecastPassword   string `json:"icecast_password"`
	IcecastBitrate    int    `json:"icecast_bitrate"`    // Specific bitrate for Icecast encoder
	StreamName        string `json:"stream_name"`        // Kept for compatibility
	StreamGenre       string `json:"stream_genre"`       // Kept for compatibility
	StreamDescription string `json:"stream_description"` // Kept for compatibility
	ServerType        string `json:"server_type"`        // Kept for compatibility
	URL               string `json:"url"`                // Kept for compatibility? Consolidate with Host?
}

type SrtSettings struct {
	SrtEnabled bool   `json:"srt_enabled"`
	SrtHost    string `json:"srt_host"` // Destination host for SRT
	SrtPort    int    `json:"srt_port"` // Destination port for SRT
	SrtBitrate int    `json:"srt_bitrate"` // Specific bitrate for SRT encoder
}

type AutoRecordSettings struct {
	Enabled           bool    `json:"enabled"`
	Directory         string  `json:"directory"`          // Renamed from RecordingsDir constant
	PrerollDuration   int     `json:"preroll_duration"`   // In seconds
	VadThreshold      float64 `json:"vad_threshold"`      // Original gvad threshold (unused now)
	VadDbThreshold    float64 `json:"vad_db_threshold"`   // New dB threshold for level element
	SmartSplitEnabled bool    `json:"smart_split_enabled"`
	SmartSplitTimeout int     `json:"smart_split_timeout"` // In seconds
	TimeoutSeconds    int     `json:"timeout_seconds"`     // Kept for compatibility/legacy? Review if needed. SmartSplitTimeout is used now.
}

type NetworkSettings struct {
	SignalingURL string `json:"signaling_url"`
	StunURL      string `json:"stun_url"`
}

// Config struct holds the application configuration
type Config struct {
	AudioSettings    AudioSettings    `json:"audio_settings"`
	IcecastSettings  IcecastSettings  `json:"icecast_settings"`
	SrtSettings      SrtSettings      `json:"srt_settings"`
	AutoRecord       AutoRecordSettings `json:"auto_record"`
	NetworkSettings  NetworkSettings  `json:"network_settings"`
	SRTEnabled     bool             `json:"srt_enabled"`      // Kept for compatibility? Consolidate into SrtSettings.
	IcecastEnabled bool             `json:"icecast_enabled"`  // Kept for compatibility? Consolidate into IcecastSettings.
}

// --- Global Variable & Mutex (Unexported) ---
var (
	globalConfig *Config
	configMutex  sync.RWMutex
	initOnce     sync.Once
)

// --- Initialization ---

// Initialize loads the config or creates a default one. MUST be called once at startup.
func Initialize() error {
	var initErr error
	initOnce.Do(func() {
		// 1. Set default config in memory immediately
		applyDefaults()

		// 2. Attempt to read config file
		data, err := os.ReadFile(configFile)
		if err != nil {
			if os.IsNotExist(err) {
				log.Println("config.json not found, creating with default values.")
				// Defaults already applied, just save them
				if saveErr := saveConfigUnlocked(); saveErr != nil {
					initErr = fmt.Errorf("failed to save initial default config: %w", saveErr)
				} else {
					log.Println("Default configuration saved to config.json")
				}
				return // Success or save error recorded
			}
			// Other read error (permissions, etc.)
			initErr = fmt.Errorf("failed to read config file '%s': %w", configFile, err)
			log.Printf("ERROR: %v. Running with default settings.", initErr) // Log error but continue with defaults
			return
		}

		// 3. Attempt to unmarshal read data
		configMutex.Lock() // Lock needed ONLY for unmarshalling into globalConfig
		defer configMutex.Unlock()
		if err := json.Unmarshal(data, globalConfig); err != nil {
			initErr = fmt.Errorf("failed to decode config file '%s': %w", configFile, err)
			log.Printf("ERROR: %v. Running with default settings.", initErr) // Log error but continue with defaults
			// Re-apply defaults to ensure a valid state if unmarshal failed partially
			applyDefaultsUnlocked()
			return
		}

		// 4. Validate and apply defaults for potentially missing fields after loading
		if applyDefaultsUnlocked() { // Returns true if changes were made
			log.Println("Applying default values for missing fields in config.json and saving.")
			if saveErr := saveConfigUnlocked(); saveErr != nil {
				// Log error but don't overwrite the primary unmarshal error if one occurred
				log.Printf("Warning: failed to save config after applying defaults: %v", saveErr)
				if initErr == nil {
					initErr = fmt.Errorf("failed to save config after applying defaults: %w", saveErr)
				}
			}
		}

		log.Printf("Configuration loaded from %s", configFile)
	})
	return initErr // Return error encountered during initOnce.Do
}

// applyDefaults sets the default configuration values. MUST be called within initOnce.Do.
func applyDefaults() {
	configMutex.Lock()
	defer configMutex.Unlock()
	applyDefaultsUnlocked()
}

// applyDefaultsUnlocked sets defaults without acquiring the lock.
// Returns true if any default was applied.
func applyDefaultsUnlocked() bool {
	changed := false
	if globalConfig == nil {
		globalConfig = &Config{} // Ensure globalConfig is initialized
		changed = true          // Mark as changed if we had to create it
	}
	// Apply defaults using temporary default struct
	d := getDefaultConfig()

	// --- System Level (Legacy/Compatibility) ---
	// Note: These should ideally be fully within their respective structs
	if !globalConfig.SRTEnabled && d.SRTEnabled { // Check if false, apply default if true
		globalConfig.SRTEnabled = d.SRTEnabled
		// changed = true // Don't mark changed for legacy fields? Or sync with struct?
	}
	if !globalConfig.IcecastEnabled && d.IcecastEnabled {
		globalConfig.IcecastEnabled = d.IcecastEnabled
		// changed = true
	}


	// --- Audio Settings ---
	if globalConfig.AudioSettings.Device == "" { globalConfig.AudioSettings.Device = d.AudioSettings.Device; changed = true }
	if globalConfig.AudioSettings.SampleRate == 0 { globalConfig.AudioSettings.SampleRate = d.AudioSettings.SampleRate; changed = true }
	if globalConfig.AudioSettings.BitDepth == 0 { globalConfig.AudioSettings.BitDepth = d.AudioSettings.BitDepth; changed = true }
	if len(globalConfig.AudioSettings.MasterChannels) == 0 { globalConfig.AudioSettings.MasterChannels = d.AudioSettings.MasterChannels; changed = true }
	// Legacy fields - apply default only if zero
	if globalConfig.AudioSettings.Bitrate == 0 { globalConfig.AudioSettings.Bitrate = d.AudioSettings.Bitrate }
	if globalConfig.AudioSettings.Channels == 0 { globalConfig.AudioSettings.Channels = d.AudioSettings.Channels }


	// --- Icecast Settings ---
	// Assume struct exists if globalConfig != nil
	// Only apply default if empty/zero, respecting user's choice to disable
	if !globalConfig.IcecastSettings.IcecastEnabled && d.IcecastSettings.IcecastEnabled { // Sync top-level if struct is enabled by default
		globalConfig.IcecastSettings.IcecastEnabled = d.IcecastSettings.IcecastEnabled
		// changed = true // Only set changed if field itself was missing/zero initially?
	}
	// Apply defaults for connection details only if empty
	if globalConfig.IcecastSettings.IcecastHost == "" { globalConfig.IcecastSettings.IcecastHost = d.IcecastSettings.IcecastHost; changed = true }
	if globalConfig.IcecastSettings.IcecastPort == 0 { globalConfig.IcecastSettings.IcecastPort = d.IcecastSettings.IcecastPort; changed = true }
	if globalConfig.IcecastSettings.IcecastMount == "" { globalConfig.IcecastSettings.IcecastMount = d.IcecastSettings.IcecastMount; changed = true }
	if globalConfig.IcecastSettings.IcecastPassword == "" { globalConfig.IcecastSettings.IcecastPassword = d.IcecastSettings.IcecastPassword; changed = true } // Consider security implications of default password
	if globalConfig.IcecastSettings.IcecastBitrate == 0 { globalConfig.IcecastSettings.IcecastBitrate = d.IcecastSettings.IcecastBitrate; changed = true }
	// Legacy metadata fields - apply default only if empty
	if globalConfig.IcecastSettings.StreamName == "" { globalConfig.IcecastSettings.StreamName = d.IcecastSettings.StreamName }
	if globalConfig.IcecastSettings.StreamGenre == "" { globalConfig.IcecastSettings.StreamGenre = d.IcecastSettings.StreamGenre }
	if globalConfig.IcecastSettings.StreamDescription == "" { globalConfig.IcecastSettings.StreamDescription = d.IcecastSettings.StreamDescription }
	if globalConfig.IcecastSettings.ServerType == "" { globalConfig.IcecastSettings.ServerType = d.IcecastSettings.ServerType }
	if globalConfig.IcecastSettings.URL == "" { globalConfig.IcecastSettings.URL = d.IcecastSettings.URL } // Sync with Host?


	// --- SRT Settings ---
	if !globalConfig.SrtSettings.SrtEnabled && d.SrtSettings.SrtEnabled { // Sync top-level if struct is enabled by default
		globalConfig.SrtSettings.SrtEnabled = d.SrtSettings.SrtEnabled
		// changed = true
	}
	if globalConfig.SrtSettings.SrtHost == "" { globalConfig.SrtSettings.SrtHost = d.SrtSettings.SrtHost; changed = true }
	if globalConfig.SrtSettings.SrtPort == 0 { globalConfig.SrtSettings.SrtPort = d.SrtSettings.SrtPort; changed = true }
	if globalConfig.SrtSettings.SrtBitrate == 0 { globalConfig.SrtSettings.SrtBitrate = d.SrtSettings.SrtBitrate; changed = true }


	// --- Auto Record Settings ---
	// Assume struct exists
	// Only apply default if empty/zero, respecting user's choice to disable
	if !globalConfig.AutoRecord.Enabled && d.AutoRecord.Enabled {
		globalConfig.AutoRecord.Enabled = d.AutoRecord.Enabled
		// changed = true
	}
	if globalConfig.AutoRecord.Directory == "" { globalConfig.AutoRecord.Directory = d.AutoRecord.Directory; changed = true }
	if globalConfig.AutoRecord.PrerollDuration == 0 { globalConfig.AutoRecord.PrerollDuration = d.AutoRecord.PrerollDuration; changed = true }
	// VadThreshold (legacy gvad) - apply default if zero
	if globalConfig.AutoRecord.VadThreshold == 0.0 { globalConfig.AutoRecord.VadThreshold = d.AutoRecord.VadThreshold }
	// VadDbThreshold (new level) - apply default if zero
	if globalConfig.AutoRecord.VadDbThreshold == 0.0 { globalConfig.AutoRecord.VadDbThreshold = d.AutoRecord.VadDbThreshold; changed = true }
	if !globalConfig.AutoRecord.SmartSplitEnabled && d.AutoRecord.SmartSplitEnabled {
		 globalConfig.AutoRecord.SmartSplitEnabled = d.AutoRecord.SmartSplitEnabled
		 // changed = true
	}
	if globalConfig.AutoRecord.SmartSplitTimeout == 0 { globalConfig.AutoRecord.SmartSplitTimeout = d.AutoRecord.SmartSplitTimeout; changed = true }
	// TimeoutSeconds (legacy) - apply default if zero
	if globalConfig.AutoRecord.TimeoutSeconds == 0 { globalConfig.AutoRecord.TimeoutSeconds = d.AutoRecord.TimeoutSeconds }


	// --- Network Settings ---
	// Assume struct exists
	if globalConfig.NetworkSettings.SignalingURL == "" { globalConfig.NetworkSettings.SignalingURL = d.NetworkSettings.SignalingURL; changed = true }
	if globalConfig.NetworkSettings.StunURL == "" { globalConfig.NetworkSettings.StunURL = d.NetworkSettings.StunURL; changed = true }


	// Sync top-level legacy bools with struct bools if structs exist
	globalConfig.SRTEnabled = globalConfig.SrtSettings.SrtEnabled
	globalConfig.IcecastEnabled = globalConfig.IcecastSettings.IcecastEnabled

	return changed
}

// getDefaultConfig returns a Config struct with default values
func getDefaultConfig() Config {
	return Config{
		// Set defaults for all fields, including nested structs
		SRTEnabled:     true, // Legacy/compatibility
		IcecastEnabled: true, // Legacy/compatibility
		AudioSettings: AudioSettings{
			Device:         "", // Default PipeWire path/name (often empty string works)
			SampleRate:     48000,
			BitDepth:       24,
			MasterChannels: []int{1, 2},
			Bitrate:  48000, // Legacy - should match SampleRate? Or Opus bitrate?
			Channels: 2, // Legacy
		},
		IcecastSettings: IcecastSettings{
			IcecastEnabled:    true,
			IcecastHost:       "127.0.0.1",
			IcecastPort:       8000,
			IcecastMount:      "/stream", // Ensure leading slash
			IcecastPassword:   "hackme",
			IcecastBitrate:    192000,      // Default specific bitrate
			StreamName:        "Nixon Stream",
			StreamGenre:       "Various",
			StreamDescription: "Live from the studio",
			ServerType:        "icecast2",
			URL:               "127.0.0.1", // Legacy - sync with Host
		},
		SrtSettings: SrtSettings{
			SrtEnabled: true,
			SrtHost:    "127.0.0.1", // Default host
			SrtPort:    9000,        // Default port
			SrtBitrate: 128000,      // Default specific bitrate
		},
		AutoRecord: AutoRecordSettings{
			Enabled:           false,
			Directory:         "./recordings", // Default directory
			PrerollDuration:   15,
			VadThreshold:      0.7, // Legacy gvad default
			VadDbThreshold:    -50.0, // New level default
			SmartSplitEnabled: false,
			SmartSplitTimeout: 10,
			TimeoutSeconds:    300, // Legacy default
		},
		NetworkSettings: NetworkSettings{
			SignalingURL: "",
			StunURL:      "stun:stun.l.google.com:19302",
		},
	}
}

// --- Access & Update ---

// GetConfig returns a read-only copy of the current configuration.
// It ensures Initialize has been called.
func GetConfig() Config {
	Initialize() // Ensure initialized (safe due to sync.Once)
	configMutex.RLock()
	defer configMutex.RUnlock()
	if globalConfig == nil {
		// This should not happen if Initialize worked, but return default as failsafe
		log.Println("Warning: GetConfig called before successful initialization, returning defaults.")
		return getDefaultConfig()
	}
	// Return a copy to prevent modification through the reference
	configCopy := *globalConfig
	return configCopy
}

// SetGlobalConfig updates the global configuration and saves it atomically.
// It ensures Initialize has been called.
func SetGlobalConfig(newConfig *Config) error {
	Initialize() // Ensure initialized
	configMutex.Lock()
	defer configMutex.Unlock()

	// Perform validation or merging if needed before assignment
	// For now, assume newConfig is complete and valid

	// Sync legacy bools before saving
	newConfig.SRTEnabled = newConfig.SrtSettings.SrtEnabled
	newConfig.IcecastEnabled = newConfig.IcecastSettings.IcecastEnabled

	globalConfig = newConfig // Update the global pointer

	return saveConfigUnlocked() // Save the updated config
}

// SaveConfig saves the current global configuration state atomically.
// Public wrapper for saveConfigUnlocked, acquires lock.
// It ensures Initialize has been called.
func SaveConfig() error {
	Initialize() // Ensure initialized
	configMutex.Lock() // Lock needed to ensure consistency during save
	defer configMutex.Unlock()
	return saveConfigUnlocked()
}

// saveConfigUnlocked saves the current globalConfig state atomically without acquiring the lock.
// Must be called with configMutex held.
func saveConfigUnlocked() error {
	if globalConfig == nil {
		return fmt.Errorf("cannot save nil configuration")
	}
	data, err := json.MarshalIndent(globalConfig, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Use natefinch/atomic for safe writing
	// Corrected: Wrap data in bytes.NewReader
	// Corrected: Removed incorrect third argument (file mode)
	var reader io.Reader = bytes.NewReader(data)
	err = atomic.WriteFile(configFile, reader) // Removed 0644
	if err != nil {
		return fmt.Errorf("failed to write config file atomically: %w", err)
	}
	return nil
}

// --- Specific Update Functions (Consider removing/refactoring) ---
// These now just call SetGlobalConfig, might be redundant unless
// they perform specific validation/logic before updating.

func UpdateIcecast(settings IcecastSettings) error {
	cfg := GetConfig() // Get a copy
	cfg.IcecastSettings = settings
	return SetGlobalConfig(&cfg)
}

func UpdateSystem(srt, icecast bool, autoRecord AutoRecordSettings) error {
	cfg := GetConfig() // Get a copy
	// Update relevant parts based on parameters
	// Note: This function mixes concerns (SRT/Icecast enable flags + AutoRecord settings)
	// It might be better handled by updating structs directly.
	cfg.SrtSettings.SrtEnabled = srt // Assuming sync logic is sound
	cfg.IcecastSettings.IcecastEnabled = icecast // Assuming sync logic is sound
	cfg.AutoRecord = autoRecord
	return SetGlobalConfig(&cfg)
}

func UpdateAudio(settings AudioSettings) error {
	cfg := GetConfig() // Get a copy
	cfg.AudioSettings = settings
	return SetGlobalConfig(&cfg)
}

