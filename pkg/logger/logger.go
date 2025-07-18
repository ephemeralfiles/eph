// Package logger provides structured logging utilities for the ephemeral files CLI.
package logger

import (
	"log/slog"
	"os"
)

// NewLogger creates a new logger
// logLevel is the level of logging
// Possible values of logLevel are: "debug", "info", "warn", "error"
// Default value is "info".
func NewLogger(logLevel string) *slog.Logger {
	var level slog.Level
	switch logLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	logHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: false,
	})
	logger := slog.New(logHandler)
	return logger
}

// NoLogger creates a logger that does not log anything.
func NoLogger() *slog.Logger {
	noLogger := slog.New(slog.DiscardHandler)
	return noLogger
}
