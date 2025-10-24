// cmd/nixon/main.go
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"nixon/internal/api"
	"nixon/internal/config"
	"nixon/internal/db"
	"nixon/internal/gstreamer"
	"nixon/internal/websocket" // Keep websocket import for callback
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.Println("Starting Nixon server...")

	// Initialize configuration first
	cfg := config.GetConfig()
	log.Println("Configuration initialized.")

	// Initialize Database
	if err := db.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Ensure recordings directory exists
	if err := os.MkdirAll(cfg.AutoRecord.Directory, 0755); err != nil {
		log.Printf("Warning: Failed to create recordings directory '%s': %v", cfg.AutoRecord.Directory, err)
	} else {
		log.Printf("Ensured recordings directory exists: %s", cfg.AutoRecord.Directory)
	}

	// Initialize GStreamer Manager, providing the broadcast callback (A3 Fix)
	// Pass websocket.BroadcastUpdate directly as the callback function
	if err := gstreamer.Init(websocket.BroadcastUpdate); err != nil {
		log.Printf("Warning: Failed to initialize GStreamer manager: %v. Server starting without pipeline.", err)
	} else {
		log.Println("GStreamer manager initialized.")
		// Attempt initial pipeline start asynchronously
		go func() {
			log.Println("Attempting initial GStreamer pipeline start...")
			if manager := gstreamer.GetManager(); manager != nil {
				if err := manager.StartPipeline(); err != nil {
					log.Printf("Initial GStreamer pipeline start failed: %v. Server will continue running.", err)
				}
			} else {
				log.Println("Error: GStreamer manager is nil after Init, cannot start pipeline.")
			}
		}()
	}

	// Start background tasks
	api.StartStatusUpdater() // Starts disk usage, listener monitoring
	go websocket.StartBroadcaster()
	// Polling is removed as gstreamer now pushes updates via callback (A3 Fix)
	// go websocket.PollAndBroadcast()
	log.Println("Background tasks started.")

	// Setup Router (A1: Using net/http)
	router := api.NewRouter()
	log.Println("API router initialized (using net/http).")

	// Setup HTTP Server with graceful shutdown
	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Goroutine for graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan // Wait for signal

		log.Println("Shutdown signal received, starting graceful shutdown...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Stop GStreamer pipeline gracefully
		if manager := gstreamer.GetManager(); manager != nil {
			log.Println("Stopping GStreamer pipeline...")
			// Task 4: Correct function call
			manager.StopPipeline()
		}

		// Shutdown HTTP server
		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}
		log.Println("Server gracefully stopped.")
	}()

	// Start the server
	log.Println("Server listening on :8080")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Failed to start server: %v", err)
	}

	log.Println("Nixon server exiting.")
}
