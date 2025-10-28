package main

import (
	"context"
	//"fmt"
	"net/http"
	"nixon/internal/api"
	"nixon/internal/config"
	"nixon/internal/control"
	"nixon/internal/slogger"
	"nixon/internal/websocket"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	slogger.InitSlogger()
	config.LoadConfig()

	ctrl, err := control.GetManager()
	if err != nil {
		slogger.Log.Error("Error initializing control manager", "err", err)
		os.Exit(1)
	}

	ctrl.StartAudio()
	go websocket.HandleMessages()

	router := api.NewRouter(ctrl)

	server := &http.Server{
		Addr: ":" + config.AppConfig.Web.ListenAddress,
		Handler: router,
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		slogger.Log.Info("Server starting", "listen_address", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slogger.Log.Error("Server failed to start", "err", err)
			os.Exit(1)
		}
	}()

	<-stopChan
	slogger.Log.Info("Shutdown signal received, starting graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slogger.Log.Error("Server shutdown failed", "err", err)
	}

	slogger.Log.Info("Server exited gracefully")
}
