// internal/config/config.go
package config

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var (
	globalConfig *Config
	configOnce   sync.Once
	configMutex  sync.RWMutex
	configFile   = "config.json"
)

// loadConfigFromFile initializes the configuration from disk.
// It assumes it is called only once by GetConfig's sync.Once.
func loadConfigFromFile() error {
	// Note: globalConfig is *already* initialized with defaults
	// by GetConfig(). This function just loads from disk.
	configMutex.Lock()
	defer configMutex.Unlock()

	// 2. Try to read the existing config file
	data, err := os.ReadFile(configFile)
	if os.IsNotExist(err) {
		// File doesn't exist, save the defaults that are already in globalConfig
		log.Printf("%s not found, creating with default values.", configFile)
		return saveConfigUnlocked(globalConfig)
	} else if err != nil {
		// Other read error (e.g., permissions)
		log.Printf("Failed to read %s: %v. Using in-memory defaults.", configFile, err)
		return err // Return the error, but globalConfig is non-nil
	}

	// 3. File exists, try to unmarshal it
	var conf Config
	if err := json.Unmarshal(data, &conf); err != nil {
		log.Printf("config.json is malformed: %v. Backing up and using defaults.", err)
		// Attempt to backup bad config
		backupFile := fmt.Sprintf("config.bad.%d.json", os.Getpid())
		if err_bak := os.Rename(configFile, backupFile); err_bak == nil {
			log.Printf("Backed up malformed config to %s", backupFile)
		}
		// Save the default config we created in step 1
		return saveConfigUnlocked(globalConfig)
	}

	// 4. File exists and is valid.
	// Overwrite globalConfig with loaded values, then re-apply defaults
	// for any fields that were missing from the file.
	globalConfig = &conf
	defaultsApplied := applyDefaults(globalConfig) // This merges loaded values with defaults

	if defaultsApplied {
		log.Println("Applied default values to missing config fields. Re-saving.")
		return saveConfigUnlocked(globalConfig) // Save merged config
	}

	// File was perfect, no save needed
	log.Println("Configuration loaded from config.json")
	return nil
}

// GetConfig returns the global config
func GetConfig() *Config {
	configOnce.Do(func() {
		// 1. ALWAYS initialize globalConfig with defaults first, under a lock.
		// This guarantees GetConfig() can never return nil.
		configMutex.Lock()
		globalConfig = &Config{}
		applyDefaults(globalConfig) // Populate the new struct with defaults
		configMutex.Unlock()

		// 2. Now that defaults are set, try to load/create the file on disk.
		if err := loadConfigFromFile(); err != nil {
			log.Printf("WARNING: Error during initial config load: %v. Server is using defaults.", err)
		}
	})
	configMutex.RLock()
	defer configMutex.RUnlock()
	return globalConfig // This is now guaranteed to be non-nil
}

// Set updates the configuration and saves it
func Set(newConfig *Config) error {
	configMutex.Lock()
	defer configMutex.Unlock()

	// Update the global config
	globalConfig = newConfig

	// Save the new config to disk
	return saveConfigUnlocked(globalConfig)
}

