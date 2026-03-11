package utils

import (
	"log/slog"
	"os"
)

// Logger is the package-level structured logger.
var Logger *slog.Logger

func init() {
	level := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	}

	Logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
}

// LogError logs an error with the given message and optional key/value pairs.
func LogError(msg string, err error, args ...any) {
	args = append(args, "error", err)
	Logger.Error(msg, args...)
}

// LogInfo logs an informational message with optional key/value pairs.
func LogInfo(msg string, args ...any) {
	Logger.Info(msg, args...)
}

// LogDebug logs a debug message with optional key/value pairs.
func LogDebug(msg string, args ...any) {
	Logger.Debug(msg, args...)
}

// LogWarn logs a warning message with optional key/value pairs.
func LogWarn(msg string, args ...any) {
	Logger.Warn(msg, args...)
}
