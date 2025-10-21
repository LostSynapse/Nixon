package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

const (
	configFile    = "./config.json"
	RecordingsDir = "./recordings"
)

type IcecastSettings struct {
	URL               string `json:"url"`
	Port              string `json:"port"`
	Mount             string `json:"mount"`
	Password          string `json:"password"`
	StreamName        string `json:"stream_name"`
	StreamGenre       string `json:"stream_genre"`
	StreamDescription string `json:"stream_description"`
	ServerType        string `json:"server_type"`
}

type AudioSettings struct {
	Device   string `json:"device"`
	Bitrate  int    `json:"bitrate"`
	BitDepth int    `json:"bit_depth"`
	Channels int    `json:"channels"`
}

type AutoRecordSettings struct {
	Enabled        bool `json:"enabled"`
	TimeoutSeconds int  `json:"timeout_seconds"`
}

type Config struct {
	sync.RWMutex
	SRTEnabled     bool             `json:"srt_enabled"`
	IcecastEnabled bool             `json:"icecast_enabled"`
	Audio          AudioSettings    `json:"audio_settings"`
	Icecast        IcecastSettings  `json:"icecast_settings"`
	AutoRecord     AutoRecordSettings `json:"auto_record"`
}

var globalConfig Config

func Load() error {
	file, err := os.Open(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("config.json not found, creating a default one.")
			defaultConfig := Config{
				SRTEnabled:     true,
				IcecastEnabled: true,
				Audio: AudioSettings{
					Device:   "hw:5,0",
					Bitrate:  48000,
					BitDepth: 24,
					Channels: 2,
				},
				Icecast: IcecastSettings{
					URL:               "localhost",
					Port:              "8000",
					Mount:             "stream",
					Password:          "hackme",
					StreamName:        "Nixon Stream",
					StreamGenre:       "Various",
					StreamDescription: "Live from the studio",
					ServerType:        "icecast2",
				},
				AutoRecord: AutoRecordSettings{
					Enabled:        false,
					TimeoutSeconds: 300,
				},
			}
			if err := saveConfig(&defaultConfig); err != nil {
				return err
			}
			globalConfig = defaultConfig
			return nil
		}
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	globalConfig.Lock()
	defer globalConfig.Unlock()
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&globalConfig); err != nil {
		return fmt.Errorf("failed to decode config file: %w", err)
	}

	log.Printf("Configuration loaded successfully.")
	return nil
}

func saveConfig(cfg *Config) error {
	cfg.Lock()
	defer cfg.Unlock()
	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	return os.WriteFile(configFile, data, 0644)
}

func Get() Config {
	globalConfig.RLock()
	defer globalConfig.RUnlock()
	return globalConfig
}

func UpdateIcecast(settings IcecastSettings) error {
	globalConfig.Lock()
	globalConfig.Icecast = settings
	globalConfig.Unlock()
	return saveConfig(&globalConfig)
}

func UpdateSystem(srt, icecast bool, autoRecord AutoRecordSettings) error {
	globalConfig.Lock()
	globalConfig.SRTEnabled = srt
	globalConfig.IcecastEnabled = icecast
	globalConfig.AutoRecord = autoRecord
	globalConfig.Unlock()
	return saveConfig(&globalConfig)
}

func UpdateAudio(settings AudioSettings) error {
	globalConfig.Lock()
	globalConfig.Audio = settings
	globalConfig.Unlock()
	return saveConfig(&globalConfig)
}

