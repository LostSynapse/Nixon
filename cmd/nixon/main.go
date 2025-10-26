package main

import (
	"fmt"
	"nixon/internal/logger"
	"net/http"

	"nixon/internal/api"
	"nixon/internal/config"
	"nixon/internal/control"
	"nixon/internal/websocket"
)

func main() {
	logger.InitLogger()
	// Load configuration
	// CHANGED: Call config.LoadConfig() directly.
	// It now populates config.AppConfig globally and handles its own errors internally.
	config.LoadConfig()
    
	// REMOVED: Previous 'if err != nil { log.Fatalf(...) }' block
	// REMOVED: Previous 'config.SetConfig(cfg)' call
	// These are no longer needed as config.LoadConfig manages the global AppConfig.

	// Initialize the Control Manager
	ctrl, err := control.GetManager()
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Error initializing control manager")
	}

	// Start background tasks
	ctrl.StartBackgroundTasks()

	// Start the WebSocket message broadcaster
	go websocket.HandleMessages()

	// Setup router
	router := api.NewRouter(ctrl)

	// Get listen address from configuration
	// CHANGED: Access config directly from config.AppConfig.Web.ListenAddress
	listenAddress := fmt.Sprintf(":%s", config.AppConfig.Web.ListenAddress)

	logger.Log.Info().Str("listen_address", listenAddress).Msg("Server starting")
	if err := http.ListenAndServe(listenAddress, router); err != nil {
		logger.Log.Fatal().Err(err).Msg("Server failed")
	}
}
