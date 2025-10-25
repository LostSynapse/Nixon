package main

import (
	"log"
	"net/http"

	"nixon/internal/api"
	"nixon/internal/config"
	"nixon/internal/control"
)

func main() {
	// Load configuration
	// FIXED: config.Load() is the correct way to initialize.
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	config.SetConfig(cfg) // Set the global config

	// Initialize the Control Manager
	// FIXED: GetManager initializes the singleton instance.
	ctrl, err := control.GetManager()
	if err != nil {
		log.Fatalf("Error initializing control manager: %v", err)
	}

	// Start background tasks
	ctrl.StartBackgroundTasks()

	// Setup router
	router := api.NewRouter(ctrl)

	// Serve static files and API
	// Note: The Chi router now handles serving static files.
	// We pass the router directly to ListenAndServe.
	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
