package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Log is the global logger instance
var Log zerolog.Logger

// InitLogger initializes the global logger with a human-friendly console writer.
func InitLogger() {
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	Log = zerolog.New(consoleWriter).With().Timestamp().Logger()
	Log.Info().Msg("Logger initialized")
}
