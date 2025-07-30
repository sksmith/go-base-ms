package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sksmith/go-base-ms/internal/health"
	"github.com/sksmith/go-base-ms/internal/logger"
	"github.com/sksmith/go-base-ms/internal/version"
)

type Router struct {
	mux    *http.ServeMux
	logger *slog.Logger
	health *health.Health
}

func NewRouter(logger *slog.Logger, health *health.Health) *Router {
	r := &Router{
		mux:    http.NewServeMux(),
		logger: logger,
		health: health,
	}

	r.setupRoutes()
	return r
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.logger.Info("request",
		"method", req.Method,
		"path", req.URL.Path,
		"remote_addr", req.RemoteAddr,
	)
	r.mux.ServeHTTP(w, req)
}

func (r *Router) setupRoutes() {
	r.mux.HandleFunc("/health/live", r.livenessHandler)
	r.mux.HandleFunc("/health/ready", r.readinessHandler)
	r.mux.HandleFunc("/version", r.versionHandler)
	r.mux.HandleFunc("/openapi.yaml", r.openapiHandler)
	r.mux.HandleFunc("/openapi.json", r.openapiHandler) // Keep backward compatibility
	r.mux.HandleFunc("/api/v1/hello", r.helloHandler)
	r.mux.HandleFunc("/api/v1/echo", r.echoHandler)
	r.mux.HandleFunc("/api/v1/admin/log-level", r.logLevelHandler)
}

func (r *Router) livenessHandler(w http.ResponseWriter, req *http.Request) {
	check := r.health.Liveness()
	r.respondJSON(w, http.StatusOK, check)
}

func (r *Router) readinessHandler(w http.ResponseWriter, req *http.Request) {
	check := r.health.Readiness(req.Context())

	status := http.StatusOK
	if check.Status == health.StatusUnhealthy {
		status = http.StatusServiceUnavailable
	}

	r.respondJSON(w, status, check)
}

func (r *Router) helloHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]string{
		"message": "Hello from Go Base Microservice",
		"version": "1.0.0",
	}
	r.respondJSON(w, http.StatusOK, response)
}

func (r *Router) echoHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body map[string]interface{}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		r.respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON body",
		})
		return
	}

	r.respondJSON(w, http.StatusOK, body)
}

func (r *Router) openapiHandler(w http.ResponseWriter, req *http.Request) {
	// Determine the file path based on the requested URL
	var filename string
	var contentType string

	if req.URL.Path == "/openapi.yaml" {
		filename = "api/openapi.yaml"
		contentType = "application/x-yaml"
	} else {
		// For backward compatibility, serve JSON version
		filename = "api/openapi.json"
		contentType = "application/json"
	}

	// Try to find the file relative to the current working directory
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// If not found, try relative to the executable
		if execPath, err := os.Executable(); err == nil {
			execDir := filepath.Dir(execPath)
			filename = filepath.Join(execDir, filename)
		}
	}

	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		r.logger.Error("OpenAPI spec file not found", "path", filename)
		http.Error(w, "OpenAPI specification not found", http.StatusNotFound)
		return
	}

	// Serve the file
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour
	http.ServeFile(w, req, filename)
}

func (r *Router) versionHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	versionInfo := version.Get()
	r.respondJSON(w, http.StatusOK, versionInfo)
}

func (r *Router) logLevelHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		response := map[string]string{
			"level": logger.GetLevel(),
		}
		r.respondJSON(w, http.StatusOK, response)

	case http.MethodPut:
		var request struct {
			Level string `json:"level"`
		}

		if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
			r.respondJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid JSON body",
			})
			return
		}

		if err := logger.SetLevel(request.Level); err != nil {
			r.respondJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		}

		r.logger.Info("log level changed", "new_level", request.Level)

		response := map[string]string{
			"level":   request.Level,
			"message": "Log level updated successfully",
		}
		r.respondJSON(w, http.StatusOK, response)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (r *Router) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		r.logger.Error("failed to encode response", "error", err)
	}
}
