// internal/gstreamer/gstreamer.go
package gstreamer

import (
	"fmt"
	"log"
	"nixon/internal/config"
	"os/exec"
	"regexp"
	// "strings" // Removed unused import
	"sync"
	"time"

	"github.com/go-gst/go-glib/glib" // Correct, standard import
	"github.com/go-gst/go-gst/gst"   // Correct, standard import
)

// PipelineManager holds the state of the GStreamer pipeline
type PipelineManager struct {
	mu                 sync.RWMutex
	pipeline           *gst.Pipeline
	mainLoop           *glib.MainLoop
	bus                *gst.Bus
	conf               *config.Config
	isRecording        bool
	isStreamingSrt     bool
	isStreamingIcecast bool
	vadActive          bool
	lastVadTime        time.Time
	silenceTimer       *time.Timer
}

// Global pipeline manager
var manager *PipelineManager

// Init initializes the GStreamer pipeline manager
// This no longer starts the pipeline, allowing the server to boot
// even if GStreamer configuration is bad.
func Init(c *config.Config) error {
	gst.Init(nil)
	log.Println("Initializing GStreamer manager...")

	manager = &PipelineManager{
		conf: c,
	}
	return nil
}

// GetManager returns the global pipeline manager instance
func GetManager() *PipelineManager {
	if manager == nil {
		log.Printf("CRITICAL: GetManager() called before gstreamer.Init(). This should not happen.")
		// Attempt to recover by initializing with default config
		// This is a safeguard, main.go should always call Init first.
		Init(config.GetConfig())
	}
	return manager
}

// StartPipeline builds and starts the pipeline.
// This is now called from main.go after the server is running.
func (m *PipelineManager) StartPipeline() error {
	log.Println("Attempting to build and start GStreamer pipeline...")
	// buildAndRunPipeline acquires its own lock
	return m.buildAndRunPipeline()
}

// buildAndRunPipeline creates and starts the master pipeline
func (m *PipelineManager) buildAndRunPipeline() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop and clean up any existing pipeline
	if m.pipeline != nil {
		m.pipeline.SetState(gst.StateNull)
		m.pipeline.Unref()
	}
	if m.mainLoop != nil {
		m.mainLoop.Quit()
	}

	// --- 1. Build Pipeline String ---
	pipelineStr, err := m.buildPipelineString()
	if err != nil {
		return fmt.Errorf("failed to build pipeline string: %w", err)
	}
	log.Printf("Using GStreamer pipeline: %s\n", pipelineStr)

	// --- 2. Create Pipeline ---
	m.pipeline, err = gst.NewPipelineFromString(pipelineStr)
	if err != nil {
		return fmt.Errorf("failed to create pipeline: %w", err)
	}

	// --- 3. Setup Bus Watch ---
	m.bus = m.pipeline.GetBus()
	m.bus.AddSignalWatch()
	m.bus.Connect("message", m.onBusMessage)

	// --- 4. Start Pipeline ---
	if err = m.pipeline.SetState(gst.StatePlaying); err != nil {
		log.Printf("Failed to set pipeline to PLAYING: %v", err)
		// Clean up the failed pipeline
		m.pipeline.Unref()
		m.pipeline = nil
		return fmt.Errorf("failed to set pipeline to playing state: %w", err)
	}

	// --- 5. Start Main Loop ---
	m.mainLoop = glib.NewMainLoop(nil, false)
	go m.mainLoop.Run()

	log.Println("GStreamer pipeline started successfully.")
	return nil
}

