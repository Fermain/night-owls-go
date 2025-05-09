package logging

import (
	"log/slog"
	"os"
)

// NewLogger creates and returns a new slog.Logger instance.
// It defaults to a JSON handler writing to os.Stdout.
func NewLogger() *slog.Logger {
	// TODO: Make log level configurable (e.g., from Config struct)
	// TODO: Allow different handlers (e.g., text vs json) based on config
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // Default to Debug for now, can be changed
	}))
	return logger
} 