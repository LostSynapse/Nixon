// internal/gstreamer/gstreamer.go
package gstreamer

import (
	"fmt"
	"log"
	"nixon/internal/config"
	"nixon/internal/db" // Needed for recording creation
	"os"
	// "os/exec" // No longer needed for pw-dump
	"path/filepath"
	"sort" // For sorting capabilities
	"strconv"
	"strings"
	"sync"
	"time"
	// "bytes" // No longer needed for pw-dump
	// "encoding/json" // No longer needed for pw-dump

	"github.com/go-gst/go-glib/glib"
	"github.com/go-gst/go-gst/gst"
)

// PipelineManager holds the state of the GStreamer pipeline and related resources.
type PipelineManager struct {
	mu           sync.RWMutex
	pipeline     *gst.Pipeline
	mainLoop     *glib.MainLoop
	bus          *gst.Bus
	conf         config.Config // Store config as value
	updateCb     func(StatusUpdate) // Callback to notify external systems (like websocket)
	lastVadTime  time.Time
	silenceTimer *time.Timer // Timer for VAD stop delay

	// Internal state tracking (consistent with consolidated state logic)
	internalIsRecording    bool
	internalIsStreamingSrt bool
	internalIsStreamingIce bool
	internalVadActive      bool
	currentRecordingFile string
}

// StatusUpdate struct for broadcasting state changes
type StatusUpdate struct {
	IsRecording        bool      `json:"is_recording"`
	IsStreamingSrt     bool      `json:"is_streaming_srt"`
	IsStreamingIcecast bool      `json:"is_streaming_icecast"`
	VadActive          bool      `json:"vad_active"`
	LastVadTime        time.Time `json:"last_vad_time"` // Keep track of last VAD activity
	CurrentRecordingFile string `json:"current_recording_file"` // Added field
}

// Global pipeline manager (unexported)
var manager *PipelineManager
var managerMutex sync.Mutex // Mutex for manager singleton access/creation

// Initialize creates the PipelineManager singleton but doesn't start the pipeline yet.
// It takes a callback function to send status updates.
func Initialize(updateCallback func(StatusUpdate)) error {
	managerMutex.Lock()
	defer managerMutex.Unlock()
	if manager != nil {
		return fmt.Errorf("GStreamer manager already initialized")
	}

	gst.Init(nil) // Correct initialization
	log.Println("Initializing GStreamer manager...")

	cfg := config.GetConfig() // Get initial config

	manager = &PipelineManager{
		conf:     cfg, // Store as value
		updateCb: updateCallback,
	}

	return nil
}

// GetManager returns the GStreamer pipeline manager singleton.
func GetManager() *PipelineManager {
	managerMutex.Lock() // Ensure manager is accessed safely, especially during init
	defer managerMutex.Unlock()
	if manager == nil {
		log.Println("CRITICAL ERROR: GStreamer manager accessed before Initialize was called")
		return nil
	}
	return manager
}

// StartPipeline attempts to build and start the GStreamer pipeline.
// Should be called after Initialize and typically after the web server starts.
func (m *PipelineManager) StartPipeline() error {
	log.Println("Attempting to build and start GStreamer pipeline...")
	m.mu.Lock()
	defer m.mu.Unlock()

	// Ensure any previous pipeline/loop is stopped cleanly
	m.stopPipelineUnlocked()

	// --- 1. Get Current Config ---
	m.conf = config.GetConfig()

	// --- 2. Build Pipeline String ---
	pipelineStr, err := m.buildPipelineString()
	if err != nil {
		return fmt.Errorf("failed to build pipeline string: %w", err)
	}
	log.Printf("Using GStreamer pipeline: %s\n", pipelineStr)

	// --- 3. Create Pipeline ---
	m.pipeline, err = gst.NewPipelineFromString(pipelineStr)
	if err != nil {
		m.pipeline = nil
		return fmt.Errorf("failed to create pipeline: %w", err)
	}

	// --- 4. Setup Bus Watch ---
	m.bus = m.pipeline.GetBus()
	if m.bus == nil {
		m.pipeline.Unref()
		m.pipeline = nil
		return fmt.Errorf("failed to get pipeline bus")
	}
	m.bus.AddSignalWatch()
	m.bus.Connect("message::element", m.onElementMessage)
	m.bus.Connect("message::error", m.onErrorMessage)
	m.bus.Connect("message::warning", m.onWarningMessage)
	m.bus.Connect("message::eos", m.onEosMessage)

	// --- 5. Start Pipeline ---
	log.Println("Setting pipeline to PLAYING state...")
	stateChangeResult := m.pipeline.SetState(gst.StatePlaying) // Returns gst.StateChangeReturn

	// Corrected comparison: use gst.StateChangeReturn enum directly
	if stateChangeResult == gst.StateChangeFailure {
		// Corrected usage of .String() on the enum result
		log.Printf("Failed to set pipeline state to PLAYING (%s)", stateChangeResult.String())
		m.bus.RemoveSignalWatch() // Clean up bus watch
		m.pipeline.Unref()
		m.pipeline = nil
		m.bus = nil
		// Corrected usage of .String() on the enum result
		return fmt.Errorf("failed to set pipeline to playing state: %s", stateChangeResult.String())
	}
	// Corrected usage of .String() on the enum result
	log.Printf("Pipeline state change initiated successfully (Result: %s).", stateChangeResult.String())


	// --- 6. Start Main Loop ---
	m.mainLoop = glib.NewMainLoop(nil, false)
	go func() {
		log.Println("GStreamer main loop started.")
		m.mainLoop.Run() // Blocks here until Quit() is called
		log.Println("GStreamer main loop stopped.")
	}()

	log.Println("GStreamer pipeline started successfully.")
	m.notifyUpdate() // Notify initial state (likely all false)
	return nil
}

