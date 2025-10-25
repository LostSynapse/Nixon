package config

import (
	"log"
	"sync"
	"time"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	Database        DatabaseSettings     `json:"database"`
	Server          ServerSettings       `json:"server"`
	Audio           AudioSettings        `json:"audio"`
	Recording       RecordingSettings    `json:"recording"` // This is the correct field
	AutoRecord      AutoRecordSettings   `json:"autoRecord"`
	SrtSettings     SrtSettings          `json:"srtSettings"`
	IcecastSettings IcecastSettings      `json:"icecastSettings"`
	StreamConfigured bool                 `json:"streamConfigured"` // Flag if streams are setup
	// FIXED: Removed redundant/incorrect line below
	// RecordingSettings common.RecordingSettings `json:"recordingSettings"`
}

// DatabaseSettings holds database config
type DatabaseSettings struct {
	DSN string `json:"dsn"`
}

// ServerSettings holds server config
type ServerSettings struct {
	ListenAddr string `json:"listenAddr"`
}

// AudioSettings holds audio processing config
type AudioSettings struct {
	DeviceName     string  `json:"deviceName"`
	MasterChannels int     `json:"masterChannels"`
	BitDepth       int     `json:"bitDepth"`
	VADThreshold   float64 `json:"vadThreshold"` // VAD threshold in dB
}

// RecordingSettings holds recording config
type RecordingSettings struct {
	Directory   string `json:"directory"`
	FilePattern string `json:"filePattern"` // e.g., {YYYY}-{MM}-{DD}_{hh}-{mm}-{ss}
	FileFormat  string `json:"fileFormat"`  // e.g., wav, flac
}

// AutoRecordSettings holds auto-record config
type AutoRecordSettings struct {
	Enabled           bool          `json:"enabled"`
	PrerollDuration   time.Duration `json:"prerollDuration"`
	VadDbThreshold    float64       `json:"vadDbThreshold"`
	SmartSplitTimeout time.Duration `json:"smartSplitTimeout"`
	SmartSplitEnabled bool `json:"smartSplitEnabled"`
}

// SrtSettings holds SRT streaming config
type SrtSettings struct {
	SrtEnabled bool   `json:"srtEnabled"`
	SrtHost    string `json:"srtHost"`
	SrtPort    int    `json:"srtPort"`
	SrtBitrate int    `json:"srtBitrate"`
	SrtMode    string `json:"srtMode"` // caller, listener
}

// IcecastSettings holds Icecast streaming config
type IcecastSettings struct {
	IcecastEnabled   bool   `json:"icecastEnabled"`
	IcecastHost      string `json:"icecastHost"`
	IcecastPort      int    `json:"icecastPort"`
	IcecastUser      string `json:"icecastUser"`
	IcecastPassword  string `json:"icecastPassword"`
	IcecastMount     string `json:"icecastMount"`
	IcecastBitrate   int    `json:"icecastBitrate"`
	IcecastGenre     string `json:"icecastGenre"`
	IcecastName      string `json:"icecastName"`
	IcecastPublic    bool   `json:"icecastPublic"`
	IcecastDescription string `json:"icecastDescription"`
}

var (
	conf Config
	once sync.Once
	v    *viper.Viper
)

// LoadConfig loads configuration from file
func LoadConfig() {
	once.Do(func() {
		v = viper.New()
		v.SetConfigName("config")
		v.SetConfigType("json")
		v.AddConfigPath(".")
		v.AddConfigPath("/etc/nixon/")
		v.AddConfigPath("$HOME/.nixon/")

		// Set defaults
		v.SetDefault("Database.DSN", "nixon.db")
		v.SetDefault("Server.ListenAddr", ":8080")
		v.SetDefault("Audio.DeviceName", "default")
		v.SetDefault("Audio.MasterChannels", 2)
		v.SetDefault("Audio.BitDepth", 16)
		v.SetDefault("Audio.VADThreshold", -40.0)
		v.SetDefault("Recording.Directory", "./recordings")
		v.SetDefault("Recording.FilePattern", "rec_{YYYY}-{MM}-{DD}_{hh}-{mm}-{ss}")
		v.SetDefault("Recording.FileFormat", "flac")
		v.SetDefault("AutoRecord.Enabled", false)
		v.SetDefault("AutoRecord.PrerollDuration", 5*time.Second)
		v.SetDefault("AutoRecord.VadDbThreshold", -40.0)
		v.SetDefault("AutoRecord.SmartSplitTimeout", 30*time.Second)
		v.SetDefault("AutoRecord.SmartSplitEnabled", false)
		// ... (defaults for SRT and Icecast)

		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				log.Println("Config file not found; writing default config.")
				if err := v.SafeWriteConfig(); err != nil {
					log.Printf("Error writing default config: %v", err)
				}
			} else {
				log.Printf("Error reading config file: %v", err)
			}
		}

		if err := v.Unmarshal(&conf); err != nil {
			log.Fatalf("Error unmarshalling config: %v", err)
		}
	})
}

// GetConfig returns the loaded configuration
func GetConfig() Config {
	return conf
}

// SaveGlobalConfig saves the *global* in-memory configuration to disk.
func SaveGlobalConfig(cfgToSave Config) error {
	v.Set("Database", cfgToSave.Database)
	v.Set("Server", cfgToSave.Server)
	v.Set("Audio", cfgToSave.Audio)
	v.Set("Recording", cfgToSave.Recording)
	v.Set("AutoRecord", cfgToSave.AutoRecord)
	v.Set("SrtSettings", cfgToSave.SrtSettings)
	v.Set("IcecastSettings", cfgToSave.IcecastSettings)
	v.Set("StreamConfigured", cfgToSave.StreamConfigured)
	// FIXED: Removed line referencing non-existent field
	// v.Set("RecordingSettings", cfgToSave.RecordingSettings)


	// Update in-memory global
	conf = cfgToSave

	// Save to file
	if err := v.WriteConfig(); err != nil {
		log.Printf("Error writing config: %v", err)
		return err
	}
	return nil
}
