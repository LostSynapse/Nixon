// internal/common/structs.go
// This package holds common, dependency-agnostic structs shared across modules
// (e.g., GStreamer, PipeWire, API).

package common

// AudioCapabilities represents the parsed hardware capabilities (used by API).
// This is the standards-based data structure for device configuration.
type AudioCapabilities struct {
	Rates  []int `json:"rates"`
	Depths []int `json:"depths"`
}

// AudioDevice represents a discovered audio device (used by API).
type AudioDevice struct {
	ID          string `json:"id"`           // Unique identifier (e.g., PipeWire node name or ALSA hw:X,Y)
	Name        string `json:"name"`         // User-friendly display name
	Description string `json:"description"`  // More detailed description
	API         string `json:"api"`          // e.g., "pipewire", "alsa"
	Class       string `json:"device_class"` // e.g., "Audio/Source", "Audio/Sink"
}