// StopPipeline safely stops the GStreamer pipeline and releases resources.
func (m *PipelineManager) StopPipeline() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopPipelineUnlocked()
}

// stopPipelineUnlocked performs the actual stopping logic (requires lock held).
func (m *PipelineManager) stopPipelineUnlocked() {
	log.Println("Stopping GStreamer pipeline...")
	if m.mainLoop != nil {
		if m.mainLoop.IsRunning() {
			m.mainLoop.Quit()
		}
		m.mainLoop = nil
	}
	if m.pipeline != nil {
		m.pipeline.SetState(gst.StateNull)
		if m.bus != nil {
			m.bus.RemoveSignalWatch() // Remove watch before unreffing pipeline
		}
		m.pipeline.Unref()
		m.pipeline = nil
	}
	if m.bus != nil {
		m.bus = nil
	}

	// Reset internal state tracking variables
	m.internalIsRecording = false
	m.internalIsStreamingSrt = false
	m.internalIsStreamingIce = false
	m.internalVadActive = false
	m.currentRecordingFile = ""

	log.Println("GStreamer pipeline stopped.")
}

// buildPipelineString constructs the pipeline string from the current config.
func (m *PipelineManager) buildPipelineString() (string, error) {

	// Validate essential config first
	if m.conf.AudioSettings.SampleRate <= 0 || m.conf.AudioSettings.BitDepth <= 0 {
		return "", fmt.Errorf("invalid audio settings: SampleRate=%d, BitDepth=%d",
			m.conf.AudioSettings.SampleRate, m.conf.AudioSettings.BitDepth)
	}

	channels := len(m.conf.AudioSettings.MasterChannels)
	if channels == 0 {
		log.Println("Warning: MasterChannels is empty, defaulting to 2 channels.")
		channels = 2
	}

	// --- Base Hardware Source ---
	// GStreamer formats map differently than simple bit depth numbers
	formatStr := fmt.Sprintf("S%dLE", m.conf.AudioSettings.BitDepth) // Assumes signed little-endian

	caps := fmt.Sprintf("audio/x-raw,format=%s,rate=%d,channels=%d",
		formatStr,
		m.conf.AudioSettings.SampleRate,
		channels,
	)
	pipeline := fmt.Sprintf(
		`pipewiresrc path="%s" ! audioconvert ! audioresample ! capsfilter caps="%s" ! tee name=master_tee `,
		m.conf.AudioSettings.Device, // Path can be empty string for default PipeWire node
		caps,
	)

	// --- Branch A: VAD (Level) ---
	pipeline += fmt.Sprintf(
		"master_tee. ! queue ! audioconvert ! audioresample ! level name=vadlevel interval=100000000 post-messages=true ! fakesink ", // interval in ns (100ms)
	)

	// --- Branch B: Pre-roll Buffer ---
	bytesPerSecond := m.conf.AudioSettings.SampleRate * (m.conf.AudioSettings.BitDepth / 8) * channels
	prerollBytes := bytesPerSecond * m.conf.AutoRecord.PrerollDuration
	if prerollBytes <= 0 {
		log.Println("Warning: Calculated pre-roll buffer size is zero or negative, defaulting to 1 second.")
		prerollBytes = bytesPerSecond
	}
	pipeline += fmt.Sprintf(
		"master_tee. ! queue name=preroll_queue max-size-buffers=0 max-size-bytes=%d max-size-time=0 ! tee name=preroll_tee ",
		prerollBytes,
	)

	// --- Dynamic Branches (fed from preroll_tee) ---

	// Recording (WAV)
	if err := os.MkdirAll(m.conf.AutoRecord.Directory, os.ModePerm); err != nil {
		log.Printf("Warning: Failed to ensure recordings directory '%s' exists: %v", m.conf.AutoRecord.Directory, err)
	}
	recEncoder := "audioresample ! audioconvert ! wavenc"
	recSinkFile := filepath.Join(m.conf.AutoRecord.Directory, "placeholder.wav")
	pipeline += fmt.Sprintf(
		`preroll_tee. ! queue ! valve name=rec_valve drop=true ! %s ! filesink name=rec_sink location="%s" async=false `,
		recEncoder,
		recSinkFile,
	)

	// SRT (Opus)
	if m.conf.SrtSettings.SrtEnabled {
		if m.conf.SrtSettings.SrtBitrate <= 0 {
			return "", fmt.Errorf("invalid SRT bitrate: %d", m.conf.SrtSettings.SrtBitrate)
		}
		srtEncoder := fmt.Sprintf("audioresample ! audioconvert ! opusenc bitrate=%d ! rtpopuspay", m.conf.SrtSettings.SrtBitrate)
		srtUri := fmt.Sprintf("srt://%s:%d?mode=listener", m.conf.SrtSettings.SrtHost, m.conf.SrtSettings.SrtPort)
		pipeline += fmt.Sprintf(
			`preroll_tee. ! queue ! valve name=srt_valve drop=true ! %s ! srtsink name=srt_sink uri="%s" wait-for-connection=false async=false `,
			srtEncoder,
			srtUri,
		)
	}

	// Icecast (MP3/LAME)
	if m.conf.IcecastSettings.IcecastEnabled {
		if m.conf.IcecastSettings.IcecastBitrate <= 0 {
			return "", fmt.Errorf("invalid Icecast bitrate: %d", m.conf.IcecastSettings.IcecastBitrate)
		}
		iceBitrate := m.conf.IcecastSettings.IcecastBitrate / 1000 // kbps
		iceEncoder := fmt.Sprintf("audioresample ! audioconvert ! lamemp3enc target=bitrate bitrate=%d quality=2 ! mpegaudiparse", iceBitrate)
		mountPoint := m.conf.IcecastSettings.IcecastMount
		if !strings.HasPrefix(mountPoint, "/") {
			mountPoint = "/" + mountPoint
		}
		pipeline += fmt.Sprintf(
			`preroll_tee. ! queue ! valve name=ice_valve drop=true ! %s ! shout2send name=ice_sink ip=%s port=%d mount=%s password=%s async=false sync=false `,
			iceEncoder,
			m.conf.IcecastSettings.IcecastHost,
			m.conf.IcecastSettings.IcecastPort,
			mountPoint,
			m.conf.IcecastSettings.IcecastPassword,
		)
	}

	return pipeline, nil
}

