package config

import (
	"github.com/spf13/viper"
	"nixon/internal/slogger"
	"os"
	"strings"
)

// Config holds the application configuration
type Config struct {
	Web      WebSettings      `mapstructure:"web"`
	Audio    AudioSettings    `mapstructure:"audio"`
	AutoRec  AutoRecord       `mapstructure:"autoRecord"`
	Icecast  IcecastSettings  `mapstructure:"icecast"`
	SRT      SrtSettings      `mapstructure:"srt"`
	Database DatabaseSettings `mapstructure:"database"`
	Pipewire PipewireSettings `mapstructure:"pipewire"`
}

// WebSettings configures the web server
type WebSettings struct {
	ListenAddress string `mapstructure:"listenAddress"`
	Secret        string `mapstructure:"secret"`
	WebDevServerURL string `mapstructure:"webDevServerURL"` // ADDED: URL for the Vite development server
}

// AudioSettings configures the audio processing
type AudioSettings struct {
	DeviceName string `mapstructure:"deviceName"`
	SampleRate int    `mapstructure:"sampleRate"`
}

// AutoRecord configures the automatic recording feature
type AutoRecord struct {
	Enabled       bool    `mapstructure:"enabled"`
	VADThreshold  float64 `mapstructure:"vadThreshold"`
	VADGraceTime  int     `mapstructure:"vadGraceTime"`
	MaxRecordMins int     `mapstructure:"maxRecordMins"`
}

// IcecastSettings configures the Icecast output
type IcecastSettings struct {
	Enabled      bool   `mapstructure:"enabled"`
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	Mountpoint   string `mapstructure:"mountpoint"`
	StreamName   string `mapstructure:"streamName"`
	StreamDesc   string `mapstructure:"streamDesc"`
	StreamURL    string `mapstructure:"streamURL"`
	StreamGenre  string `mapstructure:"streamGenre"`
	StreamPublic bool   `mapstructure:"streamPublic"`
}

// SrtSettings configures the SRT output
type SrtSettings struct {
	Enabled   bool   `mapstructure:"enabled"`
	Address   string `mapstructure:"address"`
	Port      int    `mapstructure:"port"`
	Passphase string `mapstructure:"passphase"`
	Latency   int    `mapstructure:"latency"`
	StreamID  string `mapstructure:"streamId"`
}

// DatabaseSettings configures the database connection
type DatabaseSettings struct {
	Path string `mapstructure:"path"`
}

// PipewireSettings configures the Pipewire connection
type PipewireSettings struct {
	Socket string `mapstructure:"socket"`
}

var AppConfig Config

// LoadConfig reads configuration from file or environment variables.
func LoadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/nixon/")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Set default values here
	viper.SetDefault("web.listenAddress", "8080")
	viper.SetDefault("database.path", "nixon.db")
	viper.SetDefault("audio.deviceName", "default")
	viper.SetDefault("audio.sampleRate", 48000)
	viper.SetDefault("autoRecord.enabled", false)
	viper.SetDefault("autoRecord.vadThreshold", 0.7)
	viper.SetDefault("autoRecord.vadGraceTime", 2)
	viper.SetDefault("autoRecord.maxRecordMins", 60)
	viper.SetDefault("icecast.enabled", false)
	viper.SetDefault("srt.enabled", false)
	viper.SetDefault("pipewire.socket", "") // Default socket lets the library auto-discover
	viper.SetDefault("web.secret", "nixon-default-secret")
	viper.SetDefault("web.webDevServerURL", "") // ADDED: Default empty, will be set by env for dev

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			slogger.Log.Warn("Config file not found. Attempting to create 'config.json' with default values.")
			// Attempt to write a new config file. Use WriteConfigAs to specify the full filename.
if writeErr := viper.WriteConfigAs("config.json"); writeErr != nil {
	// If we can't write the config file, that's a fatal error.
	slogger.Log.Error("Failed to write new config file", "err", writeErr)
	os.Exit(1)
}
slogger.Log.Info("Successfully created 'config.json' with default settings.")

		} else {
			slogger.Log.Error("Fatal error reading config file", "err", err)
			os.Exit(1)

		}
	}

	err = viper.Unmarshal(&AppConfig)
	if err != nil {
		slogger.Log.Error("Unable to decode config into struct", "err", err)
		os.Exit(1)

	}
}
