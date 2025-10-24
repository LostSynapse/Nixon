// internal/config/config.go
// This file manages the application configuration loaded from config.json.

package config

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time" // Added time package for Duration types
)

// Global configuration variables
var (
	cfg  Config
	once sync.Once
)

// Config represents the application's entire configuration structure.
type Config struct {
	// General settings
	AudioSettings AudioSettings `json:"audio_settings"`
	AutoRecord    AutoRecord    `json:"auto_record"`

	// Streaming settings (currently hardcoded as structs, will become modular plugins)
	SrtSettings     SrtSettings     `json:"srt_settings"`
	IcecastSettings IcecastSettings `json:"icecast_settings"`
}

// AudioSettings controls the core audio capture parameters.
type AudioSettings struct {
	// FIX: Consistent field name to satisfy gstreamer.go (was DeviceName)
	Device     string `json:"device_name"` // e.g., "alsa_input.usb-Focusrite..."
	SampleRate int    `json:"sample_rate"` // e.g., 48000
	// FIX: Added fields required by gstreamer.go
	MasterChannels int `json:"master_channels"` // e.g., 2 (stereo)
	BitDepth       int `json:"bit_depth"`       // e.g., 16, 24, or 32
}

// AutoRecord controls the file recording and cleanup parameters.
type AutoRecord struct {
	Directory         string `json:"directory"`
	AutoRecordEnabled bool   `json:"auto_record_enabled"`
	MaxDiskUsage      int    `json:"max_disk_usage"` // Percentage
	// FIX: Aligned field names and types with gstreamer.go requirements
	PrerollDuration   time.Duration `json:"preroll_duration"`     // Buffering duration before recording (e.g., 10s)
	VadDbThreshold    float64       `json:"vad_db_threshold"`     // Threshold for Voice Activity Detection in dB (e.g., -40.0)
	SmartSplitTimeout time.Duration `json:"smart_split_timeout"`  // Silence threshold for smart file splitting (e.g., 5m)
	// FIX: Added missing boolean fields required by gstreamer.go
	Enabled           bool          `json:"enabled"`              // Global control for auto-record functionality
	SmartSplitEnabled bool          `json:"smart_split_enabled"`  // Control for smart file splitting
	// NOTE: PreRollSeconds was replaced by PrerollDuration (time.Duration is superior for config)
}

// SrtSettings holds configuration for the SRT streaming output.
type SrtSettings struct {
	SrtEnabled bool   `json:"srt_enabled"`
	Mode       string `json:"mode"` // "listener" or "caller"
	// FIX: Added fields required by gstreamer.go
	SrtHost    string `json:"srt_host"`
	SrtPort    int    `json:"srt_port"`
	SrtBitrate int    `json:"srt_bitrate"` // Streaming bitrate in kbps (e.g., 128)
	LatencyMS  int    `json:"latency_ms"` // SRT latency in milliseconds
}

// IcecastSettings holds configuration for the Icecast streaming output.
type IcecastSettings struct {
	IcecastEnabled bool `json:"icecast_enabled"`
	// FIX: Added IcecastBitrate required by gstreamer.go
	IcecastBitrate  int    `json:"icecast_bitrate"` // Streaming bitrate in kbps (e.g., 128)
	IcecastHost     string `json:"icecast_host"`
	IcecastPort     int    `json:"icecast_port"`
	IcecastMount    string `json:"icecast_mount"`
	IcecastUser     string `json:"icecast_user"`
	IcecastPassword string `json:"icecast_password"`
}

// LoadConfig initializes the configuration singleton from config.json.
func LoadConfig() {
	once.Do(func() {
		data, err := os.ReadFile("config.json")
		if err != nil {
			log.Printf("WARNING: config.json not found or could not be read: %v. Using defaults.", err)
			// Use reasonable defaults if file is missing
			cfg = Config{
				AudioSettings: AudioSettings{Device: "default", SampleRate: 48000, MasterChannels: 2, BitDepth: 24},
				AutoRecord: AutoRecord{
					Directory:         "./recordings",
					AutoRecordEnabled: false,
					PrerollDuration:   10 * time.Second, // 10 seconds default
					MaxDiskUsage:      80,
					VadDbThreshold:    -40.0,
					SmartSplitTimeout: 5 * time.Minute, // 5 minutes default
					Enabled:           false,
					SmartSplitEnabled: false,
				},
				SrtSettings: SrtSettings{SrtEnabled: false, Mode: "listener", SrtPort: 9000, SrtBitrate: 128},
				IcecastSettings: IcecastSettings{
					IcecastEnabled: false,
					IcecastMount:   "/stream",
					IcecastHost:    "127.0.0.1",
					IcecastPort:    8000,
					IcecastBitrate: 128,
				},
			}
			return
		}

		if err := json.Unmarshal(data, &cfg); err != nil {
			log.Fatalf("FATAL: Error unmarshaling config.json: %v", err)
		}
		log.Println("Configuration loaded successfully.")
	})
}

// GetConfig returns the current global configuration.
func GetConfig() Config {
	return cfg
}

// SaveGlobalConfig serializes the current configuration to config.json.
func SaveGlobalConfig(c Config) error {
	cfg = c // Update the runtime configuration first

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	// Use 0644 for file permissions
	if err := os.WriteFile("config.json", data, 0644); err != nil {
		return err
	}
	return nil
}