// --- Bus Message Handlers ---

func (m *PipelineManager) onElementMessage(_ *gst.Bus, msg *gst.Message) {
	s := msg.GetStructure()
	if s == nil {
		return
	}
	srcObjName := msg.Source() // Element name as string
	if srcObjName == "vadlevel" {
		m.handleVadMessage(s)
	}
}

func (m *PipelineManager) onErrorMessage(_ *gst.Bus, msg *gst.Message) {
	// Corrected: ParseError returns *glib.Error
	gstErr := msg.ParseError() // This returns *glib.Error
	errDetails := "unknown error"
	debugInfo := "no debug info"
	if gstErr != nil {
		errDetails = gstErr.Error()
		// Corrected: Access Domain and Code from glib.Error
		domainStr := ""
		if domain := gstErr.Domain(); domain != 0 {
			domainStr = domain.String() // Use QuarkToString internally
		}
		debugInfo = fmt.Sprintf("Domain: %s, Code: %d", domainStr, gstErr.Code())
	}
	log.Printf("GStreamer pipeline ERROR: %s", errDetails)
	log.Printf("Debug info: %s", debugInfo)
	log.Printf("Error Source Element: %s", msg.Source())
}

func (m *PipelineManager) onWarningMessage(_ *gst.Bus, msg *gst.Message) {
	// Corrected: ParseWarning returns *glib.Error
	gstWarn := msg.ParseWarning() // This returns *glib.Error
	warnDetails := "unknown warning"
	debugInfo := "no debug info"
	if gstWarn != nil {
		warnDetails = gstWarn.Error()
		// Corrected: Access Domain and Code from glib.Error
		domainStr := ""
		if domain := gstWarn.Domain(); domain != 0 {
			domainStr = domain.String()
		}
		debugInfo = fmt.Sprintf("Domain: %s, Code: %d", domainStr, gstWarn.Code())
	}
	log.Printf("GStreamer pipeline WARNING: %s", warnDetails)
	log.Printf("Debug info: %s", debugInfo)
	log.Printf("Warning Source Element: %s", msg.Source())
}

func (m *PipelineManager) onEosMessage(_ *gst.Bus, msg *gst.Message) {
	log.Printf("GStreamer pipeline EOS from element: %s", msg.Source())
	if m.pipeline != nil && msg.Source() == m.pipeline.GetName() {
		log.Println("EOS received from pipeline.")
	}
}

