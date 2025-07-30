package health

import (
	"context"
	"sync"
	"time"
)

type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
)

type Check struct {
	Status    Status                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

type Checker interface {
	Ping(ctx context.Context) error
}

type Health struct {
	checks map[string]Checker
	mu     sync.RWMutex
}

func New(db Checker, kafka Checker) *Health {
	return &Health{
		checks: map[string]Checker{
			"database": db,
			"kafka":    kafka,
		},
	}
}

func (h *Health) Liveness() Check {
	return Check{
		Status:    StatusHealthy,
		Timestamp: time.Now(),
	}
}

func (h *Health) Readiness(ctx context.Context) Check {
	h.mu.RLock()
	defer h.mu.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	allHealthy := true
	details := make(map[string]interface{})

	for name, checker := range h.checks {
		if err := checker.Ping(ctx); err != nil {
			allHealthy = false
			details[name] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
		} else {
			details[name] = map[string]interface{}{
				"status": "healthy",
			}
		}
	}

	status := StatusHealthy
	if !allHealthy {
		status = StatusUnhealthy
	}

	return Check{
		Status:    status,
		Timestamp: time.Now(),
		Details:   details,
	}
}