// buildPipelineString constructs the pipeline string from config
func (m *PipelineManager) buildPipelineString() (string, error) {
	// --- Base Hardware Source ---
	channels := 2
	if len(m.conf.AudioSettings.MasterChannels) > 0 {
		channels = len(m.conf.AudioSettings.MasterChannels)
	}

	// Corrected: Use 'path' property for pipewiresrc, not 'device'
	// A "default" path might not exist, but an empty path usually works to find
	// the default configured source.
	devicePath := m.conf.AudioSettings.Device
	if devicePath == "default" {
		devicePath = "" // Use empty path for default pipewire source
	}
	pipeline := fmt.Sprintf(
		"pipewiresrc path=%s ! audioconvert ! audioresample ! capsfilter caps=\"audio/x-raw,format=S%dLE,rate=%d,channels=%d\" ! tee name=master_tee ",
		devicePath,
		m.conf.AudioSettings.BitDepth,
		m.conf.AudioSettings.SampleRate,
		channels,
	)

	// --- Branch A: VAD (Always on) ---
	// Use 'level' element for VAD, posting messages every 100ms
	pipeline += "master_tee. ! queue ! audioconvert ! audioresample ! level name=vadlevel interval=100000000 ! fakesink "

	// --- Branch B: Pre-roll Buffer (Always on) ---
	prerollBytes := m.conf.AudioSettings.SampleRate * (m.conf.AudioSettings.BitDepth / 8) * channels * m.conf.AutoRecord.PrerollDuration
	pipeline += fmt.Sprintf(
		"master_tee. ! queue name=preroll_queue max-size-buffers=0 max-size-bytes=%d max-size-time=0 ! tee name=preroll_tee ",
		prerollBytes,
	)

	// --- Dynamic Branches (fed from preroll_tee) ---
	recEncoder := "audioresample ! audioconvert ! wavenc"
	srtBitrate := m.conf.SrtSettings.SrtBitrate
	srtEncoder := fmt.Sprintf("audioresample ! audioconvert ! opusenc bitrate=%d ! rtpopuspay", srtBitrate)
	iceBitrate := m.conf.IcecastSettings.IcecastBitrate / 1000
	iceEncoder := fmt.Sprintf("audioresample ! audioconvert ! lamemp3enc bitrate=%d ! mpegaudioparse", iceBitrate)

	// --- Branch C: Recording (Dynamic) ---
	pipeline += fmt.Sprintf(
		"preroll_tee. ! queue ! valve name=rec_valve drop=true ! %s ! filesink name=rec_sink location=/tmp/placeholder.wav ",
		recEncoder,
	)

	// --- Branch D: SRT (Dynamic) ---
	if m.conf.SrtSettings.SrtEnabled {
		pipeline += fmt.Sprintf(
			"preroll_tee. ! queue ! valve name=srt_valve drop=true ! %s ! srtsink name=srt_sink uri=\"srt://%s:%d\" ",
			srtEncoder,
			m.conf.SrtSettings.SrtHost,
			m.conf.SrtSettings.SrtPort,
		)
	}

	// --- Branch E: Icecast (Dynamic) ---
	if m.conf.IcecastSettings.IcecastEnabled {
		pipeline += fmt.Sprintf(
			"preroll_tee. ! queue ! valve name=ice_valve drop=true ! %s ! shout2send name=ice_sink ip=%s port=%d mount=%s password=%s ",
			iceEncoder,
			m.conf.IcecastSettings.IcecastHost,
			m.conf.IcecastSettings.IcecastPort,
			m.conf.IcecastSettings.IcecastMount,
			m.conf.IcecastSettings.IcecastPassword,
		)
	}

	return pipeline, nil
}

// onBusMessage handles messages from the GStreamer bus
func (m *PipelineManager) onBusMessage(bus *gst.Bus, msg *gst.Message) {
	switch msg.Type() {
	case gst.MessageError:
		err := msg.ParseError()
		log.Printf("GStreamer pipeline ERROR: %s", err.Error())
		log.Printf("Debug info: %s", err.DebugString())
		go m.RestartPipeline()

	case gst.MessageWarning:
		err := msg.ParseWarning()
		log.Printf("GStreamer pipeline WARNING: %s", err.Error())

	case gst.MessageEOS:
		log.Println("GStreamer pipeline: End-Of-Stream received.")

	case gst.MessageElement:
		// Check if the message is from our 'vadlevel' element
		src := msg.Source()
		if src == "vadlevel" {
			s := msg.GetStructure()
			if s != nil && s.Name() == "level" {
				m.handleVadMessage(s)
			}
		}
	}
}

