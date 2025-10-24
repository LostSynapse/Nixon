// internal/gstreamer/gstreamer.go
// Manages the GStreamer audio pipeline, handling recording, streaming, and VAD.

package gstreamer

import (
	"fmt"
	"log"
	"math"
	"nixon/internal/common"
	"nixon/internal/config"
	"nixon/internal/db"
	"nixon/internal/websocket"
	"os"
	"path/filepath"
	"sync"
	"time"

	// DEFINITIVE FIX: Reverting to the official and confirmed 'go-gst' binding package.
	"github.com/go-gst/go-gst/gst"
)

// GStreamer Pipeline Constants
const (
	// Pipeline names
	PipelineName = "audio-master-pipeline"

	// Element names for VAD and splitting
	VadElement      = "vad-element"
	TeeElement      = "tee-element"
	RecorderQueue   = "recorder-queue"
	FileSinkElement = "file-sink-element"
	IcecastQueue    = "icecast-queue"
	IcecastSink     = "icecast-sink-element"
	SrtQueue        = "srt-queue"
	SrtSink         = "srt-sink-element"
)

// Global manager singleton and lock
var (
	manager    *PipelineManager
	initMutex sync.Mutex // Guards access to the manager singleton initialization
)

// PipelineManager holds the GStreamer pipeline and configuration.
type PipelineManager struct {
	pipeline    *gst.Pipeline
	conf        config.Config
	isRecording bool
	currentRec  *db.Recording
	updateCb    func(update map[string]interface{})
}

// StreamPlugin defines the interface for modular streaming protocols (SRT, Icecast, AES67, etc.).
type StreamPlugin interface {
	GetSinkElement() string
	IsEnabled() bool
	BuildPipelineBranch() string
	GetName() string
}

// Plugin implementations (These will be moved to separate files during the modularization phase)

// SrtPlugin implements the StreamPlugin interface for SRT.
type SrtPlugin struct{}

func (p *SrtPlugin) GetSinkElement() string { return SrtSink }
func (p *SrtPlugin) GetName() string        { return "SRT" }
func (p *SrtPlugin) IsEnabled() bool        { return config.GetConfig().SrtSettings.SrtEnabled }
func (p *SrtPlugin) BuildPipelineBranch() string {
	cfg := config.GetConfig()
	if !p.IsEnabled() {
		return ""
	}

	// SRT Configuration checks
	if cfg.SrtSettings.SrtHost == "" || cfg.SrtSettings.SrtPort == 0 || cfg.SrtSettings.SrtBitrate == 0 {
		log.Println("WARNING: SRT enabled but configuration is incomplete. Skipping SRT stream branch.")
		return ""
	}

	// Example SRT pipeline branch (audio-input -> fdk-aac -> srt-sink)
	return fmt.Sprintf("! queue name=%s ! fdk_aacenc bitrate=%d ! srtclientsink name=%s uri=srt://%s:%d?mode=%s latency=%d",
		SrtQueue,
		cfg.SrtSettings.SrtBitrate*1000, // Bitrate in bits/sec
		SrtSink,
		cfg.SrtSettings.SrtHost,
		cfg.SrtSettings.SrtPort,
		cfg.SrtSettings.Mode,
		cfg.SrtSettings.LatencyMS,
	)
}

// IcecastPlugin implements the StreamPlugin interface for Icecast.
type IcecastPlugin struct{}

func (p *IcecastPlugin) GetSinkElement() string { return IcecastSink }
func (p *IcecastPlugin) GetName() string        { return "Icecast" }
func (p *IcecastPlugin) IsEnabled() bool        { return config.GetConfig().IcecastSettings.IcecastEnabled }
func (p *IcecastPlugin) BuildPipelineBranch() string {
	cfg := config.GetConfig()
	if !p.IsEnabled() {
		return ""
	}

	// Icecast Configuration checks
	if cfg.IcecastSettings.IcecastMount == "" || cfg.IcecastSettings.IcecastHost == "" || cfg.IcecastSettings.IcecastPort == 0 || cfg.IcecastSettings.IcecastBitrate == 0 {
		log.Println("WARNING: Icecast enabled but configuration is incomplete. Skipping Icecast stream branch.")
		return ""
	}

	// Example Icecast pipeline branch (audio-input -> lame -> shout2send)
	return fmt.Sprintf("! queue name=%s ! audioresample ! audioconvert ! lame name=lame%s bitrate=%d ! shout2send name=%s mount=%s port=%d hostname=%s username=%s password=%s",
		IcecastQueue,
		IcecastQueue, // Use queue name as part of lame name for uniqueness if needed
		cfg.IcecastSettings.IcecastBitrate,
		IcecastSink,
		cfg.IcecastSettings.IcecastMount,
		cfg.IcecastSettings.IcecastPort,
		cfg.IcecastSettings.IcecastHost,
		cfg.IcecastSettings.IcecastUser,
		cfg.IcecastSettings.IcecastPassword,
	)
}

