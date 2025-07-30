package health

import (
	"context"
	"fmt"
	"testing"
	"time"
)

type mockChecker struct {
	shouldFail bool
	err        error
}

func (m *mockChecker) Ping(ctx context.Context) error {
	if m.shouldFail {
		return m.err
	}
	return nil
}

func TestHealth_Liveness(t *testing.T) {
	db := &mockChecker{}
	kafka := &mockChecker{}
	h := New(db, kafka)

	check := h.Liveness()

	if check.Status != StatusHealthy {
		t.Errorf("Liveness() status = %v, want %v", check.Status, StatusHealthy)
	}

	if check.Timestamp.IsZero() {
		t.Error("Liveness() timestamp should not be zero")
	}

	if check.Details != nil {
		t.Error("Liveness() details should be nil")
	}
}

func TestHealth_Readiness(t *testing.T) {
	tests := []struct {
		name         string
		dbHealthy    bool
		dbError      error
		kafkaHealthy bool
		kafkaError   error
		wantStatus   Status
		wantDetails  int
	}{
		{
			name:         "all healthy",
			dbHealthy:    true,
			kafkaHealthy: true,
			wantStatus:   StatusHealthy,
			wantDetails:  2,
		},
		{
			name:         "database unhealthy",
			dbHealthy:    false,
			dbError:      fmt.Errorf("connection refused"),
			kafkaHealthy: true,
			wantStatus:   StatusUnhealthy,
			wantDetails:  2,
		},
		{
			name:         "kafka unhealthy",
			dbHealthy:    true,
			kafkaHealthy: false,
			kafkaError:   fmt.Errorf("broker not available"),
			wantStatus:   StatusUnhealthy,
			wantDetails:  2,
		},
		{
			name:         "both unhealthy",
			dbHealthy:    false,
			dbError:      fmt.Errorf("connection refused"),
			kafkaHealthy: false,
			kafkaError:   fmt.Errorf("broker not available"),
			wantStatus:   StatusUnhealthy,
			wantDetails:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &mockChecker{
				shouldFail: !tt.dbHealthy,
				err:        tt.dbError,
			}
			kafka := &mockChecker{
				shouldFail: !tt.kafkaHealthy,
				err:        tt.kafkaError,
			}
			h := New(db, kafka)

			ctx := context.Background()
			check := h.Readiness(ctx)

			if check.Status != tt.wantStatus {
				t.Errorf("Readiness() status = %v, want %v", check.Status, tt.wantStatus)
			}

			if check.Timestamp.IsZero() {
				t.Error("Readiness() timestamp should not be zero")
			}

			if len(check.Details) != tt.wantDetails {
				t.Errorf("Readiness() details length = %v, want %v", len(check.Details), tt.wantDetails)
			}

			dbDetail, ok := check.Details["database"].(map[string]interface{})
			if !ok {
				t.Fatal("database detail should exist and be a map")
			}

			if tt.dbHealthy {
				if dbDetail["status"] != "healthy" {
					t.Errorf("database status = %v, want healthy", dbDetail["status"])
				}
				if _, exists := dbDetail["error"]; exists {
					t.Error("database error should not exist when healthy")
				}
			} else {
				if dbDetail["status"] != "unhealthy" {
					t.Errorf("database status = %v, want unhealthy", dbDetail["status"])
				}
				if dbDetail["error"] != tt.dbError.Error() {
					t.Errorf("database error = %v, want %v", dbDetail["error"], tt.dbError.Error())
				}
			}

			kafkaDetail, ok := check.Details["kafka"].(map[string]interface{})
			if !ok {
				t.Fatal("kafka detail should exist and be a map")
			}

			if tt.kafkaHealthy {
				if kafkaDetail["status"] != "healthy" {
					t.Errorf("kafka status = %v, want healthy", kafkaDetail["status"])
				}
				if _, exists := kafkaDetail["error"]; exists {
					t.Error("kafka error should not exist when healthy")
				}
			} else {
				if kafkaDetail["status"] != "unhealthy" {
					t.Errorf("kafka status = %v, want unhealthy", kafkaDetail["status"])
				}
				if kafkaDetail["error"] != tt.kafkaError.Error() {
					t.Errorf("kafka error = %v, want %v", kafkaDetail["error"], tt.kafkaError.Error())
				}
			}
		})
	}
}

func TestHealth_ReadinessTimeout(t *testing.T) {
	// Create a slow checker that simulates a timeout
	slowChecker := &slowMockChecker{}

	h := New(slowChecker, &mockChecker{})

	ctx := context.Background()
	start := time.Now()
	h.Readiness(ctx)
	duration := time.Since(start)

	// Should timeout within 5 seconds + some buffer
	if duration > 6*time.Second {
		t.Errorf("Readiness() took %v, should timeout at 5s", duration)
	}
}

type slowMockChecker struct{}

func (s *slowMockChecker) Ping(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(10 * time.Second):
		return nil
	}
}
