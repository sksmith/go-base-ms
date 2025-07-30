package logger

import (
	"fmt"
	"log/slog"
	"os"
	"sync"
)

var (
	currentLevel = new(slog.LevelVar)
	mu           sync.RWMutex
)

func init() {
	// Set initial level from environment
	if os.Getenv("LOG_LEVEL") == "debug" {
		currentLevel.Set(slog.LevelDebug)
	} else {
		currentLevel.Set(slog.LevelInfo)
	}
}

func New() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: currentLevel,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	return slog.New(handler)
}

func SetLevel(level string) error {
	mu.Lock()
	defer mu.Unlock()

	switch level {
	case "debug":
		currentLevel.Set(slog.LevelDebug)
	case "info":
		currentLevel.Set(slog.LevelInfo)
	case "warn":
		currentLevel.Set(slog.LevelWarn)
	case "error":
		currentLevel.Set(slog.LevelError)
	default:
		return fmt.Errorf("invalid log level: %s", level)
	}
	return nil
}

func GetLevel() string {
	mu.RLock()
	defer mu.RUnlock()

	switch currentLevel.Level() {
	case slog.LevelDebug:
		return "debug"
	case slog.LevelInfo:
		return "info"
	case slog.LevelWarn:
		return "warn"
	case slog.LevelError:
		return "error"
	default:
		return "unknown"
	}
}
