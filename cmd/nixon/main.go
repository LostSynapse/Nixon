// cmd/nixon/main.go
package main

import (
	"log"
	"net/http"
	"nixon/internal/api"
	"nixon/internal/config" // Ensure config is imported
	"nixon/internal/db"
	"nixon/internal/gstreamer"
	"nixon/internal/websocket"
	"os"
	"time"
)

func main() {
	log.Println("Starting Nixon server...")

	// config.InitConfig() is no longer needed.
	// Initialization happens automatically on the first call to config.GetConfig() below.

	// Initialize Database first (relies on default config path)
	if err := db.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Now get the config, which implicitly initializes it if needed.
	conf := config.GetConfig()

	// Create recordings directory using the now-loaded config path
	if err := os.MkdirAll(conf.AutoRecord.Directory, os.ModePerm); err != nil {
		// Log as warning, not fatal, as the dir might already exist or fail later
		log.Printf("Warning: Failed to create recordings directory '%s': %v", conf.AutoRecord.Directory, err)
	}

	// Initialize GStreamer Manager (but don't start the pipeline yet)
	gstreamer.Init(conf)

	// Initialize API Router (needs config for recordings path)
	router := api.NewRouter()

	// Start background tasks
	log.Println("Starting background tasks...")
	go api.StartStatusUpdater()      // This was missing
	go websocket.StartBroadcaster() // This was missing
	go gstreamer.GetManager().StartVadMonitor()

	// Start the web server
	log.Println("Server listening on :8080")
	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Attempt initial GStreamer pipeline start *after* server starts
	go func() {
		log.Println("Attempting initial GStreamer pipeline start...")
		// Wait a tiny bit for the server goroutine to start listening
		time.Sleep(100 * time.Millisecond)
		if err := gstreamer.GetManager().StartPipeline(); err != nil {
			log.Printf("Initial GStreamer pipeline start failed: %v. Server will continue running.", err)
		}
	}()

	// Start the HTTP server (blocking call)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start web server: %v", err)
	}
}

