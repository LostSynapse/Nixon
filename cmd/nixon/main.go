package main

import (
	"fmt"
	"net/http"
    "nixon/internal/slogger"
	"nixon/internal/api"
	"nixon/internal/config"
	"nixon/internal/control"
	"nixon/internal/websocket"
    "os"
)

func main() {
	slogger.InitSlogger()
	// Load configuration
	// CHANGED: Call config.LoadConfig() directly
	// It now populates config.AppConfig globally and handles its own errors internally.
	config.LoadConfig()
    
	// REMOVED: Previous 'if err != nil { log.Fatalf(...) }' block
	// REMOVED: Previous 'config.SetConfig(cfg)' call
	// These are no longer needed as config.LoadConfig manages the global AppConfig.

	// Initialize the Control Manager
	ctrl, err := control.GetManager()
	if err != nil {
		slogger.Log.Error("Error initializing control manager", "err", err)
os.Exit(1)

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

	slogger.Log.Info("Server starting", "listen_address", listenAddress)
	if err := http.ListenAndServe(listenAddress, router); err != nil {
		slogger.Log.Error("Server failed", "err", err)
		os.Exit(1)

	}
}
