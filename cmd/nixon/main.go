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
	"context"
    "os/signal"
    "syscall"
    "time"

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

	// --- Graceful Shutdown Logic ---
	server := &http.Server{
		Addr:    listenAddress,
		Handler: router,
	}

	// Create a channel to listen for OS signals
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	// Run the server in a goroutine so that it doesn't block
	go func() {
		slogger.Log.Info("Server starting", "listen_address", listenAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slogger.Log.Error("Server failed to start", "err", err)
			os.Exit(1)
		}
	}()

	// Block until a signal is received
	<-stopChan
	slogger.Log.Info("Shutdown signal received, starting graceful shutdown...")

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait until the timeout deadline.
	if err := server.Shutdown(ctx); err != nil {
		slogger.Log.Error("Server shutdown failed", "err", err)
	}

	// TODO: Add any other cleanup tasks here (e.g., closing DB, stopping manager)

	slogger.Log.Info("Server exited gracefully")

}
