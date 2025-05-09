package logging

import (
	"log/slog"
	"os"
	"strings"

	"night-owls-go/internal/config"
)

// NewLogger creates and returns a new slog.Logger instance based on config.
func NewLogger(cfg *config.Config) *slog.Logger {
	var level slog.Level
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo // Default to Info if invalid or empty
	}

	opts := &slog.HandlerOptions{
		Level: level,
		// AddSource: true, // Optionally add source file and line to logs
	}

	var handler slog.Handler
	switch strings.ToLower(cfg.LogFormat) {
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		handler = slog.NewJSONHandler(os.Stdout, opts) // Default to JSON
	}

	logger := slog.New(handler)
	return logger
} 