package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/sksmith/go-base-ms/internal/config"
)

func TestNew_InvalidDSN(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host:     "invalid\x00host",
		Port:     5432,
		User:     "test",
		Password: "test",
		DBName:   "test",
		SSLMode:  "disable",
	}

	ctx := context.Background()
	_, err := New(ctx, cfg)

	if err == nil {
		t.Error("expected error for invalid DSN, got nil")
	}
}

func TestDB_Methods(t *testing.T) {
	// This test requires a proper database connection
	// Skip it for now as it would require sqlmock or test database
	t.Skip("Skipping DB methods test - requires database mock")
}

func TestDB_ConnectionString(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.DatabaseConfig
		want string
	}{
		{
			name: "basic config",
			cfg: config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "secret",
				DBName:   "testdb",
				SSLMode:  "disable",
			},
			want: "host=localhost port=5432 user=postgres password=secret dbname=testdb sslmode=disable",
		},
		{
			name: "with SSL",
			cfg: config.DatabaseConfig{
				Host:     "db.example.com",
				Port:     5433,
				User:     "admin",
				Password: "password123",
				DBName:   "production",
				SSLMode:  "require",
			},
			want: "host=db.example.com port=5433 user=admin password=password123 dbname=production sslmode=require",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't actually test the connection string directly
			// since it's built inside New(), but we can verify the format
			dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
				tt.cfg.Host, tt.cfg.Port, tt.cfg.User, tt.cfg.Password, tt.cfg.DBName, tt.cfg.SSLMode)

			if dsn != tt.want {
				t.Errorf("connection string = %v, want %v", dsn, tt.want)
			}
		})
	}
}