// --- VAD Logic ---

// handleVadMessage processes level messages for VAD
func (m *PipelineManager) handleVadMessage(s *gst.Structure) {
	val, err := s.GetValue("rms")
	if err != nil {
		log.Printf("VAD: could not get rms field: %v", err)
		return
	}
	rmsArray, ok := val.([]float64)
	if !ok || len(rmsArray) == 0 {
		log.Printf("VAD: rms field was not a []float64 or was empty")
		return
	}
	rmsDb := rmsArray[0]

	m.mu.Lock()
	defer m.mu.Unlock()

	currentVadDbThreshold := m.conf.AutoRecord.VadDbThreshold
	isAudioDetected := rmsDb > currentVadDbThreshold

	m.lastVadTime = time.Now()

	currentVadActive := m.internalVadActive
	stateChanged := false

	if isAudioDetected {
		if !currentVadActive {
			log.Printf("VAD: Voice detected (RMS: %.2f dB > Threshold: %.2f dB)", rmsDb, currentVadDbThreshold)
			m.internalVadActive = true
			stateChanged = true
			if m.silenceTimer != nil {
				log.Println("VAD: Silence timer cancelled.")
				m.silenceTimer.Stop()
				m.silenceTimer = nil
			}
			if m.conf.AutoRecord.Enabled && !m.internalIsRecording {
				log.Println("VAD: Triggering auto-record START.")
				go m.StartRecording()
				stateChanged = false // StartRecording will notify
			}
		}
	} else { // Silence detected
		if currentVadActive {
			log.Printf("VAD: Silence detected (RMS: %.2f dB <= Threshold: %.2f dB)", rmsDb, currentVadDbThreshold)
			m.internalVadActive = false
			stateChanged = true
			if m.internalIsRecording && m.silenceTimer == nil {
				timeout := time.Duration(m.conf.AutoRecord.SmartSplitTimeout) * time.Second
				if timeout <= 0 {
					log.Println("VAD: Invalid or zero SmartSplitTimeout, cannot start timer.")
				} else {
					log.Printf("VAD: Starting %s silence timer.", timeout)
					m.silenceTimer = time.AfterFunc(timeout, func() {
						log.Printf("VAD: Silence timeout (%s) reached.", timeout)
						if m.conf.AutoRecord.SmartSplitEnabled {
							log.Println("VAD: Triggering smart-split (stop).")
							m.StopRecording()
						} else {
							log.Println("VAD: Triggering auto-record STOP.")
							m.StopRecording()
						}
						m.mu.Lock() // Need lock to safely modify shared timer variable
						m.silenceTimer = nil
						m.mu.Unlock()
					})
				}
			}
		}
	}

	if stateChanged {
		m.notifyUpdate()
	}
}

// monitorSilenceTimeout (Placeholder - currently unused, logic is in AfterFunc)
func (m *PipelineManager) monitorSilenceTimeout() {}

// --- Control Functions ---

// StartRecording starts the recording branch
func (m *PipelineManager) StartRecording() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.pipeline == nil {
		log.Println("StartRecording: Pipeline not running.")
		if m.internalIsRecording {
			m.internalIsRecording = false
			m.currentRecordingFile = ""
			m.notifyUpdate()
		}
		return fmt.Errorf("pipeline not running")
	}
	if m.internalIsRecording {
		log.Println("StartRecording: Already recording.")
		return nil
	}

	log.Println("Starting recording...")

	recSinkElement, err := m.pipeline.GetElementByName("rec_sink")
	if err != nil || recSinkElement == nil {
		log.Printf("ERROR: could not get rec_sink: %v", err)
		return fmt.Errorf("could not get rec_sink: %w", err)
	}

	filename := fmt.Sprintf("rec_%s.wav", time.Now().Format("20060102_150405"))
	if err := os.MkdirAll(m.conf.AutoRecord.Directory, os.ModePerm); err != nil {
		log.Printf("Warning: Failed to ensure recordings directory '%s' exists: %v", m.conf.AutoRecord.Directory, err)
	}
	fullPath := filepath.Join(m.conf.AutoRecord.Directory, filename)
	log.Printf("Setting recording location to: %s", fullPath)
	recSinkElement.SetProperty("location", fullPath)

	recValve, err := m.pipeline.GetElementByName("rec_valve")
	if err != nil || recValve == nil {
		log.Printf("ERROR: could not get rec_valve: %v", err)
		return fmt.Errorf("could not get rec_valve: %w", err)
	}
	log.Println("Opening recording valve...")
	recValve.SetProperty("drop", false)

	recordingEntry := db.Recording{
		Filename: filename,
		Name:     filename,
	}
	// Corrected: db.AddRecording expects *db.Recording and returns error
	dbErr := db.AddRecording(&recordingEntry) // Pass pointer, expect only error
	if dbErr != nil {
		log.Printf("CRITICAL: Failed to add recording '%s' to database: %v. Closing valve.", filename, dbErr)
		recValve.SetProperty("drop", true)
		return fmt.Errorf("failed to add recording to database: %w", dbErr)
	}
	log.Printf("Recording added to database with ID: %d", recordingEntry.ID)

	m.internalIsRecording = true
	m.currentRecordingFile = filename

	log.Printf("Recording started, saving to %s", fullPath)
	m.notifyUpdate()
	return nil
}

