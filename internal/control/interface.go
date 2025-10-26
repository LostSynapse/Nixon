package control

import (
	"nixon/internal/common"
	"nixon/internal/config"
	"time"
)

// ControlHandler defines the abstraction layer for all system control
type ControlHandler interface {
	// --- Audio Pipeline ---
	StartRecording() (string, error)
	StopRecording() error
	GetAudioStatus() common.AudioStatus

	// --- Streaming ---
	StartStream(streamName string) error
	StopStream(streamName string) error

	// --- Configuration ---
	ReloadConfig()
	GetConfig() config.Config
	SaveConfig(cfg config.Config) error

	// --- Hardware ---
	ListAudioDevices() ([]common.AudioDevice, error)
	GetAudioCapabilities(deviceName string) (common.AudioCapabilities, error)

	// --- Database / Recordings ---
	GetAllRecordings() ([]common.Recording, error)
	GetRecordingByID(id uint) (*common.Recording, error)
	UpdateRecording(id uint, notes string, genre string, endTime time.Time, duration time.Duration) error
	DeleteRecording(id uint) error
}