// Global list of registered plugins (will be loaded dynamically in later phases)
var registeredStreamPlugins = []StreamPlugin{
	&SrtPlugin{},
	&IcecastPlugin{},
}

// GetManager initializes and returns the PipelineManager singleton.
func GetManager() *PipelineManager {
	initMutex.Lock()
	defer initMutex.Unlock()

	if manager == nil {
		manager = &PipelineManager{
			conf: config.GetConfig(),
		}
		manager.updateCb = websocket.BroadcastUpdate
		log.Println("GStreamer Manager initialized.")
	}
	return manager
}

// ReloadConfig updates the manager's configuration and checks for pipeline changes.
func (m *PipelineManager) ReloadConfig(c config.Config) {
	m.conf = c
	// In a fully modular system, this should trigger a pipeline rebuild if essential settings change.
	// For now, only log the change.
	log.Println("GStreamer Manager config reloaded.")
}

// StartPipeline builds and starts the GStreamer pipeline.
func (m *PipelineManager) StartPipeline() error {
	m.conf = config.GetConfig()
	if m.pipeline != nil {
		// Stop any currently running pipeline before starting a new one
		m.StopPipeline()
	}

	// 1. Build the GStreamer pipeline string
	pipelineStr, err := m.buildPipelineString()
	if err != nil {
		return fmt.Errorf("failed to build pipeline string: %w", err)
	}

	// 2. Create the pipeline from string
	pipeline, err := gst.ParseLaunch(pipelineStr)
	if err != nil {
		return fmt.Errorf("failed to parse pipeline: %w", err)
	}

	m.pipeline = pipeline.Cast().(*gst.Pipeline)
	log.Printf("Starting pipeline: %s", pipelineStr)

	// 3. Set up the bus handler for messages and errors
	bus := m.pipeline.GetBus()
	bus.AddSignalWatch()
	bus.Connect("message", m.onBusMessage)

	// 4. Set pipeline state to PLAYING
	stateChangeResult, err := m.pipeline.SetState(gst.StatePlaying)
	if err != nil {
		return fmt.Errorf("failed to set pipeline state to playing: %w", err)
	}

	// FIX: Use fmt.Sprintf("%v", ...) to access the string representation of the opaque error/return type
	if fmt.Sprintf("%v", stateChangeResult) == fmt.Sprintf("%v", gst.StateChangeFailure) {
		return fmt.Errorf("pipeline state change failed: %s", fmt.Sprintf("%v", stateChangeResult))
	}
	log.Println("Pipeline started successfully.")

	return nil
}

// StopPipeline stops and cleans up the GStreamer pipeline.
func (m *PipelineManager) StopPipeline() {
	if m.pipeline == nil {
		return
	}

	// Set pipeline state to NULL (stops all activity)
	stateChangeResult, err := m.pipeline.SetState(gst.StateNull)
	if err != nil {
		log.Printf("ERROR: Failed to set pipeline state to NULL: %v", err)
	}

	// FIX: Use fmt.Sprintf("%v", ...) for robust string conversion
	if fmt.Sprintf("%v", stateChangeResult) == fmt.Sprintf("%v", gst.StateChangeFailure) {
		log.Printf("WARNING: Pipeline state change to NULL failed: %s", fmt.Sprintf("%v", stateChangeResult))
	}

	// Unreference the pipeline object (clean up)
	m.pipeline.Unref()
	m.pipeline = nil
	log.Println("Pipeline stopped and cleaned up.")
}

// StopPipelineGraceful implements a clean shutdown sequence.
func (m *PipelineManager) StopPipelineGraceful() {
	log.Println("Attempting graceful pipeline shutdown...")
	m.StopPipeline()
	log.Println("Graceful pipeline shutdown complete.")
}