// saveConfigUnlocked saves the current global config to disk.
// Assumes a write lock is already held.
func saveConfigUnlocked(conf *Config) error {
	// Create a temp file
	tmpFile, err := os.CreateTemp(filepath.Dir(configFile), "config-*.tmp")
	if err != nil {
		log.Printf("Failed to create temp config file: %v", err)
		return err
	}
	defer os.Remove(tmpFile.Name()) // Clean up temp file

	// Write pretty JSON to temp file
	encoder := json.NewEncoder(tmpFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(conf); err != nil {
		log.Printf("Failed to write to temp config file: %v", err)
		tmpFile.Close()
		return err
	}
	if err := tmpFile.Close(); err != nil {
		log.Printf("Failed to close temp config file: %v", err)
		return err
	}

	// Atomically rename temp file to the final config file
	if err := os.Rename(tmpFile.Name(), configFile); err != nil {
		log.Printf("Failed to atomically move config file: %v", err)
		// Fallback for cross-device link (e.g., Docker mounts)
		if fin, err_in := os.Open(tmpFile.Name()); err_in == nil {
			defer fin.Close()
			if fout, err_out := os.Create(configFile); err_out == nil {
				defer fout.Close()
				if _, err_cp := io.Copy(fout, fin); err_cp == nil {
					log.Println("Fallback config copy successful.")
					return nil
				}
			}
		}
		log.Printf("FATAL: Failed to save config file: %v", err)
		return err
	}

	log.Println("Configuration saved to config.json")
	return nil
}

// applyDefaults ensures all config fields have values and returns true if any were changed.
func applyDefaults(c *Config) bool {
	changed := false

	// --- AudioSettings ---
	if c.AudioSettings.Device == "" {
		c.AudioSettings.Device = "default"
		changed = true
	}
	if c.AudioSettings.SampleRate == 0 {
		c.AudioSettings.SampleRate = 48000
		changed = true
	}
	if c.AudioSettings.BitDepth == 0 {
		c.AudioSettings.BitDepth = 24
		changed = true
	}
	if len(c.AudioSettings.MasterChannels) == 0 {
		c.AudioSettings.MasterChannels = []int{1, 2}
		changed = true
	}

	// --- SrtSettings ---
	if c.SrtSettings.SrtHost == "" {
		c.SrtSettings.SrtHost = "127.0.0.1"
		changed = true
	}
	if c.SrtSettings.SrtPort == 0 {
		c.SrtSettings.SrtPort = 9000
		changed = true
	}
	if c.SrtSettings.SrtBitrate == 0 {
		c.SrtSettings.SrtBitrate = 128000
		changed = true
	}
	// SrtEnabled defaults to false (zero-value for bool), which is fine.

	// --- IcecastSettings ---
	if c.IcecastSettings.IcecastHost == "" {
		c.IcecastSettings.IcecastHost = "127.0.0.1"
		changed = true
	}
	if c.IcecastSettings.IcecastPort == 0 {
		c.IcecastSettings.IcecastPort = 8000
		changed = true
	}
	if c.IcecastSettings.IcecastMount == "" {
		c.IcecastSettings.IcecastMount = "stream"
		changed = true
	}
	if c.IcecastSettings.IcecastPassword == "" {
		c.IcecastSettings.IcecastPassword = "hackme"
		changed = true
	}
	if c.IcecastSettings.IcecastBitrate == 0 {
		c.IcecastSettings.IcecastBitrate = 192000
		changed = true
	}
	// IcecastEnabled defaults to false, which is fine.

	// --- AutoRecord ---
	if c.AutoRecord.Directory == "" {
		c.AutoRecord.Directory = "./recordings"
		changed = true
	}
	if c.AutoRecord.PrerollDuration == 0 {
		c.AutoRecord.PrerollDuration = 15
		changed = true
	}
	if c.AutoRecord.VadDbThreshold == 0 {
		c.AutoRecord.VadDbThreshold = -50.0
		changed = true
	}
	if c.AutoRecord.SmartSplitTimeout == 0 {
		c.AutoRecord.SmartSplitTimeout = 10
		changed = true
	}
	// Enabled/SmartSplitEnabled default to false, which is fine.

	// --- NetworkSettings ---
	if c.NetworkSettings.SignalingURL == "" {
		c.NetworkSettings.SignalingURL = "" // Default is empty
		changed = true
	}
	if c.NetworkSettings.StunURL == "" {
		c.NetworkSettings.StunURL = "stun:stun.l.google.com:19302"
		changed = true
	}

	return changed
}

// --- Config Structs ---

type Config struct {
	AudioSettings   AudioSettings   `json:"audio_settings"`
	SrtSettings     SrtSettings     `json:"srt_settings"`
	IcecastSettings IcecastSettings `json:"icecast_settings"`
	AutoRecord      AutoRecord      `json:"auto_record"`
	NetworkSettings NetworkSettings `json:"network_settings"`
}

type AudioSettings struct {
	Device         string `json:"device"`
	SampleRate     int    `json:"sample_rate"`
	BitDepth       int    `json:"bit_depth"`
	MasterChannels []int  `json:"master_channels"`
}

type SrtSettings struct {
	SrtEnabled bool   `json:"srt_enabled"`
	SrtHost    string `json:"srt_host"`
	SrtPort    int    `json:"srt_port"`
	SrtBitrate int    `json:"srt_bitrate"`
}

type IcecastSettings struct {
	IcecastEnabled  bool   `json:"icecast_enabled"`
	IcecastHost     string `json:"icecast_host"`
	IcecastPort     int    `json:"icecast_port"`
	IcecastMount    string `json:"icecast_mount"`
	IcecastPassword string `json:"icecast_password"`
	IcecastBitrate  int    `json:"icecast_bitrate"`
}

type AutoRecord struct {
	Enabled           bool    `json:"enabled"`
	Directory         string  `json:"directory"`
	PrerollDuration   int     `json:"preroll_duration"`
	VadDbThreshold    float64 `json:"vad_db_threshold"`
	SmartSplitEnabled bool    `json:"smart_split_enabled"`
	SmartSplitTimeout int     `json:
"smart_split_timeout"`
}

type NetworkSettings struct {
	SignalingURL string `json:"signaling_url"`
	StunURL      string `json:"stun_url"`
}

