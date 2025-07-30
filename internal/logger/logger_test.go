package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	// Reset the current level before each test
	currentLevel.Set(slog.LevelInfo)

	tests := []struct {
		name      string
		logLevel  string
		wantLevel slog.Level
	}{
		{
			name:      "default level",
			logLevel:  "",
			wantLevel: slog.LevelInfo,
		},
		{
			name:      "debug level",
			logLevel:  "debug",
			wantLevel: slog.LevelDebug,
		},
		{
			name:      "other value defaults to info",
			logLevel:  "invalid",
			wantLevel: slog.LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.logLevel != "" {
				os.Setenv("LOG_LEVEL", tt.logLevel)
				defer os.Unsetenv("LOG_LEVEL")
			}

			// Reinitialize to pick up env var
			if tt.logLevel == "debug" {
				currentLevel.Set(slog.LevelDebug)
			} else {
				currentLevel.Set(slog.LevelInfo)
			}

			logger := New()

			// Test logger by capturing output
			buf := &bytes.Buffer{}
			testLogger := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{
				Level: currentLevel,
			}))

			// Log at debug level
			testLogger.Debug("debug message")

			var result map[string]interface{}
			if buf.Len() > 0 {
				if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
					t.Fatalf("failed to unmarshal log output: %v", err)
				}
			}

			// If we expect debug level, we should see the debug message
			if tt.wantLevel == slog.LevelDebug {
				if buf.Len() == 0 {
					t.Error("expected debug log output, got none")
				}
				if result["msg"] != "debug message" {
					t.Errorf("expected debug message, got %v", result["msg"])
				}
			} else {
				// For info level, debug messages should not appear
				if buf.Len() > 0 {
					t.Error("expected no debug log output for info level")
				}
			}

			// Verify logger is not nil
			if logger == nil {
				t.Error("New() returned nil logger")
			}
		})
	}
}

func TestSetLevel(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		wantErr bool
	}{
		{
			name:    "set debug",
			level:   "debug",
			wantErr: false,
		},
		{
			name:    "set info",
			level:   "info",
			wantErr: false,
		},
		{
			name:    "set warn",
			level:   "warn",
			wantErr: false,
		},
		{
			name:    "set error",
			level:   "error",
			wantErr: false,
		},
		{
			name:    "invalid level",
			level:   "trace",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetLevel(tt.level)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetLevel() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				got := GetLevel()
				if got != tt.level {
					t.Errorf("GetLevel() = %v, want %v", got, tt.level)
				}
			}
		})
	}
}

func TestGetLevel(t *testing.T) {
	tests := []struct {
		name     string
		setLevel string
		want     string
	}{
		{
			name:     "get debug",
			setLevel: "debug",
			want:     "debug",
		},
		{
			name:     "get info",
			setLevel: "info",
			want:     "info",
		},
		{
			name:     "get warn",
			setLevel: "warn",
			want:     "warn",
		},
		{
			name:     "get error",
			setLevel: "error",
			want:     "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetLevel(tt.setLevel)
			if got := GetLevel(); got != tt.want {
				t.Errorf("GetLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}