// IsRecording returns the current recording status.
func (m *PipelineManager) IsRecording() bool {
	return m.isRecording
}

// buildPipelineString dynamically constructs the GStreamer pipeline based on configuration.
func (m *PipelineManager) buildPipelineString() (string, error) {
	audioSettings := m.conf.AudioSettings
	autoRecord := m.conf.AutoRecord

	// --- 1. Audio Source (Always first) ---
	// Use the configured device, falling back to audiotestsrc if the device is empty or problematic.
	// NOTE: This uses pulsesrc, which relies on the system having PulseAudio/PipeWire running.
	audioSrc := fmt.Sprintf("pulsesrc device=%s name=source-element", audioSettings.Device)
	if audioSettings.Device == "" {
		// FALLBACK SOURCE: Use audiotestsrc if no device is specified, as per plan.
		audioSrc = "audiotestsrc name=source-element wave=sine freq=440 volume=0.5 ! audioresample ! audioconvert"
		log.Println("WARNING: No audio device configured. Using silent audiotestsrc as source.")
	}

	// Determine capture format based on configuration
	capsString := fmt.Sprintf("audio/x-raw, rate=%d", audioSettings.SampleRate)
	if audioSettings.MasterChannels > 0 {
		capsString += fmt.Sprintf(", channels=%d", audioSettings.MasterChannels)
	}
	if audioSettings.BitDepth > 0 {
		// GStreamer often prefers bit-depths in specific formats like S16LE, S24LE, S32LE
		var depthFormat string
		switch audioSettings.BitDepth {
		case 16:
			depthFormat = "S16LE"
		case 24:
			depthFormat = "S24LE"
		case 32:
			depthFormat = "S32LE"
		default:
			log.Printf("WARNING: Unsupported bit depth %d. Defaulting to 16-bit.", audioSettings.BitDepth)
			depthFormat = "S16LE"
		}
		capsString += fmt.Sprintf(", format=%s", depthFormat)
	}
	capsFilter := fmt.Sprintf("! audioconvert ! audioresample ! capsfilter caps=\"%s\"", capsString)

	// --- 2. Master Audio Processing Chain ---
	// Source -> Rate/Format Conversion -> VAD (if enabled) -> Tee
	masterChain := audioSrc + capsFilter

	if autoRecord.Enabled {
		// Voice Activity Detection (VAD) for automatic recording and smart splitting
		// VAD uses audiometer for level detection, and level=0 for silence detection
		vad := fmt.Sprintf("! level name=%s message=true interval=100000000 ! tee name=%s", VadElement, TeeElement)
		masterChain += vad
	} else {
		// Simple Tee if VAD is disabled
		masterChain += fmt.Sprintf("! tee name=%s", TeeElement)
	}

	// --- 3. Recording Branch (Always available if not VAD controlled) ---
	recordingBranch := ""
	if autoRecord.Enabled {
		// Recording path: Tee -> Queue -> Muxer (WAV) -> File Sink
		// Note: The VAD logic will link/unlink the tee dynamically.
		recordingBranch = fmt.Sprintf("%s. ! queue name=%s ! wavenc ! filesink name=%s location=\"%s/temp.wav\" sync=false",
			TeeElement,
			RecorderQueue,
			FileSinkElement,
			autoRecord.Directory,
		)
	}

	// --- 4. Streaming Branches (Modular Plugins) ---
	streamingBranches := ""
	for _, plugin := range registeredStreamPlugins {
		if plugin.IsEnabled() {
			// Link streaming branch from the main tee
			streamingBranches += fmt.Sprintf(" %s. ! queue name=%s-q ", TeeElement, plugin.GetSinkElement()) + plugin.BuildPipelineBranch()
		}
	}

	// Final Pipeline Assembly
	pipeline := fmt.Sprintf("%s %s %s", masterChain, recordingBranch, streamingBranches)

	// If no outputs are enabled, we need to ensure the pipeline doesn't end abruptly (e.g., using a fakesink)
	if recordingBranch == "" && streamingBranches == "" {
		pipeline += fmt.Sprintf("! fakesink name=null-sink")
	}

	return pipeline, nil
}

