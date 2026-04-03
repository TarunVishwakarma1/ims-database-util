package logger

import (
	"log/slog"
	"os"
	"strings"
)

// NewLogger creates a structured slog logger.
// It reads LOG_LEVEL env var (debug, info, warn, error) and defaults to INFO.
// Output is JSON for easy ingestion by log aggregators.
func NewLogger() *slog.Logger {
	levelStr := strings.ToLower(os.Getenv("LOG_LEVEL"))
	var level slog.Level
	switch levelStr {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	opts := &slog.HandlerOptions{Level: level}
	// JSONHandler is available in Go 1.21+. For earlier versions, use TextHandler.
	// We'll use JSONHandler for production-friendly logs.
	handler := slog.NewJSONHandler(os.Stdout, opts)
	return slog.New(handler)
}