// handleVadMessage processes VAD events from the 'level' element
func (m *PipelineManager) handleVadMessage(s *gst.Structure) {
	// Corrected parsing logic for 'rms'
	// The 'rms' property is a G_TYPE_DOUBLE_ARRAY, which Go-GST
	// represents as an interface{} containing []float64.
	val, err := s.GetValue("rms")
	if err != nil {
		log.Printf("VAD: could not get rms value: %v", err)
		return
	}

	// Correct parsing: Type assert the interface{} directly to []float64
	rmsArray, ok := val.([]float64)
	if !ok {
		log.Printf("VAD: could not assert rms value to []float64 (type was %T)", val)
		return
	}

	// For VAD, we only care about the first channel (or if any channel is loud)
	if len(rmsArray) == 0 {
		log.Println("VAD: rms array was empty")
		return
	}
	rmsDb := rmsArray[0] // Use the first channel's level

	m.mu.Lock()
	defer m.mu.Unlock()

	m.lastVadTime = time.Now()
	isAudioDetected := rmsDb > m.conf.AutoRecord.VadDbThreshold

	if isAudioDetected {
		if !m.vadActive {
			log.Println("VAD: Voice detected.")
			m.vadActive = true
			// If a silence timer is running, stop it.
			if m.silenceTimer != nil {
				m.silenceTimer.Stop()
				m.silenceTimer = nil
				log.Println("VAD: Silence timer cancelled, recording continues.")
			}
			// If not recording, start.
			if m.conf.AutoRecord.Enabled && !m.isRecording {
				log.Println("VAD: Triggering auto-record START.")
				go m.StartRecording() // Start in a new goroutine
			}
		}
	} else { // Silence is detected
		if m.vadActive {
			log.Println("VAD: Silence detected.")
			m.vadActive = false
			// If we are recording and a timer isn't already running, start one.
			if m.isRecording && m.silenceTimer == nil {
				timeout := time.Duration(m.conf.AutoRecord.SmartSplitTimeout) * time.Second
				log.Printf("VAD: Starting %d second silence timer...", m.conf.AutoRecord.SmartSplitTimeout)
				m.silenceTimer = time.AfterFunc(timeout, func() {
					log.Println("VAD: Silence timeout reached, stopping auto-record.")
					// Corrected: Call StopRecording in a new goroutine
					// to avoid deadlocking the VAD monitor.
					go m.StopRecording()
				})
			}
		}
	}
}

// StartVadMonitor checks for VAD timeouts (Now handled by handleVadMessage)
// This function can be simplified or removed if VAD logic is purely message-based
// For now, it will just start the ticker.
func (m *PipelineManager) StartVadMonitor() {
	// The core logic is now in handleVadMessage.
	// We can leave this function empty or use it for other monitoring.
	log.Println("VAD message handler is now active.")
}

// StartRecording starts the recording branch
func (m *PipelineManager) StartRecording() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isRecording {
		log.Println("StartRecording called but recording is already active.")
		return nil
	}

	log.Println("Starting recording...")
	if m.pipeline == nil {
		return fmt.Errorf("cannot start recording, pipeline is nil")
	}

	// 1. Set new file location on filesink
	filename := fmt.Sprintf("%s/rec_%s.wav", m.conf.AutoRecord.Directory, time.Now().Format("20060102_150405"))
	recSink, err := m.pipeline.GetElementByName("rec_sink")
	if err != nil {
		log.Printf("ERROR: could not get rec_sink: %v", err)
		return fmt.Errorf("could not get rec_sink: %w", err)
	}
	recSink.SetProperty("location", filename)

	// 2. Open the valve
	recValve, err := m.pipeline.GetElementByName("rec_valve")
	if err != nil {
		log.Printf("ERROR: could not get rec_valve: %v", err)
		return fmt.Errorf("could not get rec_valve: %w", err)
	}
	recValve.SetProperty("drop", false)

	log.Printf("Recording started, saving to %s", filename)
	m.isRecording = true
	return nil
}

// StopRecording stops the recording branch
func (m *PipelineManager) StopRecording() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isRecording {
		log.Println("StopRecording called but recording is not active.")
		return nil
	}
	log.Println("Stopping recording...")
	if m.pipeline == nil {
		return fmt.Errorf("cannot stop recording, pipeline is nil")
	}

	// 1. Close the valve
	recValve, err := m.pipeline.GetElementByName("rec_valve")
	if err != nil {
		log.Printf("ERROR: could not get rec_valve: %v", err)
		return fmt.Errorf("could not get rec_valve: %w", err)
	}
	recValve.SetProperty("drop", true)

	// 2. Clear the silence timer
	if m.silenceTimer != nil {
		m.silenceTimer.Stop()
		m.silenceTimer = nil
	}

	log.Println("Recording stopped.")
	m.isRecording = false
	return nil
}

// ToggleSrtStream toggles the SRT stream
func (m *PipelineManager) ToggleSrtStream(enable bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.conf.SrtSettings.SrtEnabled {
		return fmt.Errorf("SRT is not enabled in config")
	}
	if m.pipeline == nil {
		return fmt.Errorf("cannot toggle SRT, pipeline is nil")
	}

	srtValve, err := m.pipeline.GetElementByName("srt_valve")
	if err != nil {
		log.Printf("ERROR: could not get srt_valve: %v", err)
		return fmt.Errorf("could not get srt_valve: %w", err)
	}

	srtValve.SetProperty("drop", !enable)
	m.isStreamingSrt = enable
	log.Printf("SRT stream set to: %v", enable)
	return nil
}

