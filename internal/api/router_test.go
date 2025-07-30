package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/dks0523168/go-base-ms/internal/health"
	internalLogger "github.com/dks0523168/go-base-ms/internal/logger"
)

type mockChecker struct {
	shouldFail bool
}

func (m *mockChecker) Ping(ctx context.Context) error {
	if m.shouldFail {
		return fmt.Errorf("mock error")
	}
	return nil
}

func TestRouter_LivenessHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
	h := health.New(&mockChecker{}, &mockChecker{})
	router := NewRouter(logger, h)

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response health.Check
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != health.StatusHealthy {
		t.Errorf("expected status %s, got %s", health.StatusHealthy, response.Status)
	}
}

func TestRouter_ReadinessHandler(t *testing.T) {
	tests := []struct {
		name           string
		dbHealthy      bool
		kafkaHealthy   bool
		expectedStatus int
		expectedHealth health.Status
	}{
		{
			name:           "all healthy",
			dbHealthy:      true,
			kafkaHealthy:   true,
			expectedStatus: http.StatusOK,
			expectedHealth: health.StatusHealthy,
		},
		{
			name:           "db unhealthy",
			dbHealthy:      false,
			kafkaHealthy:   true,
			expectedStatus: http.StatusServiceUnavailable,
			expectedHealth: health.StatusUnhealthy,
		},
		{
			name:           "kafka unhealthy",
			dbHealthy:      true,
			kafkaHealthy:   false,
			expectedStatus: http.StatusServiceUnavailable,
			expectedHealth: health.StatusUnhealthy,
		},
		{
			name:           "both unhealthy",
			dbHealthy:      false,
			kafkaHealthy:   false,
			expectedStatus: http.StatusServiceUnavailable,
			expectedHealth: health.StatusUnhealthy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
			db := &mockChecker{shouldFail: !tt.dbHealthy}
			kafka := &mockChecker{shouldFail: !tt.kafkaHealthy}
			h := health.New(db, kafka)
			router := NewRouter(logger, h)

			req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response health.Check
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if response.Status != tt.expectedHealth {
				t.Errorf("expected health status %s, got %s", tt.expectedHealth, response.Status)
			}
		})
	}
}

func TestRouter_HelloHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
	h := health.New(&mockChecker{}, &mockChecker{})
	router := NewRouter(logger, h)

	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "GET request",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST request",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v1/hello", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]string
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if response["message"] != "Hello from Go Base Microservice" {
					t.Errorf("unexpected message: %s", response["message"])
				}

				if response["version"] != "1.0.0" {
					t.Errorf("unexpected version: %s", response["version"])
				}
			}
		})
	}
}

func TestRouter_EchoHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
	h := health.New(&mockChecker{}, &mockChecker{})
	router := NewRouter(logger, h)

	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid POST request",
			method:         http.MethodPost,
			body:           `{"test": "data", "number": 123}`,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"test": "data", "number": 123}`,
		},
		{
			name:           "invalid JSON",
			method:         http.MethodPost,
			body:           `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error": "Invalid JSON body"}`,
		},
		{
			name:           "GET request",
			method:         http.MethodGet,
			body:           "",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body *strings.Reader
			if tt.body != "" {
				body = strings.NewReader(tt.body)
			} else {
				body = strings.NewReader("")
			}

			req := httptest.NewRequest(tt.method, "/api/v1/echo", body)
			if tt.method == http.MethodPost {
				req.Header.Set("Content-Type", "application/json")
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			responseBody := strings.TrimSpace(w.Body.String())

			if tt.expectedStatus == http.StatusOK || tt.expectedStatus == http.StatusBadRequest {
				var expected, actual map[string]interface{}
				if err := json.Unmarshal([]byte(tt.expectedBody), &expected); err != nil {
					t.Fatalf("failed to unmarshal expected body: %v", err)
				}
				if err := json.Unmarshal([]byte(responseBody), &actual); err != nil {
					t.Fatalf("failed to unmarshal actual body: %v", err)
				}

				for k, v := range expected {
					if actual[k] != v {
						t.Errorf("expected %s=%v, got %v", k, v, actual[k])
					}
				}
			} else if !strings.Contains(responseBody, tt.expectedBody) {
				t.Errorf("expected body to contain %q, got %q", tt.expectedBody, responseBody)
			}
		})
	}
}

func TestRouter_OpenapiHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
	h := health.New(&mockChecker{}, &mockChecker{})
	router := NewRouter(logger, h)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		contentType    string
	}{
		{
			name:           "openapi.json",
			path:           "/openapi.json",
			expectedStatus: http.StatusNotFound, // File doesn't exist in test environment
			contentType:    "text/plain; charset=utf-8",
		},
		{
			name:           "openapi.yaml",
			path:           "/openapi.yaml",
			expectedStatus: http.StatusNotFound, // File doesn't exist in test environment
			contentType:    "text/plain; charset=utf-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != tt.contentType {
				t.Errorf("expected Content-Type %q, got %q", tt.contentType, contentType)
			}
		})
	}
}

