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
	State          AudioState `json:"state,omitempty"`
	CurrentRecFile string     `json:"currentRecFile,omitempty"`
	IsAutoRec      bool       `json:"isAutoRec,omitempty"`

	// --- Fields required by pipewire.go ---
	ActiveStreams map[string]bool `json:"activeStreams,omitempty"`
	VADStatus     bool            `json:"vadStatus,omitempty"`
	MasterPeak    float64         `json:"masterPeak,omitempty"`   // Current master peak in dB
	LastVADEvent  time.Time       `json:"lastVadEvent,omitempty"` // Last time VAD triggered
}

// AudioDevice represents a single discoverable audio device
type AudioDevice struct {
	DeviceName  string `json:"deviceName,omitempty"`
	Driver      string `json:"driver,omitempty"`
	Description string `json:"description,omitempty"`
}

// AudioSource represents a specific audio source, like a microphone, used by PipeWire.
type AudioSource struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// AudioCapabilities defines the supported formats/rates of a device
type AudioCapabilities struct {
	Formats     []string `json:"formats,omitempty"`
	SampleRates []int    `json:"sampleRates,omitempty"`
	Channels    []int    `json:"channels,omitempty"`
}

// Recording represents a single recording entry in the database
type Recording struct {
	ID        uint          `json:"id,omitempty" gorm:"primaryKey"`
	Filename  string        `json:"filename,omitempty"`
	StartTime time.Time     `json:"startTime,omitempty"`
	EndTime   time.Time     `json:"endTime,omitempty"`
	Duration  time.Duration `json:"duration,omitempty"`
	FileSize  int64         `json:"fileSize,omitempty"`
	Notes     string        `json:"notes,omitempty"`
	Genre     string        `json:"genre,omitempty"`
}
