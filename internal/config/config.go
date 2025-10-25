package config

import (
	"log"
	"sync"

	"github.com/spf13/viper"
)

// Config holds the application's configuration.
type Config struct {
	Pipewire        PipewireSettings  `json:"pipewire"`
	Recording       RecordingSettings `json:"recording"`
	NetworkSettings NetworkSettings   `json:"networkSettings"`
	SrtSettings     SrtSettings       `json:"srtSettings"`
	IcecastSettings IcecastSettings   `json:"icecastSettings"`
}

// PipewireSettings holds PipeWire-related config.
type PipewireSettings struct {
	Socket string `json:"socket"`
}

// RecordingSettings holds recording configuration.
type RecordingSettings struct {
	Enabled     bool   `json:"enabled"`
	Directory   string `json:"directory"`
	FilePattern string `json:"filePattern"`
	FileFormat  string `json:"fileFormat"`
}

// NetworkSettings holds WebRTC and signaling server config.
type NetworkSettings struct {
	SignalingURL string `json:"signalingUrl"`
	StunURL      string `json:"stunUrl"`
}

// SrtSettings holds SRT streaming config.
type SrtSettings struct {
	SrtEnabled bool   `json:"srtEnabled"`
	SrtHost    string `json:"srtHost"`
	SrtPort    int    `json:"srtPort"`
	SrtStream  string `json:"srtStream"`
	SrtLatency int    `json:"srtLatency"`
}

// IcecastSettings holds Icecast streaming config.
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

// Load initializes the configuration from file.
// It returns the loaded config and an error if one occurred.
func Load() (Config, error) {
	var err error
	once.Do(func() {
		v = viper.New()
		v.SetConfigName("config")
		v.SetConfigType("json")
		v.AddConfigPath(".")
		v.AddConfigPath("/etc/nixon/")
		v.AddConfigPath("$HOME/.nixon/")

		// Set default values
		v.SetDefault("pipewire.socket", "pipewire-0")
		v.SetDefault("recording.enabled", true)
		v.SetDefault("recording.directory", "./recordings")
		v.SetDefault("recording.filePattern", "rec_{YYYY}-{MM}-{DD}_{hh-mm-ss}")
		v.SetDefault("recording.fileFormat", "flac")

		if err = v.ReadInConfig(); err != nil {
			log.Printf("Warning: Could not read config file: %v. Using defaults.", err)
		}

		if err = v.Unmarshal(&conf); err != nil {
			log.Fatalf("Fatal error: could not unmarshal config: %v", err)
		}
	})
	return conf, err
}

// GetConfig returns the singleton config instance.
func GetConfig() Config {
	return conf
}

// SetConfig sets the global config instance (useful for testing).
func SetConfig(c Config) {
	conf = c
}