func TestRouter_OpenapiHandler_WithFile(t *testing.T) {
	// This test runs only if the OpenAPI files exist
	logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
	h := health.New(&mockChecker{}, &mockChecker{})
	router := NewRouter(logger, h)

	// First generate the OpenAPI files
	if err := generateTestOpenAPIFiles(t); err != nil {
		t.Skip("Skipping OpenAPI file test: ", err)
	}
	defer cleanupTestOpenAPIFiles(t)

	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type %q, got %q", "application/json", contentType)
	}

	var spec map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&spec); err != nil {
		t.Fatalf("failed to decode OpenAPI spec: %v", err)
	}

	if spec["openapi"] != "3.0.3" {
		t.Errorf("expected OpenAPI version 3.0.3, got %v", spec["openapi"])
	}

	info, ok := spec["info"].(map[string]interface{})
	if !ok {
		t.Fatal("expected info field in OpenAPI spec")
	}

	if info["title"] != "Go Base Microservice" {
		t.Errorf("unexpected title: %v", info["title"])
	}
}

func TestRouter_LogLevelHandler(t *testing.T) {
	// Save original log level to restore after tests
	originalLevel := internalLogger.GetLevel()
	defer internalLogger.SetLevel(originalLevel)

	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
		expectedLevel  string
		expectError    bool
	}{
		{
			name:           "GET log level",
			method:         http.MethodGet,
			body:           "",
			expectedStatus: http.StatusOK,
			expectedLevel:  "", // Will check current level
			expectError:    false,
		},
		{
			name:           "PUT debug level",
			method:         http.MethodPut,
			body:           `{"level": "debug"}`,
			expectedStatus: http.StatusOK,
			expectedLevel:  "debug",
			expectError:    false,
		},
		{
			name:           "PUT info level",
			method:         http.MethodPut,
			body:           `{"level": "info"}`,
			expectedStatus: http.StatusOK,
			expectedLevel:  "info",
			expectError:    false,
		},
		{
			name:           "PUT warn level",
			method:         http.MethodPut,
			body:           `{"level": "warn"}`,
			expectedStatus: http.StatusOK,
			expectedLevel:  "warn",
			expectError:    false,
		},
		{
			name:           "PUT error level",
			method:         http.MethodPut,
			body:           `{"level": "error"}`,
			expectedStatus: http.StatusOK,
			expectedLevel:  "error",
			expectError:    false,
		},
		{
			name:           "PUT invalid level",
			method:         http.MethodPut,
			body:           `{"level": "trace"}`,
			expectedStatus: http.StatusBadRequest,
			expectedLevel:  "",
			expectError:    true,
		},
		{
			name:           "PUT invalid JSON",
			method:         http.MethodPut,
			body:           `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
			expectedLevel:  "",
			expectError:    true,
		},
		{
			name:           "DELETE method not allowed",
			method:         http.MethodDelete,
			body:           "",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedLevel:  "",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
			h := health.New(&mockChecker{}, &mockChecker{})
			router := NewRouter(logger, h)

			var body *strings.Reader
			if tt.body != "" {
				body = strings.NewReader(tt.body)
			} else {
				body = strings.NewReader("")
			}

			req := httptest.NewRequest(tt.method, "/api/v1/admin/log-level", body)
			if tt.method == http.MethodPut {
				req.Header.Set("Content-Type", "application/json")
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]string
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if tt.method == http.MethodGet {
					if _, ok := response["level"]; !ok {
						t.Error("expected level field in GET response")
					}
				} else if tt.method == http.MethodPut && !tt.expectError {
					if response["level"] != tt.expectedLevel {
						t.Errorf("expected level %s, got %s", tt.expectedLevel, response["level"])
					}
					if response["message"] != "Log level updated successfully" {
						t.Errorf("unexpected message: %s", response["message"])
					}
					// Verify the level was actually changed
					if internalLogger.GetLevel() != tt.expectedLevel {
						t.Errorf("log level not actually changed, expected %s, got %s", tt.expectedLevel, internalLogger.GetLevel())
					}
				}
			}

			if tt.expectError && tt.expectedStatus == http.StatusBadRequest {
				var response map[string]string
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode error response: %v", err)
				}
				if _, ok := response["error"]; !ok {
					t.Error("expected error field in error response")
				}
			}
		})
	}
}

// Helper functions for OpenAPI testing
func generateTestOpenAPIFiles(t *testing.T) error {
	// Create a minimal test OpenAPI spec
	spec := `{
  "openapi": "3.0.3",
  "info": {
    "title": "Go Base Microservice",
    "version": "1.0.0"
  },
  "paths": {}
}`

	// Create api directory if it doesn't exist
	if err := os.MkdirAll("api", 0755); err != nil {
		return err
	}

	// Write JSON file
	if err := os.WriteFile("api/openapi.json", []byte(spec), 0644); err != nil {
		return err
	}

	return nil
}

func cleanupTestOpenAPIFiles(t *testing.T) {
	os.Remove("api/openapi.json")
	os.Remove("api/openapi.yaml")
	os.Remove("api")
}