// StopRecording stops the recording branch
func (m *PipelineManager) StopRecording() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pipelineRunning := m.pipeline != nil

	if !m.internalIsRecording {
		log.Println("StopRecording: Not currently recording (internal state).")
		if pipelineRunning {
			recValve, _ := m.pipeline.GetElementByName("rec_valve")
			if recValve != nil {
				recValve.SetProperty("drop", true)
			}
		}
		return nil
	}

	log.Println("Stopping recording...")

	if pipelineRunning {
		recValve, err := m.pipeline.GetElementByName("rec_valve")
		if err != nil || recValve == nil {
			log.Printf("ERROR: could not get rec_valve while stopping: %v", err)
		} else {
			log.Println("Closing recording valve...")
			recValve.SetProperty("drop", true)
		}
	} else {
		log.Println("StopRecording: Pipeline not running, only updating internal state.")
	}

	stoppedFile := m.currentRecordingFile
	m.internalIsRecording = false
	m.currentRecordingFile = ""

	log.Printf("Recording stopped (was saving to %s).", stoppedFile)
	m.notifyUpdate()
	return nil
}

// ToggleSrtStream toggles the SRT stream valve
func (m *PipelineManager) ToggleSrtStream(enable bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.pipeline == nil {
		if m.internalIsStreamingSrt {
			m.internalIsStreamingSrt = false
			m.notifyUpdate()
		}
		return fmt.Errorf("pipeline not running")
	}
	if !m.conf.SrtSettings.SrtEnabled {
		if m.internalIsStreamingSrt {
			m.internalIsStreamingSrt = false
			m.notifyUpdate()
		}
		return fmt.Errorf("SRT is not enabled in config")
	}

	srtValve, err := m.pipeline.GetElementByName("srt_valve")
	if err != nil || srtValve == nil {
		log.Printf("ERROR: could not get srt_valve: %v", err)
		return fmt.Errorf("could not get srt_valve: %w", err)
	}

	dropValAny, propErr := srtValve.GetProperty("drop")
	if propErr != nil {
		log.Printf("ERROR: could not get srt_valve drop property: %v", propErr)
		return fmt.Errorf("could not get srt_valve property: %w", propErr)
	}
	dropVal, ok := dropValAny.(bool)
	if !ok {
		log.Printf("ERROR: srt_valve drop property was not a boolean")
		return fmt.Errorf("srt_valve drop property type error")
	}
	isCurrentlyEnabled := !dropVal

	if enable == isCurrentlyEnabled {
		log.Printf("SRT stream already in desired state (%v).", enable)
		if m.internalIsStreamingSrt != enable {
			m.internalIsStreamingSrt = enable
			m.notifyUpdate()
		}
		return nil
	}

	log.Printf("Setting SRT valve drop property to: %v", !enable)
	srtValve.SetProperty("drop", !enable)
	m.internalIsStreamingSrt = enable

	log.Printf("SRT stream toggled to: %v", enable)
	m.notifyUpdate()
	return nil
}

// ToggleIcecastStream toggles the Icecast stream valve
func (m *PipelineManager) ToggleIcecastStream(enable bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.pipeline == nil {
		if m.internalIsStreamingIce {
			m.internalIsStreamingIce = false
			m.notifyUpdate()
		}
		return fmt.Errorf("pipeline not running")
	}
	if !m.conf.IcecastSettings.IcecastEnabled {
		if m.internalIsStreamingIce {
			m.internalIsStreamingIce = false
			m.notifyUpdate()
		}
		return fmt.Errorf("Icecast is not enabled in config")
	}

	iceValve, err := m.pipeline.GetElementByName("ice_valve")
	if err != nil || iceValve == nil {
		log.Printf("ERROR: could not get ice_valve: %v", err)
		return fmt.Errorf("could not get ice_valve: %w", err)
	}

	dropValAny, propErr := iceValve.GetProperty("drop")
	if propErr != nil {
		log.Printf("ERROR: could not get ice_valve drop property: %v", propErr)
		return fmt.Errorf("could not get ice_valve property: %w", propErr)
	}
	dropVal, ok := dropValAny.(bool)
	if !ok {
		log.Printf("ERROR: ice_valve drop property was not a boolean")
		return fmt.Errorf("ice_valve drop property type error")
	}
	isCurrentlyEnabled := !dropVal

	if enable == isCurrentlyEnabled {
		log.Printf("Icecast stream already in desired state (%v).", enable)
		if m.internalIsStreamingIce != enable {
			m.internalIsStreamingIce = enable
			m.notifyUpdate()
		}
		return nil
	}

	log.Printf("Setting Icecast valve drop property to: %v", !enable)
	iceValve.SetProperty("drop", !enable)
	m.internalIsStreamingIce = enable

	log.Printf("Icecast stream toggled to: %v", enable)
	m.notifyUpdate()
	return nil
}


