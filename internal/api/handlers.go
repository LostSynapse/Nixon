package api

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"nixon/internal/config"
	"nixon/internal/db"
	"nixon/internal/gstreamer"
	"nixon/internal/state"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Recording struct for API responses
type Recording struct {
	ID        int64     `json:"id"`
	Filename  string    `json:"filename"`
	CreatedAt time.Time `json:"createdAt"`
	Name      string    `json:"name"`
	Notes     *string   `json:"notes"`
	Genre     *string   `json:"genre"`
	Protected bool      `json:"protected"`
}

// AudioDevice struct for API responses
type AudioDevice struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func getStatus(c *gin.Context) {
	s := state.Get()
	c.JSON(http.StatusOK, s)
}

func handleGetRecordings(c *gin.Context) {
	rows, err := db.DB.Query("SELECT id, filename, created_at, name, notes, genre, protected FROM recordings ORDER BY created_at DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query recordings"})
		return
	}
	defer rows.Close()

	recordings := []Recording{}
	for rows.Next() {
		var rec Recording
		if err := rows.Scan(&rec.ID, &rec.Filename, &rec.CreatedAt, &rec.Name, &rec.Notes, &rec.Genre, &rec.Protected); err != nil {
			log.Printf("Error scanning recording row: %v", err)
			continue
		}
		recordings = append(recordings, rec)
	}
	c.JSON(http.StatusOK, recordings)
}

func handleUpdateRecording(c *gin.Context) {
	id := c.Param("id")
	var rec Recording
	if err := c.ShouldBindJSON(&rec); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}
	stmt, err := db.DB.Prepare("UPDATE recordings SET name = ?, notes = ?, genre = ? WHERE id = ?")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare update"})
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(rec.Name, rec.Notes, rec.Genre, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute update"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func handleToggleProtect(c *gin.Context) {
	id := c.Param("id")
	stmt, err := db.DB.Prepare("UPDATE recordings SET protected = NOT protected WHERE id = ?")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare update"})
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute update"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func handleDeleteRecording(c *gin.Context) {
	id := c.Param("id")
	var filename string
	var protected bool
	err := db.DB.QueryRow("SELECT filename, protected FROM recordings WHERE id = ?", id).Scan(&filename, &protected)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recording not found"})
		return
	}
	if protected {
		c.JSON(http.StatusForbidden, gin.H{"error": "Recording is protected"})
		return
	}

	if err := os.Remove(filepath.Join(config.RecordingsDir, filename)); err != nil {
		log.Printf("Failed to delete file %s: %v", filename, err)
	}

	stmt, err := db.DB.Prepare("DELETE FROM recordings WHERE id = ?")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare delete statement"})
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete record from database"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func getFullConfig(c *gin.Context) {
	cfg := config.Get()
	c.JSON(http.StatusOK, cfg)
}

func updateIcecastSettings(c *gin.Context) {
	var newSettings config.IcecastSettings
	if err := c.ShouldBindJSON(&newSettings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}
	if err := config.UpdateIcecast(newSettings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save config"})
		return
	}
	log.Printf("Icecast settings updated")
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func updateSystemSettings(c *gin.Context) {
	var settings struct {
		SRTEnabled     bool                  `json:"srt_enabled"`
		IcecastEnabled bool                  `json:"icecast_enabled"`
		AutoRecord     config.AutoRecordSettings `json:"auto_record"`
	}
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}
	if err := config.UpdateSystem(settings.SRTEnabled, settings.IcecastEnabled, settings.AutoRecord); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save config"})
		return
	}
	log.Println("System settings updated")
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func updateAudioSettings(c *gin.Context) {
	var settings config.AudioSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}
	if err := config.UpdateAudio(settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save config"})
		return
	}
	log.Println("Audio settings updated")
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func handleGetAudioDevices(c *gin.Context) {
	cmd := exec.Command("aplay", "-l")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Printf("Error running 'aplay -l': %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not list audio devices"})
		return
	}

	re := regexp.MustCompile(`card (\d+): ([a-zA-Z0-9\s_]+) \[(.*)\], device (\d+):`)
	matches := re.FindAllStringSubmatch(out.String(), -1)

	devices := []AudioDevice{}
	for _, match := range matches {
		if len(match) == 5 {
			cardNum, cardId, cardName, devNum := match[1], match[2], match[3], match[4]
			id := fmt.Sprintf("hw:CARD=%s,DEV=%s", cardId, devNum)
			if cardId == cardNum {
				id = fmt.Sprintf("hw:%s,%s", cardNum, devNum)
			}
			devices = append(devices, AudioDevice{ID: id, Name: cardName})
		}
	}
	c.JSON(http.StatusOK, devices)
}

func handleSRTStream(c *gin.Context) {
	action := c.Param("action")
	switch action {
	case "start":
		s := state.Get()
		if !s.SRTStreamActive {
			state.SetSRTStatus(true)
			cfg := config.Get()
			if cfg.AutoRecord.Enabled {
				gstreamer.TriggerRecording("srt")
			}
		}
	case "stop":
		s := state.Get()
		if s.SRTStreamActive {
			state.SetSRTStatus(false)
			if s.RecordingActive && strings.HasPrefix(s.CurrentRecordingFile, "srt") {
				gstreamer.TriggerRecording("stop")
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func handleIcecastStream(c *gin.Context) {
	action := c.Param("action")
	switch action {
	case "start":
		s := state.Get()
		if !s.IcecastStreamActive {
			state.SetIcecastStatus(true)
			cfg := config.Get()
			if cfg.AutoRecord.Enabled {
				gstreamer.TriggerRecording("icecast")
			}
		}
	case "stop":
		s := state.Get()
		if s.IcecastStreamActive {
			state.SetIcecastStatus(false)
			if s.RecordingActive && strings.HasPrefix(s.CurrentRecordingFile, "icecast") {
				gstreamer.TriggerRecording("stop")
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func handleAllStreams(c *gin.Context) {
	action := c.Param("action")
	cfg := config.Get()

	switch action {
	case "start":
		if cfg.SRTEnabled {
			handleSRTStream(c)
		}
		if cfg.IcecastEnabled {
			handleIcecastStream(c)
		}
	case "stop":
		handleSRTStream(c)
		handleIcecastStream(c)
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func handleRecording(c *gin.Context) {
	action := c.Param("action")
	gstreamer.TriggerRecording(action)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

