// internal/api/router.go
// This file sets up the API routes for configuration, control, and data access.

package api

import (
	"encoding/json"
	"log"
	"net/http"
	"nixon/internal/config"
	"nixon/internal/db"
	"nixon/internal/gstreamer"
	"nixon/internal/websocket"
	"strconv"

	"github.com/gin-gonic/gin"
)

// InitRouter initializes the Gin router and defines all API endpoints.
func InitRouter() *gin.Engine {
	// Gin is used here for its simplicity and robustness, though for a lean appliance
	// a custom net/http router could be considered later.
	router := gin.Default()

	// Serve the static React assets from the 'web/dist' directory
	// In the future, this will be replaced by an embedded Go server solution.
	router.StaticFile("/", "./web/dist/index.html")
	router.Static("/assets", "./web/dist/assets")

	// API routes group
	api := router.Group("/api")
	{
		// --- Status and Configuration ---

		// GET /api/status - Returns the current combined status of the application
		api.GET("/status", getStatusHandler)

		// GET /api/config - Returns the current application configuration
		api.GET("/config", getConfigHandler)

		// POST /api/config - Updates and saves the entire application configuration
		api.POST("/config", updateConfigHandler)

		// GET /api/devices - Returns a list of available audio input devices
		api.GET("/devices", getAudioDevicesHandler)

		// --- Control ---

		// POST /api/start - Starts the GStreamer audio pipeline
		api.POST("/start", startPipelineHandler)

		// POST /api/stop - Stops the GStreamer audio pipeline
		api.POST("/stop", stopPipelineHandler)

		// --- Database/Recordings ---

		// GET /api/recordings - Lists all recordings
		api.GET("/recordings", listRecordingsHandler)

		// DELETE /api/recordings/:id - Deletes a specific recording entry
		api.DELETE("/recordings/:id", deleteRecordingHandler)

		// PUT /api/recordings/:id - Updates metadata (notes/genre) for a recording
		api.PUT("/recordings/:id", updateRecordingHandler)
	}

	// WebSocket route
	router.GET("/ws", websocket.HandleConnections)

	return router
}

// --- Handler Implementations ---

func getStatusHandler(c *gin.Context) {
	// Combine GStreamer status with monitoring data for a comprehensive view
	gstStatus := gstreamer.GetManager().GetStatus()
	
	// Create a map to combine GStreamer status and monitoring data
	statusMap := make(map[string]interface{})
	
	// Marshal and unmarshal GStreamer status to convert it to a map[string]interface{}
	gstBytes, _ := json.Marshal(gstStatus)
	json.Unmarshal(gstBytes, &statusMap)

	// Add listener and disk info from the manager's state
	current, peak := gstreamer.GetManager().GetListeners()
	statusMap["listener_current"] = current
	statusMap["listener_peak"] = peak
	statusMap["disk_usage"] = gstreamer.GetManager().GetDiskUsage()

	c.JSON(http.StatusOK, statusMap)
}

func getConfigHandler(c *gin.Context) {
	c.JSON(http.StatusOK, config.GetConfig())
}

// Request body structure for configuration updates
type updateConfigRequest struct {
	config.Config
}

func updateConfigHandler(c *gin.Context) {
	var reqBody updateConfigRequest

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body format"})
		return
	}

	// FIX: Use the new SaveGlobalConfig function
	if err := config.SaveGlobalConfig(reqBody.Config); err != nil {
		log.Printf("ERROR: Failed to save config: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save configuration"})
		return
	}

	// FIX: Convert the struct to a map[string]interface{} before broadcasting
	// This is necessary because the WebSocket broadcast function expects a map for JSON serialization.
	cfgMap := make(map[string]interface{})
	cfgBytes, _ := json.Marshal(reqBody.Config)
	json.Unmarshal(cfgBytes, &cfgMap)

	websocket.BroadcastUpdate(map[string]interface{}{
		"config": cfgMap,
	})

	// Re-start pipeline to apply new settings if it was running (e.g., sample rate change)
	gstreamer.GetManager().RestartPipeline()

	c.JSON(http.StatusOK, gin.H{"message": "Configuration updated successfully"})
}

func startPipelineHandler(c *gin.Context) {
	if err := gstreamer.GetManager().StartPipeline(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Pipeline started"})
}

func stopPipelineHandler(c *gin.Context) {
	if err := gstreamer.GetManager().StopPipeline(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Pipeline stopped"})
}

func listRecordingsHandler(c *gin.Context) {
	recordings, err := db.ListRecordings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve recordings"})
		return
	}
	c.JSON(http.StatusOK, recordings)
}

func deleteRecordingHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recording ID format"})
		return
	}

	// TODO: Add logic to delete the actual file from disk here (Phase 2/4)

	if err := db.DeleteRecording(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete recording from database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Recording deleted"})
}

// Request body structure for updating recording metadata
type updateRecordingRequest struct {
	// Pointers are used for optional fields that may not be present in the request
	Notes *string `json:"notes"`
	Genre *string `json:"genre"`
}

func updateRecordingHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recording ID format"})
		return
	}

	var reqBody updateRecordingRequest
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body format"})
		return
	}

	// FIX: Safely dereference pointers to concrete strings for db.UpdateRecording
	notes := ""
	if reqBody.Notes != nil {
		notes = *reqBody.Notes
	}
	genre := ""
	if reqBody.Genre != nil {
		genre = *reqBody.Genre
	}

	// db.UpdateRecording now accepts concrete strings (as fixed in internal/db/db.go)
	if err := db.UpdateRecording(uint(id), notes, genre); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update recording metadata"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Recording metadata updated"})
}

func getAudioDevicesHandler(c *gin.Context) {
	// Call the new PipeWire-based device lister
	devices, err := gstreamer.ListAudioDevices()
	if err != nil {
		log.Printf("ERROR: Failed to list audio devices: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve audio devices"})
		return
	}

	c.JSON(http.StatusOK, devices)
}
