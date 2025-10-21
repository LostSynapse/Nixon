package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"nixon/internal/config"
	"nixon/internal/state"
	"os"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

func monitorDiskUsage() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for ; ; <-ticker.C {
		updateDiskUsage()
	}
}

func updateDiskUsage() {
	var stat unix.Statfs_t
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting working directory for disk usage: %v", err)
		return
	}
	if err := unix.Statfs(wd, &stat); err != nil {
		log.Printf("Error getting disk stats: %v", err)
		return
	}

	totalBytes := float64(stat.Blocks) * float64(stat.Bsize)
	freeBytes := float64(stat.Bfree) * float64(stat.Bsize)
	usedBytes := totalBytes - freeBytes
	percent := 0
	if totalBytes > 0 {
		percent = int((usedBytes / totalBytes) * 100)
	}

	if state.SetDiskUsage(percent) {
		log.Printf("Disk usage updated: %d%%", percent)
	}
}

func monitorIcecastListeners() {
	type IcecastStats struct {
		Icestats struct {
			Source interface{} `json:"source"`
		} `json:"icestats"`
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for ; ; <-ticker.C {
		cfg := config.Get()
		s := state.Get()

		if !s.IcecastStreamActive {
			if state.SetListeners(0, 0) {
				// State changed, broadcast will be handled by poller
			}
			continue
		}

		statusURL := fmt.Sprintf("http://%s:%s/status-json.xsl", cfg.Icecast.URL, cfg.Icecast.Port)
		if cfg.Icecast.ServerType == "icecast-kh" {
			statusURL = fmt.Sprintf("http://%s:%s/stats", cfg.Icecast.URL, cfg.Icecast.Port)
		}

		resp, err := http.Get(statusURL)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		var stats IcecastStats
		if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
			continue
		}

		var currentListeners, currentPeak int
		mountWithSlash := "/" + strings.TrimPrefix(cfg.Icecast.Mount, "/")
		foundSource := false

		processSource := func(sourceData map[string]interface{}) (int, int, bool) {
			listenURL, _ := sourceData["listenurl"].(string)
			if listenURL != "" && strings.HasSuffix(listenURL, mountWithSlash) {
				listeners, peak := 0.0, 0.0
				if l, ok := sourceData["listeners"].(float64); ok {
					listeners = l
				}
				if p, ok := sourceData["peak_listeners"].(float64); ok {
					peak = p
				} else if p, ok := sourceData["listener_peak"].(float64); ok {
					peak = p
				}
				return int(listeners), int(peak), true
			}
			return 0, 0, false
		}

		switch src := stats.Icestats.Source.(type) {
		case map[string]interface{}:
			l, p, found := processSource(src)
			if found {
				currentListeners, currentPeak = l, p
				foundSource = true
			}
		case []interface{}:
			for _, sourceItem := range src {
				if sourceMap, ok := sourceItem.(map[string]interface{}); ok {
					l, p, found := processSource(sourceMap)
					if found {
						currentListeners, currentPeak = l, p
						foundSource = true
						break
					}
				}
			}
		}

		if !foundSource {
			log.Printf("Did not find specified mount point '%s' in Icecast status JSON from %s", mountWithSlash, statusURL)
		}

		if state.SetListeners(currentListeners, currentPeak) {
			log.Printf("Icecast listeners updated: Current=%d, Peak=%d", currentListeners, state.Get().ListenerPeak)
		}
	}
}