// RestartPipeline attempts to tear down and rebuild the pipeline.
func (m *PipelineManager) RestartPipeline() {
	log.Println("Attempting to restart GStreamer pipeline...")
	m.StopPipeline() // Acquires lock

	time.Sleep(2 * time.Second)

	err := m.StartPipeline() // Acquires lock
	if err != nil {
		log.Printf("Failed to restart pipeline: %v. Retrying in 10s...", err)
		time.Sleep(10 * time.Second)
		go m.RestartPipeline()
	} else {
		log.Println("Pipeline restarted successfully.")
	}
}

// --- Status Getters (Thread-safe) ---

// GetStatus returns the current status of the pipeline manager.
func (m *PipelineManager) GetStatus() StatusUpdate {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return StatusUpdate{
		IsRecording:        m.internalIsRecording,
		IsStreamingSrt:     m.internalIsStreamingSrt,
		IsStreamingIcecast: m.internalIsStreamingIce,
		VadActive:          m.internalVadActive,
		LastVadTime:        m.lastVadTime,
		CurrentRecordingFile: m.currentRecordingFile,
	}
}

// --- Internal Status Setters (Used by control funcs - lock must be held by caller) ---

func (m *PipelineManager) setRecordingStatus(active bool, filename string) {
	if m.internalIsRecording != active || m.currentRecordingFile != filename {
		m.internalIsRecording = active
		m.currentRecordingFile = filename
	}
}

func (m *PipelineManager) setSrtStatus(active bool) {
	if m.internalIsStreamingSrt != active {
		m.internalIsStreamingSrt = active
	}
}

func (m *PipelineManager) setIcecastStatus(active bool) {
	if m.internalIsStreamingIce != active {
		m.internalIsStreamingIce = active
	}
}

func (m *PipelineManager) setVadStatus(active bool) {
	if m.internalVadActive != active {
		m.internalVadActive = active
	}
}


// notifyUpdate calls the registered callback function with the current status.
func (m *PipelineManager) notifyUpdate() {
	if m.updateCb != nil {
		status := m.GetStatus() // Safely get current status
		go func(s StatusUpdate) { // Pass status copy
			m.updateCb(s)
		}(status)
	}
}

// --- Audio Capabilities (Native GStreamer/PipeWire Implementation - A5) ---

// AudioCapabilities represents the parsed hardware capabilities.
type AudioCapabilities struct {
	Rates  []int `json:"rates"`
	Depths []int `json:"depths"`
	// Could add Channels []int if needed
}

// AudioDevice represents a discovered audio device.
type AudioDevice struct {
	ID          string `json:"id"`           // Unique identifier (e.g., PipeWire node name or ALSA hw:X,Y)
	Name        string `json:"name"`         // User-friendly display name
	Description string `json:"description"`  // More detailed description
	API         string `json:"api"`          // e.g., "pipewire", "alsa"
	Class       string `json:"device_class"` // e.g., "Audio/Source", "Audio/Sink"
}

