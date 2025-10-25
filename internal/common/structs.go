package common

import "time"

// AudioState represents the high-level state of the audio manager
type AudioState string

// Defines the possible states of the audio manager
const (
	StateStopped   AudioState = "stopped"
	StateRecording AudioState = "recording"
	StateStreaming AudioState = "streaming" // Example future state
)

// AudioStatus defines the real-time status of the audio manager
type AudioStatus struct {
	State          AudioState `json:"state"`
	CurrentRecFile string     `json:"currentRecFile"`
	IsAutoRec      bool       `json:"isAutoRec"`

	// --- Fields required by pipewire.go ---
	ActiveStreams map[string]bool `json:"activeStreams"`
	VADStatus     bool            `json:"vadStatus"`
	MasterPeak    float64         `json:"masterPeak"`    // Current master peak in dB
	LastVADEvent  time.Time       `json:"lastVadEvent"` // Last time VAD triggered
}

// AudioDevice represents a single discoverable audio device
type AudioDevice struct {
	DeviceName  string `json:"deviceName"`
	Driver      string `json:"driver"`
	Description string `json:"description"`
}

// AudioCapabilities defines the supported formats/rates of a device
type AudioCapabilities struct {
	Formats     []string `json:"formats"`
	SampleRates []int    `json:"sampleRates"`
	Channels    []int    `json:"channels"`
}

// Recording represents a single recording entry in the database
type Recording struct {
	ID        uint          `json:"id" gorm:"primaryKey"`
	Filename  string        `json:"filename"`
	StartTime time.Time     `json:"startTime"`
	EndTime   time.Time     `json:"endTime"`
	Duration  time.Duration `json:"duration"`
	FileSize  int64         `json:"fileSize"`
	Notes     string        `json:"notes"`
	Genre     string        `json:"genre"`
}