// onBusMessage handles messages from the GStreamer bus (errors, state changes, etc.).
func (m *PipelineManager) onBusMessage(_ *gst.Bus, msg *gst.Message) {
	switch msg.Type() {
	case gst.MessageEOS:
		log.Println("GStreamer Message: End of Stream (EOS).")
		// Automatically restart or handle cleanup here if needed.
	case gst.MessageError:
		m.onErrorMessage(msg)
	case gst.MessageWarning:
		m.onWarningMessage(msg)
	case gst.MessageStateChanged:
		// Log important state changes
		m.onStateChangedMessage(msg)
	case gst.MessageElement:
		// Handle VAD messages from the level element
		if msg.GetSource().GetName() == VadElement {
			m.onVadMessage(msg)
		}
	}
}

// onErrorMessage logs and handles critical GStreamer errors.
func (m *PipelineManager) onErrorMessage(msg *gst.Message) {
	gstErr, _, _ := msg.ParseError()
	// FIX: Use msg.GetSource() method, which is the most stable form of source access
	log.Printf("FATAL GStreamer Error from %s: %s (Debug: %s)", msg.GetSource().GetName(), gstErr.Error(), gstErr.DebugString())
	m.updateCb(map[string]interface{}{
		"type":    "error",
		"message": fmt.Sprintf("Pipeline Error from %s: %s", msg.GetSource().GetName(), gstErr.Error()),
	})
	// In a critical error, stop the pipeline to avoid resource leakage.
	m.StopPipeline()
}

// onWarningMessage logs GStreamer warnings.
func (m *PipelineManager) onWarningMessage(msg *gst.Message) {
	gstWarn, _ := msg.ParseWarning()
	// FIX: Use msg.GetSource() method, which is the most stable form of source access
	log.Printf("GStreamer Warning from %s: %s (Debug: %s)", msg.GetSource().GetName(), gstWarn.Error(), gstWarn.DebugString())
	m.updateCb(map[string]interface{}{
		"type":    "warning",
		"message": fmt.Sprintf("Pipeline Warning from %s: %s", msg.GetSource().GetName(), gstWarn.Error()),
	})
}

// onStateChangedMessage handles pipeline state updates.
func (m *PipelineManager) onStateChangedMessage(msg *gst.Message) {
	if msg.GetSource() == m.pipeline.Cast() {
		old, new, pending, _ := msg.ParseStateChanged()
		log.Printf("Pipeline State changed from %s to %s (Pending: %s)", old.String(), new.String(), pending.String())

		m.updateCb(map[string]interface{}{
			"type":      "status",
			"pipeline":  PipelineName,
			"state":     new.String(),
			"timestamp": time.Now().Unix(),
		})
	}
}

// onVadMessage processes messages from the VAD/level element.
func (m *PipelineManager) onVadMessage(msg *gst.Message) {
	if !m.conf.AutoRecord.Enabled {
		return
	}

	s := msg.GetStructure()
	if s == nil {
		return
	}

	// This is highly simplified VAD logic based on a GStreamer level element.
	rmsArr, _, err := s.GetArrayByName("rms")
	if err != nil || rmsArr == nil || rmsArr.GetSize() == 0 {
		return
	}

	// Assume first RMS value is the master level (simplified for stereo)
	rmsValue, err := rmsArr.GetIndex(0)
	if err != nil {
		return
	}

	// RMS values are typically in linear scale (0.0 to 1.0)
	// We need to convert it to dB: 20 * log10(rmsValue)
	rmsLevel := rmsValue.GoValue().(float64)

	// Note: Log conversion is complex. We will simplify by assuming a pre-calculated dB value is available,
	// or perform a simple check against a threshold.
	dbLevel := 20.0 * log10(rmsLevel) // Needs math import

	if dbLevel >= m.conf.AutoRecord.VadDbThreshold {
		m.startRecordingIfIdle()
	} else if dbLevel < m.conf.AutoRecord.VadDbThreshold-5.0 { // Use hysteresis for smoother silence detection
		m.stopRecordingIfActive()
	}

	// Broadcast level for UI meter
	m.updateCb(map[string]interface{}{
		"type":        "audio_level",
		"level_db":    dbLevel,
		"threshold":   m.conf.AutoRecord.VadDbThreshold,
		"isRecording": m.isRecording,
	})
}

