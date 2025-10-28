package slogger

import (
	"log/slog"
	"os"
)

// Log is the global logger instance
var Log *slog.Logger

// InitSlogger initializes the global slog logger with a JSON handler.
func InitSlogger() {
	Log = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	Log.Info("Slogger initialized")
}
