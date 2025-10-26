package main

import (
	"fmt"
	"log" // Standard log package is still in use here
	"net/http"

	"nixon/internal/api"
	"nixon/internal/config"
	"nixon/internal/control"
	"nixon/internal/websocket"
)

func main() {
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
		log.Fatalf("Error initializing control manager: %v", err) // Standard log.Fatalf still used
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

	log.Printf("Server starting on %s", listenAddress) // Standard log.Printf still used
	if err := http.ListenAndServe(listenAddress, router); err != nil {
		log.Fatalf("Server failed: %v", err) // Standard log.Fatalf still used
	}
}