// ListAudioDevices uses GStreamer's DeviceMonitor to find audio source devices.
func ListAudioDevices() ([]AudioDevice, error) {
	log.Println("Listing available audio source devices using GStreamer DeviceMonitor...")
	monitor := gst.NewDeviceMonitor()
	defer monitor.Stop() // Ensure monitor is stopped

	// Add filters for audio sources (adjust caps if needed, e.g., include sinks)
	// Using generic audio/x-raw, might need audio/L16, audio/L24 etc. if raw isn't enough
	// Filter for PipeWire specific devices if possible? GStreamer abstracts this.
	caps := gst.CapsFromString("audio/x-raw")
	if !monitor.AddFilter("Audio/Source", caps) {
		log.Println("Warning: Failed to add Audio/Source filter to device monitor.")
		// Proceed without filter? Or return error? For now, proceed.
	}

	if !monitor.Start() {
		return nil, fmt.Errorf("failed to start GStreamer device monitor")
	}

	devices := monitor.GetDevices()
	if devices == nil {
		log.Println("Device monitor returned no devices.")
		return []AudioDevice{}, nil
	}

	var audioDevices []AudioDevice
	for i := uint(0); i < devices.Length(); i++ {
		dev := devices.Nth(i).(*gst.Device) // Cast glib.List element
		if dev == nil {
			continue
		}
		props := dev.GetProperties()
		if props == nil {
			log.Printf("Warning: Could not get properties for device %d", i)
			continue
		}

		// Extract properties
		deviceName := props.GetString("device.name")       // Often the unique ID for ALSA/Pulse
		deviceDesc := props.GetString("device.description") // User-friendly name
		deviceAPI := props.GetString("device.api")         // alsa, pulse, pipewire etc.
		deviceClass := props.GetString("device.class")     // Source, Sink, etc.

		// For PipeWire, node.name might be the better ID if available
		nodeName := props.GetString("node.name")
		deviceID := deviceName // Default to device.name
		if deviceAPI == "pipewire" && nodeName != "" {
			deviceID = nodeName // Use node.name as ID for pipewiresrc path
		}

		log.Printf("Found device: ID=%s, Name=%s, Desc=%s, API=%s, Class=%s",
			deviceID, deviceName, deviceDesc, deviceAPI, deviceClass)

		// Filter only sources for input selection
		if deviceClass == "Audio/Source" {
			audioDevices = append(audioDevices, AudioDevice{
				ID:          deviceID,
				Name:        deviceDesc, // Use description as the primary Name
				Description: fmt.Sprintf("%s (%s)", deviceDesc, deviceName), // Combine for more info
				API:         deviceAPI,
				Class:       deviceClass,
			})
		}
	}

	log.Printf("Found %d audio source devices.", len(audioDevices))
	return audioDevices, nil
}


// GetAudioCapabilities gets capabilities for a specific device using native GStreamer.
// deviceID should correspond to the ID from ListAudioDevices (often node.name for PipeWire).
func GetAudioCapabilities(deviceID string) (*AudioCapabilities, error) {
	log.Printf("Getting native capabilities for device ID: '%s'", deviceID)
	monitor := gst.NewDeviceMonitor()
	defer monitor.Stop()

	// Add filters (optional, might speed up GetDevices)
	capsFilter := gst.CapsFromString("audio/x-raw")
	monitor.AddFilter("Audio/Source", capsFilter)

	if !monitor.Start() {
		return nil, fmt.Errorf("failed to start GStreamer device monitor")
	}

	devices := monitor.GetDevices()
	if devices == nil {
		return nil, fmt.Errorf("device monitor returned no devices")
	}

	var targetDevice *gst.Device
	for i := uint(0); i < devices.Length(); i++ {
		dev := devices.Nth(i).(*gst.Device)
		if dev == nil { continue }
		props := dev.GetProperties()
		if props == nil { continue }

		// Match based on the ID determined in ListAudioDevices
		nodeName := props.GetString("node.name")
		deviceName := props.GetString("device.name")
		api := props.GetString("device.api")
		currentDeviceID := deviceName
		if api == "pipewire" && nodeName != "" {
			currentDeviceID = nodeName
		}

		if currentDeviceID == deviceID {
			targetDevice = dev
			log.Printf("Found matching device: %s", deviceID)
			break
		}
	}

	if targetDevice == nil {
		return nil, fmt.Errorf("could not find device with ID '%s'", deviceID)
	}

	// --- Get Device Capabilities ---
	devCaps := targetDevice.GetCaps()
	if devCaps == nil {
		return nil, fmt.Errorf("failed to get capabilities for device '%s'", deviceID)
	}
	defer devCaps.Unref() // Ensure caps are unreferenced

	log.Printf("Raw Caps for %s:\n%s", deviceID, devCaps.String())

	rates := make(map[int]bool)
	depths := make(map[int]bool)

	// Iterate through all structures within the capabilities
	for i := uint(0); i < devCaps.GetSize(); i++ {
		structure := devCaps.GetStructure(i)
		if structure == nil { continue }

		// Check if it's audio/x-raw
		if structure.Name() != "audio/x-raw" { continue }

		// --- Extract Rate ---
		rateVal, err := structure.GetValue("rate")
		if err == nil {
			parseGstValueRates(rateVal, rates)
		} else {
			//log.Printf("Could not get 'rate' field from caps structure %d", i)
		}

		// --- Extract Format (for Bit Depth) ---
		formatVal, err := structure.GetValue("format")
		if err == nil {
			parseGstValueFormats(formatVal, depths)
		} else {
			//log.Printf("Could not get 'format' field from caps structure %d", i)
		}
	}

	// Convert maps to sorted slices
	ratesSlice := make([]int, 0, len(rates))
	for r := range rates { ratesSlice = append(ratesSlice, r) }
	sort.Ints(ratesSlice)

	depthsSlice := make([]int, 0, len(depths))
	for d := range depths { depthsSlice = append(depthsSlice, d) }
	sort.Ints(depthsSlice)

	if len(ratesSlice) == 0 || len(depthsSlice) == 0 {
		log.Printf("Warning: Parsed native capabilities are incomplete for device '%s'. Rates: %d, Depths: %d", deviceID, len(ratesSlice), len(depthsSlice))
		// Optionally add defaults if parsing completely failed
		// if len(ratesSlice)==0 { ratesSlice = []int{48000} }
		// if len(depthsSlice)==0 { depthsSlice = []int{16, 24} }
	}

	log.Printf("Parsed Native Caps for %s - Rates: %v, Depths: %v", deviceID, ratesSlice, depthsSlice)

	return &AudioCapabilities{
		Rates:  ratesSlice,
		Depths: depthsSlice,
	}, nil
}