// ToggleIcecastStream toggles the Icecast stream
func (m *PipelineManager) ToggleIcecastStream(enable bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.conf.IcecastSettings.IcecastEnabled {
		return fmt.Errorf("Icecast is not enabled in config")
	}
	if m.pipeline == nil {
		return fmt.Errorf("cannot toggle Icecast, pipeline is nil")
	}

	iceValve, err := m.pipeline.GetElementByName("ice_valve")
	if err != nil {
		log.Printf("ERROR: could not get ice_valve: %v", err)
		return fmt.Errorf("could not get ice_valve: %w", err)
	}

	// Update Icecast sink properties (in case they changed)
	iceSink, err := m.pipeline.GetElementByName("ice_sink")
	if err != nil {
		log.Printf("ERROR: could not get ice_sink: %v", err)
		return fmt.Errorf("could not get ice_sink: %w", err)
	}
	iceSink.SetProperty("ip", m.conf.IcecastSettings.IcecastHost)
	iceSink.SetProperty("port", m.conf.IcecastSettings.IcecastPort)
	iceSink.SetProperty("mount", m.conf.IcecastSettings.IcecastMount)
	iceSink.SetProperty("password", m.conf.IcecastSettings.IcecastPassword)

	iceValve.SetProperty("drop", !enable)
	m.isStreamingIcecast = enable
	log.Printf("Icecast stream set to: %v", enable)
	return nil
}

// RestartPipeline attempts to tear down and rebuild the pipeline
func (m *PipelineManager) RestartPipeline() {
	log.Println("Attempting to restart GStreamer pipeline...")
	// This function acquires the lock, so it's safe to call from anywhere
	err := m.buildAndRunPipeline()
	if err != nil {
		log.Printf("Failed to restart pipeline: %v. Retrying in 10s...", err)
		time.Sleep(10 * time.Second)
		go m.RestartPipeline() // Recursive retry
	} else {
		log.Println("Pipeline restarted successfully.")
	}
}

// GetStatus returns the current status of the pipeline
func (m *PipelineManager) GetStatus() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return map[string]interface{}{
		"is_recording":         m.isRecording,
		"is_streaming_srt":     m.isStreamingSrt,
		"is_streaming_icecast": m.isStreamingIcecast,
		"vad_active":           m.vadActive,
		"last_vad_time":        m.lastVadTime,
	}
}

// GetAudioCapabilities parses hardware capabilities from pw-dump
func GetAudioCapabilities(device string) (map[string][]string, error) {
	// Use pw-dump to get system capabilities. This is less precise
	// than arecord, as pw-dump doesn't easily filter by device.
	// We will parse all available rates and formats.
	cmd := exec.Command("pw-dump")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("pw-dump command failed: %s\nOutput: %s", err, string(out))
		return nil, fmt.Errorf("failed to execute pw-dump: %s. Output: %s", err, string(out))
	}

	caps := make(map[string][]string)
	rates := make(map[string]bool)
	depths := make(map[string]bool)

	// Regex for available sample rates
	rateRegex := regexp.MustCompile(`"audio.rate":\s*(\d+)`)
	rateMatches := rateRegex.FindAllStringSubmatch(string(out), -1)
	for _, match := range rateMatches {
		if len(match) > 1 {
			rates[match[1]] = true
		}
	}

	// Regex for available formats (bit depth)
	formatRegex := regexp.MustCompile(`"audio.format":\s*"S(\d+)_?LE"`)
	formatMatches := formatRegex.FindAllStringSubmatch(string(out), -1)
	for _, match := range formatMatches {
		if len(match) > 1 {
			depths[match[1]] = true
		}
	}

	// Convert maps to slices
	for rate := range rates {
		caps["rates"] = append(caps["rates"], rate)
	}
	// Corrected line: append to caps["depths"] slice
	for depth := range depths {
		caps["depths"] = append(caps["depths"], depth)
	}

	// Add defaults if pw-dump finds nothing (e.g., in a minimal container)
	if len(caps["rates"]) == 0 {
		caps["rates"] = []string{"44100", "48000"}
	}
	if len(caps["depths"]) == 0 {
		caps["depths"] = []string{"16", "24", "32"}
	}

	return caps, nil
}

