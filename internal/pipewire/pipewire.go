// internal/pipewire/pipewire.go
// This module provides native, robust access to PipeWire device information
// via the system's DBus API, eliminating reliance on unstable go-gst device
// enumeration methods.

package pipewire

import (
	"fmt"
	"log"
	"nixon/internal/common" // NEW: Import common structs
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

// ListAudioDevices queries the PipeWire system using 'pw-cli' (as a robust,
// low-dependency standard tool for appliance environments) to list all available
// audio source nodes and returns them in a format compatible with the API.
func ListAudioDevices() ([]common.AudioDevice, error) {
	log.Println("Listing audio source devices via native 'pw-cli'...")

	// Use 'pw-cli dump short Node' to get a concise list of all PipeWire nodes.
	cmd := exec.Command("pw-cli", "dump", "short", "Node")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("ERROR: Failed to execute pw-cli dump: %v", err)
		return nil, fmt.Errorf("could not execute 'pw-cli': %w, output: %s", err, output)
	}

	lines := strings.Split(string(output), "\n")
	var audioDevices []common.AudioDevice
	
	for _, line := range lines {
		// Example line format: "id 30 name.item=alsa_input.usb-Focusrite_Focusrite_USB_Audio-00.analog-stereo type=Audio/Source"
		
		if !strings.Contains(line, "type=Audio/Source") {
			continue // Only interested in audio input sources
		}

		parts := make(map[string]string)
		
		// Simple, robust parsing logic
		fields := strings.Fields(line)
		for _, field := range fields {
			if kv := strings.SplitN(field, "=", 2); len(kv) == 2 {
				parts[kv[0]] = kv[1]
			}
		}

		// Ensure we have necessary identifying information
		_, hasID := parts["id"] // FIXED: Changed nodeIDStr to blank identifier '_'
		nodeName, hasName := parts["name.item"]
		nodeDesc, hasDesc := parts["node.description"]
		
		if !hasID || !hasName || !hasDesc {
			// Skip nodes without necessary metadata
			continue
		}

		// The ID for pipewiresrc is the node's path/name, not the runtime ID
		// Use the descriptive name for the UI
		audioDevices = append(audioDevices, common.AudioDevice{
			ID:          nodeName, // PipeWire node name is the stable path for pipewiresrc
			Name:        nodeDesc, // Descriptive name for user interface
			Description: nodeDesc,
			API:         "pipewire",
			Class:       "Audio/Source",
		})
	}
	
	log.Printf("Discovered %d PipeWire audio sources.", len(audioDevices))
	return audioDevices, nil
}

// GetAudioCapabilities queries PipeWire using 'pw-cli info' for a specific node to determine
// its supported sample rates and bit depths (formats).
func GetAudioCapabilities(nodeName string) (*common.AudioCapabilities, error) {
	log.Printf("Querying capabilities for PipeWire node: %s", nodeName)
	
	// Use 'pw-cli info <ID>' where ID is the node's path/name
	cmd := exec.Command("pw-cli", "info", nodeName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("could not execute 'pw-cli info %s': %w, output: %s", nodeName, err, output)
	}
	
	// Parse the raw JSON output from 'pw-cli info'

	rates := make(map[int]bool)
	depths := make(map[int]bool)
	
	outputStr := string(output)
	
	// Search for standard capability fields in the JSON output

	// --- Format/Depth Parsing ---
	// PipeWire usually provides formats in a list of strings
	formatLines := regexpFindAllString(outputStr, `"format":`)
	for _, f := range formatLines {
		if depth := parseFormatToBitDepth(f); depth > 0 {
			depths[depth] = true
		}
	}
	
	// --- Rate Parsing ---
	rateLines := regexpFindAllString(outputStr, `"rate":`)
	for _, rStr := range rateLines {
		if rate, err := strconv.Atoi(rStr); err == nil && rate > 0 {
			rates[rate] = true
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
		log.Printf("Warning: Capabilities query for %s returned incomplete data.", nodeName)
		// Fallback to common standard if nothing found
		if len(ratesSlice) == 0 { ratesSlice = []int{48000} }
		if len(depthsSlice) == 0 { depthsSlice = []int{16, 24} }
	}

	log.Printf("Capabilities for %s - Rates: %v, Depths: %v", nodeName, ratesSlice, depthsSlice)

	return &common.AudioCapabilities{
		Rates:  ratesSlice,
		Depths: depthsSlice,
	}, nil
}

// Helper for simplified regex matching (avoids importing "regexp" globally)
func regexpFindAllString(s, key string) []string {
    // Naive parsing for format/rate: Find key, extract value between quotes/commas
	var results []string
	
	// Find the value associated with the key (e.g., "format": or "rate":)
	index := 0
	for {
		i := strings.Index(s[index:], key)
		if i == -1 {
			break
		}
		
		currentStart := index + i + len(key)
		// Find start of value: skip whitespace/colon
		start := currentStart
		for start < len(s) && (s[start] == ' ' || s[start] == ':') {
			start++
		}
		
		// If value starts with quote (string/format)
		if start < len(s) && s[start] == '"' {
			start++
			if end := strings.Index(s[start:], `"`); end != -1 {
				results = append(results, s[start:start+end])
				index = start + end + 1 // Continue search after closing quote
				continue
			}
		} 
		
		// If value is a number (rate)
		end := start
		for end < len(s) && (s[end] >= '0' && s[end] <= '9') {
			end++
		}
		if end > start {
			results = append(results, s[start:end])
			index = end
			continue
		}
		
		// Fallback to continue search
		index = currentStart + 1
		if index >= len(s) {
			break
		}
	}
	return results
}

// parseFormatToBitDepth converts PipeWire/GStreamer format strings (e.g., "S16LE") to bit depth.
func parseFormatToBitDepth(format string) int {
	format = strings.ToUpper(format)
	// Simplified parsing logic from gstreamer.go
	if strings.HasPrefix(format, "S") || strings.HasPrefix(format, "U") || strings.HasPrefix(format, "F") {
		numStr := ""
		isFloat := strings.HasPrefix(format, "F")
		for _, r := range format[1:] {
			if r >= '0' && r <= '9' {
				numStr += string(r)
			} else {
				break
			}
		}
		if depth, err := strconv.Atoi(numStr); err == nil {
			if isFloat && (depth == 32 || depth == 64) { return 32 } // Map floats to 32 bit for config ease
			if !isFloat && (depth == 8 || depth == 16 || depth == 24 || depth == 32) {
				return depth
			}
		}
	}
	return 0
}
