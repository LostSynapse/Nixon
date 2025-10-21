package state

import "sync"

// AppStateStruct is the exported type for the application state.
type AppStateStruct struct {
	sync.RWMutex
	SRTStreamActive      bool
	IcecastStreamActive  bool
	RecordingActive      bool
	CurrentRecordingFile string
	DiskUsagePercent     int
	Listeners            int
	ListenerPeak         int
}

var AppState AppStateStruct

func Get() AppStateStruct {
	AppState.RLock()
	defer AppState.RUnlock()
	return AppState
}

func SetSRTStatus(active bool) {
	AppState.Lock()
	defer AppState.Unlock()
	AppState.SRTStreamActive = active
}

func SetIcecastStatus(active bool) {
	AppState.Lock()
	defer AppState.Unlock()
	AppState.IcecastStreamActive = active
}

func SetRecordingState(active bool, file string) {
	AppState.Lock()
	defer AppState.Unlock()
	AppState.RecordingActive = active
	AppState.CurrentRecordingFile = file
}

func SetDiskUsage(percent int) bool {
	AppState.Lock()
	defer AppState.Unlock()
	if AppState.DiskUsagePercent != percent {
		AppState.DiskUsagePercent = percent
		return true
	}
	return false
}

func SetListeners(current, peak int) bool {
	AppState.Lock()
	defer AppState.Unlock()
	changed := false
	if peak > AppState.ListenerPeak {
		AppState.ListenerPeak = peak
		changed = true
	}
	if AppState.Listeners != current {
		AppState.Listeners = current
		changed = true
	}
	return changed
}

