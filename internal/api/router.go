package api

import (
	"github.com/gin-gonic/gin"
	"nixon/internal/config"
	"nixon/internal/websocket"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	router.Static("/assets", "./web/assets")
	router.Static("/recordings", config.RecordingsDir)
	router.StaticFile("/nixon_logo.svg", "./web/nixon_logo.svg")

	api := router.Group("/api")
	{
		api.GET("/status", getStatus)
		api.POST("/stream/srt/:action", handleSRTStream)
		api.POST("/stream/icecast/:action", handleIcecastStream)
		api.POST("/stream/all/:action", handleAllStreams)
		api.POST("/recording/:action", handleRecording)

		api.GET("/recordings", handleGetRecordings)
		api.PUT("/recordings/:id", handleUpdateRecording)
		api.POST("/recordings/:id/protect", handleToggleProtect)
		api.DELETE("/recordings/:id", handleDeleteRecording)

		api.GET("/settings/all", getFullConfig)
		api.POST("/settings/icecast", updateIcecastSettings)
		api.POST("/settings/system", updateSystemSettings)
		api.POST("/settings/audio", updateAudioSettings)

		api.GET("/system/audiodevices", handleGetAudioDevices)
	}

	router.GET("/ws", websocket.HandleWebSocket)
	router.NoRoute(func(c *gin.Context) {
		c.File("./web/index.html")
	})

	go monitorDiskUsage()
	go monitorIcecastListeners()
	go websocket.PollAndBroadcast()
	go websocket.HandleBroadcast()

	return router
}