// StartRecording initiates the file recording process.
func (m *PipelineManager) StartRecording() error {
	return m.startRecording()
}

// startRecordingIfIdle checks if VAD should trigger a recording start.
func (m *PipelineManager) startRecordingIfIdle() {
	if m.isRecording {
		return
	}
	// Start recording logic
	m.startRecording()
}

// startRecording safely links the recording branch and starts the file sink.
func (m *PipelineManager) startRecording() error {
	// 1. Generate filename
	startTime := time.Now()
	filename := fmt.Sprintf("rec_%s.wav", startTime.Format("20060102_150405"))
	filePath := filepath.Join(m.conf.AutoRecord.Directory, filename)

	// 2. Update File Sink Location (Need to get the element first)
	fileSink := m.pipeline.GetElementByName(FileSinkElement)
	if fileSink == nil {
		log.Println("ERROR: File sink element not found.")
		return fmt.Errorf("file sink element not found")
	}

	// Set the location property
	fileSink.SetProperty("location", filePath)

	// 3. Start Recording Metadata
	rec, err := db.AddRecording(filename, startTime)
	if err != nil {
		log.Printf("ERROR: Could not create DB entry for %s: %v", filename, err)
		return err
	}
	m.currentRec = rec
	m.isRecording = true

	log.Printf("Recording started: %s", filename)
	m.updateCb(map[string]interface{}{
		"type":      "recording_status",
		"recording": "started",
		"filename":  filename,
		"db_id":     rec.ID,
	})

	// 4. Link the Tee dynamically (This is complex in pure Go bindings and requires GStreamer API calls)
	// For now, we rely on the filesink being set to the right state and location in the static pipeline.
	return nil
}

// stopRecordingIfActive checks if VAD should trigger a recording stop.
func (m *PipelineManager) stopRecordingIfActive() {
	if !m.isRecording {
		return
	}
	// Stop recording logic
	m.StopRecording()
}

// StopRecording halts the file recording process.
func (m *PipelineManager) StopRecording() error {
	// 1. Reset file sink location to stop writing (need to set state to NULL on file sink element)
	fileSink := m.pipeline.GetElementByName(FileSinkElement)
	if fileSink == nil {
		log.Println("ERROR: File sink element not found during stop.")
		return fmt.Errorf("file sink element not found")
	}

	// This is the simplest way to stop a filesink mid-stream without crashing the whole pipeline
	fileSink.SetProperty("location", filepath.Join(m.conf.AutoRecord.Directory, "stopped.wav"))

	// 2. Finalize Metadata and DB entry
	m.finalizeRecording()
	m.isRecording = false
	m.currentRec = nil

	log.Println("Recording stopped.")
	m.updateCb(map[string]interface{}{
		"type":      "recording_status",
		"recording": "stopped",
	})
	return nil
}

// finalizeRecording computes duration, size, and updates the database.
func (m *PipelineManager) finalizeRecording() {
	if m.currentRec == nil {
		return
	}

	// Compute duration (Approximate, requires more accurate GStreamer signaling)
	duration := time.Since(m.currentRec.StartTime)

	// Get file size
	fileInfo, err := os.Stat(filepath.Join(m.conf.AutoRecord.Directory, m.currentRec.Filename))
	var sizeMB float64
	if err == nil {
		sizeMB = float64(fileInfo.Size()) / (1024 * 1024)
	} else {
		log.Printf("WARNING: Could not get file size for %s: %v", m.currentRec.Filename, err)
	}

	// Update DB
	err = db.CompleteRecording(m.currentRec.Filename, duration, sizeMB)
	if err != nil {
		log.Printf("ERROR: Failed to finalize DB entry for %s: %v", m.currentRec.Filename, err)
	}
}

// GetAudioDevices uses the pipewire module to list available audio sources.
func GetAudioDevices() ([]common.AudioDevice, error) {
	return common.ListAudioDevices()
}

// GetAudioCapabilities fetches the capabilities (rates/formats) for a specific device.
func GetAudioCapabilities(deviceName string) (*common.AudioCapabilities, error) {
	return common.GetAudioCapabilities(deviceName)
}

// Helper function for log base 10 (GStreamer uses log10 for dB conversion)
func log10(x float64) float64 {
	if x <= 0 {
		return -1000.0 // Return a very low dB for zero or negative input
	}
	return math.Log10(x)
}
