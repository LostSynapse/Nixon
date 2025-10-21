package gstreamer

import (
	"bufio"
	"fmt"
	"log"
	"nixon/internal/config"
	"nixon/internal/db"
	"nixon/internal/state"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type ProcessType int

const (
	SRT ProcessType = iota
	Icecast
	Recording
	VAD
)

var (
	srtProcess       *exec.Cmd
	icecastProcess   *exec.Cmd
	recordingProcess *exec.Cmd
	vadProcess       *exec.Cmd
	silenceTimer     *time.Timer
	silenceTimerMutex sync.Mutex
)

func BuildSRTCommand(device string, bitrate int) *exec.Cmd {
	return exec.Command("gst-launch-1.0", "-v", "alsasrc", "device="+device, "!", "audioconvert", "!", "audioresample", "!", "queue", "!", "opusenc", "bitrate="+strconv.Itoa(bitrate), "!", "mpegtsmux", "!", "srtsink", "uri=srt://:8888?mode=listener", "latency=200")
}

func BuildIcecastCommand(audioCfg config.AudioSettings, icecastCfg config.IcecastSettings) *exec.Cmd {
	return exec.Command("gst-launch-1.0", "-v", "alsasrc", "device="+audioCfg.Device, "!", "audioconvert", "!", "audioresample", "!", "queue", "!", "vorbisenc", "bitrate="+strconv.Itoa(audioCfg.Bitrate), "!", "oggmux", "!", "shout2send", "ip="+icecastCfg.URL, "port="+icecastCfg.Port, "mount="+icecastCfg.Mount, "password="+icecastCfg.Password, "streamname="+icecastCfg.StreamName, "genre="+icecastCfg.StreamGenre, "description="+icecastCfg.StreamDescription)
}

func BuildRecordCommand(device string, filename string) *exec.Cmd {
	filepath := fmt.Sprintf("%s/%s", config.RecordingsDir, filename)
	log.Println("Starting new recording:", filepath)
	return exec.Command("gst-launch-1.0", "-v", "alsasrc", "device="+device, "!", "audioconvert", "!", "audioresample", "!", "queue", "!", "flacenc", "!", "filesink", "location="+filepath)
}

func StartProcess(pType ProcessType, cmd *exec.Cmd) error {
	var process **exec.Cmd
	switch pType {
	case SRT:
		process = &srtProcess
	case Icecast:
		process = &icecastProcess
	case Recording:
		process = &recordingProcess
	case VAD:
		process = &vadProcess
	default:
		return fmt.Errorf("unknown process type")
	}

	*process = cmd
	err := (*process).Start()
	if err != nil {
		log.Printf("Failed to start process: %v", err)
		return fmt.Errorf("failed to start process: %w", err)
	}
	return nil
}

func StopProcess(pType ProcessType) error {
	var process **exec.Cmd
	switch pType {
	case SRT:
		process = &srtProcess
	case Icecast:
		process = &icecastProcess
	case Recording:
		process = &recordingProcess
	case VAD:
		process = &vadProcess
	default:
		return fmt.Errorf("unknown process type")
	}

	if *process != nil && (*process).Process != nil {
		if err := (*process).Process.Signal(syscall.SIGTERM); err != nil {
			if err.Error() != "os: process already finished" {
				log.Printf("Failed to terminate process: %v", err)
				if errKill := (*process).Process.Kill(); errKill != nil {
					log.Printf("Failed to kill process: %v", errKill)
					return fmt.Errorf("failed to kill process: %w", errKill)
				}
			}
		}
		(*process).Wait()
	}
	*process = nil
	return nil
}

func TriggerRecording(action string) {
	s := state.Get()
	startNew := func(prefix string) {
		cfg := config.Get()
		filename := fmt.Sprintf("%s_%s.flac", prefix, time.Now().Format("20060102_150405"))
		cmd := BuildRecordCommand(cfg.Audio.Device, filename)

		if err := StartProcess(Recording, cmd); err != nil {
			log.Printf("Error starting new recording: %v", err)
			return
		}
		if _, err := db.AddRecording(filename); err != nil {
			log.Printf("CRITICAL: Failed to add new recording %s to database: %v", filename, err)
			StopProcess(Recording)
			return
		}
		state.SetRecordingState(true, filename)
	}

	stopCurrent := func() {
		if err := StopProcess(Recording); err != nil {
			log.Printf("Error stopping recording: %v", err)
			return
		}
		state.SetRecordingState(false, "")
	}

	switch action {
	case "start", "vad":
		if !s.RecordingActive {
			startNew("rec")
		}
	case "srt":
		if !s.RecordingActive {
			startNew("srt")
		}
	case "icecast":
		if !s.RecordingActive {
			startNew("icecast")
		}
	case "stop":
		if s.RecordingActive {
			stopCurrent()
		}
	case "split":
		if s.RecordingActive {
			stopCurrent()
			time.Sleep(200 * time.Millisecond)
			startNew("rec")
		}
	}
}

func MonitorVAD() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        cfg := config.Get()
        shouldBeRunning := cfg.AutoRecord.Enabled
        
        isRunning := vadProcess != nil && vadProcess.Process != nil && vadProcess.ProcessState == nil

        if shouldBeRunning && !isRunning {
            log.Println("Auto-record enabled, starting VAD pipeline...")
            vadProcess = exec.Command("gst-launch-1.0", "-m", "alsasrc", "device="+cfg.Audio.Device, "!", "audioconvert", "!", "gvad", "!", "fakesink")
            
            stderr, err := vadProcess.StderrPipe()
            if err != nil {
                log.Printf("VAD stderr pipe error: %v.", err)
                vadProcess = nil 
                continue
            }
            if err := vadProcess.Start(); err != nil {
                log.Printf("VAD process start error: %v.", err)
                vadProcess = nil 
                continue
            }

            go func() {
                scanner := bufio.NewScanner(stderr)
                for scanner.Scan() {
                    line := scanner.Text()
                    
                    cfgInLoop := config.Get()
                    if !cfgInLoop.AutoRecord.Enabled {
                        continue
                    }
                    
                    if strings.Contains(line, "gvad-start") {
						silenceTimerMutex.Lock()
						if silenceTimer != nil {
							silenceTimer.Stop()
							silenceTimer = nil
							log.Println("VAD: Silence interrupted, recording continues.")
						} else {
							s := state.Get()
							if !s.RecordingActive {
								log.Println("VAD: Audio detected, starting auto-record.")
								TriggerRecording("vad")
							}
						}
						silenceTimerMutex.Unlock()
					} else if strings.Contains(line, "gvad-stop") {
						silenceTimerMutex.Lock()
						s := state.Get()
						if s.RecordingActive && silenceTimer == nil {
							timeout := cfgInLoop.AutoRecord.TimeoutSeconds
							log.Printf("VAD: Silence detected, starting %d second stop timer.", timeout)
							silenceTimer = time.AfterFunc(time.Duration(timeout)*time.Second, func() {
								log.Println("VAD: Silence timeout reached, stopping auto-record.")
								TriggerRecording("stop")
								silenceTimerMutex.Lock()
								silenceTimer = nil
								silenceTimerMutex.Unlock()
							})
						}
						silenceTimerMutex.Unlock()
					}
                }
            }()
            
            go func() {
                vadProcess.Wait()
                log.Printf("VAD pipeline stopped. Will be restarted if enabled.")
                vadProcess = nil 
            }()

        } else if !shouldBeRunning && isRunning {
            log.Println("Auto-record disabled, stopping VAD pipeline...")
            StopProcess(VAD) 
        }
    }
}

