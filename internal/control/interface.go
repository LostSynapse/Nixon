package control

import "nixon/internal/common"

// ControlHandler defines the high-level interface for managing the audio engine.
// This abstraction allows the API layer to remain decoupled from the underlying
// implementation (e.g., GStreamer, PipeWire).
type ControlHandler interface {
	// State Management
	GetStatus() common.AudioStatus
	StartAudio() error
	StopAudio() error

	// Recording Control
	StartRecording() error
	StopRecording() error

	// Streaming Control
	StartStream(streamType string) error
	StopStream(streamType string) error

	// Device Management
	GetAudioDevices() ([]common.AudioDevice, error)
}
