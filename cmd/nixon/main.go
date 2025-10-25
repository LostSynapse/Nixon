package main

import (
	"context"
	"log"
	"net/http"
	"nixon/internal/api"
	"nixon/internal/config"
	"nixon/internal/control"
	"nixon/internal/db"
	"nixon/internal/pipewire"
	"nixon/internal/websocket"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux" // This is required by api.NewRouter
)

func main() {
	log.Println("Starting Nixon v2...")

	// --- Configuration ---
	// LoadConfig panics on fatal error
	config.LoadConfig()
	cfg := config.GetConfig()

	// --- Database ---
	if err := db.Init(cfg.Database.DSN); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// --- WebSocket Hub ---
	// GetHub uses sync.Once for safe initialization
	hub := websocket.GetHub()
	go hub.Run()
	log.Println("WebSocket hub initialized.")

	// --- Core Services ---
	// GetManager now requires config
	audioManager := pipewire.GetManager(cfg)
	log.Println("PipeWire Audio Manager initialized.")

	ctrl := control.GetManager(cfg, audioManager)
	log.Println("Control Manager initialized.")

	// --- API Tasks ---
	// FIXED: InitTasks now requires config
	api.InitTasks(cfg)
	log.Println("API tasks initialized.")

	// --- HTTP Server & API Router ---
	// FIXED: NewRouter now accepts the control manager and the hub
	r := api.NewRouter(ctrl, hub)

	// Serve the static React build
	// This assumes the 'web/dist' directory exists relative to the binary
	// or in the working directory.
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/dist")))

	srv := &http.Server{
		Addr:    cfg.Server.ListenAddr,
		Handler: r,
	}

	// --- Graceful Shutdown ---
	go func() {
		log.Printf("Server starting on %s", cfg.Server.ListenAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Shutdown PipeWire Manager (TODO: Implement audioManager.Shutdown())
	// if err := audioManager.Shutdown(); err != nil {
	// 	log.Printf("Error during audio manager shutdown: %v", err)
	// }
	// log.Println("Audio manager shut down.")

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server gracefully stopped.")
}