// parseGstValueRates processes a glib.Value containing rate info (int, list, range)
func parseGstValueRates(value *glib.Value, rateMap map[int]bool) {
	if value == nil { return }

	if value.Type() == glib.TypeInt {
		rate, _ := value.GetInt()
		if rate > 0 { rateMap[rate] = true }
	} else if value.Type().Name() == "GstValueList" {
		list, err := value.GetValue() // Gets interface{} containing []interface{}
		if err == nil {
			if listArr, ok := list.([]interface{}); ok {
				for _, item := range listArr {
					// Items in list might be simple ints or other types
					if subVal, ok := item.(*glib.Value); ok { // Check if item is *glib.Value
						if subVal.Type() == glib.TypeInt {
							rate, _ := subVal.GetInt()
							if rate > 0 { rateMap[rate] = true }
						}
					} else if rate, ok := item.(int); ok { // Check if item is plain int
						 if rate > 0 { rateMap[rate] = true }
					} else if rateF, ok := item.(float64); ok { // Sometimes parsed as float
						if int(rateF) > 0 { rateMap[int(rateF)] = true}
					}
				}
			}
		}
	} else if value.Type().Name() == "GstIntRange" {
		// gst.IntRange parsing is complex with current bindings, might need CGo or simplification
		// For now, let's try getting min/max if possible (this is a guess at API)
		rangeVal, err := value.GetValue() // Returns interface{} representing the range
		if err == nil {
			// How to access min/max from interface{}? Reflection? Unsafe?
			// Let's log and skip for now
			log.Printf("Warning: GstIntRange parsing for rates not fully implemented: %v", rangeVal)
			// Heuristic: Add common rates if a range is present?
			// rateMap[44100] = true; rateMap[48000] = true
		}
	}
}

// parseGstValueFormats processes a glib.Value containing format info (string, list)
func parseGstValueFormats(value *glib.Value, depthMap map[int]bool) {
	if value == nil { return }

	if value.Type() == glib.TypeString {
		format, _ := value.GetString()
		if depth := parseGstFormatToBitDepth(format); depth > 0 { depthMap[depth] = true }
	} else if value.Type().Name() == "GstValueList" {
		list, err := value.GetValue() // Gets interface{} containing []interface{}
		if err == nil {
			if listArr, ok := list.([]interface{}); ok {
				for _, item := range listArr {
					if subVal, ok := item.(*glib.Value); ok { // Check if item is *glib.Value
						if subVal.Type() == glib.TypeString {
							format, _ := subVal.GetString()
							if depth := parseGstFormatToBitDepth(format); depth > 0 { depthMap[depth] = true }
						}
					} else if format, ok := item.(string); ok { // Check if item is plain string
						if depth := parseGstFormatToBitDepth(format); depth > 0 { depthMap[depth] = true }
					}
				}
			}
		}
	}
	// GstValueRange for formats? Unlikely.
}

// parseGstFormatToBitDepth converts GStreamer format strings (e.g., "S16LE") to bit depth.
func parseGstFormatToBitDepth(format string) int {
	format = strings.ToUpper(format) // GStreamer formats are usually uppercase
	// Look for S<bits>LE or S<bits>BE, F<bits>LE/BE etc.
	if strings.HasPrefix(format, "S") || strings.HasPrefix(format, "U") || strings.HasPrefix(format, "F") {
		numStr := ""
		isFloat := strings.HasPrefix(format, "F")
		for _, r := range format[1:] {
			if r >= '0' && r <= '9' {
				numStr += string(r)
			} else {
				break // Stop at non-digit (like L or B)
			}
		}
		if depth, err := strconv.Atoi(numStr); err == nil {
			// Only accept common depths, map floats if needed
			if isFloat && depth == 32 { return 32 } // Map F32 to 32
			if isFloat && depth == 64 { return 64 } // Map F64 to 64? Or ignore? GStreamer uses F64LE.
			if !isFloat && (depth == 8 || depth == 16 || depth == 24 || depth == 32) {
				return depth
			}
		}
	}
	return 0 // Unknown or unsupported format
}


// uniqueInts removes duplicate integers from a slice and sorts them.
func uniqueInts(slice []int) []int {
	if len(slice) == 0 {
		return slice
	}
	keys := make(map[int]bool)
	list := []int{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	sort.Ints(list) // Sort the result
	return list
}

