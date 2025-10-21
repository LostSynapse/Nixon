package main

import (
	"log"
	"nixon/internal/api"
	"nixon/internal/config"
	"nixon/internal/db"
	"nixon/internal/gstreamer"
	"os"
)

func main() {
	log.Println("Starting Nixon backend...")

	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if err := db.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	if err := os.MkdirAll(config.RecordingsDir, 0755); err != nil {
		log.Fatalf("Failed to create recordings directory: %v", err)
	}

	go gstreamer.MonitorVAD()

	router := api.SetupRouter()
	log.Println("Backend started on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start web server: %v", err)
	}
}

