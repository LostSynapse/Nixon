package main

import (
	"log"
	"net/http"
	"nixon/internal/api"
	"nixon/internal/config"
	"nixon/internal/control"
	"nixon/internal/websocket"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	config.SetConfig(cfg) // Set the global config

	// Initialize the Control Manager
	ctrl, err := control.GetManager()
	if err != nil {
		log.Fatalf("Error initializing control manager: %v", err)
	}

	// Start background tasks
	ctrl.StartBackgroundTasks()

	// Start the WebSocket message broadcaster
	go websocket.HandleMessages()

	// Setup router
	router := api.NewRouter(ctrl)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
