package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	Web      WebSettings      `mapstructure:"web"`
	Audio    AudioSettings    `mapstructure:"audio"`
	AutoRec  AutoRecord       `mapstructure:"autoRecord"`
	Icecast  IcecastSettings  `mapstructure:"icecast"`
	SRT      SrtSettings      `mapstructure:"srt"`
	Database DatabaseSettings `mapstructure:"database"`
}

// WebSettings configures the web server
type WebSettings struct {
	ListenAddress string `mapstructure:"listenAddress"`
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

var AppConfig Config

// LoadConfig reads configuration from file or environment variables.
func LoadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/nixon/")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// --- THIS BLOCK WAS ADDED TO IMPLEMENT TASK 1.2 FROM DEVPLAN.TXT ---
	viper.SetDefault("web.listenAddress", ":8080")
	viper.SetDefault("database.path", "nixon.db")
	viper.SetDefault("audio.deviceName", "default")
	viper.SetDefault("audio.sampleRate", 48000)
	viper.SetDefault("autoRecord.enabled", false)
	viper.SetDefault("autoRecord.vadThreshold", 0.7)
	viper.SetDefault("autoRecord.vadGraceTime", 2)
	viper.SetDefault("autoRecord.maxRecordMins", 60)
	viper.SetDefault("icecast.enabled", false)
	viper.SetDefault("srt.enabled", false)
	// --- END OF ADDED BLOCK ---

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Config file not found, using defaults.")
		} else {
			log.Fatalf("Fatal error config file: %s \n", err)
		}
	}

	err = viper.Unmarshal(&AppConfig)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}
}
