package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		want    *Config
		wantErr bool
	}{
		{
			name:    "default values",
			envVars: map[string]string{},
			want: &Config{
				Port: 8080,
				Database: DatabaseConfig{
					Host:            "localhost",
					Port:            5432,
					User:            "postgres",
					Password:        "",
					DBName:          "gobase",
					SSLMode:         "disable",
					MaxOpenConns:    25,
					MaxIdleConns:    5,
					ConnMaxLifetime: 5,
				},
				Kafka: KafkaConfig{
					Brokers: []string{"localhost:9092"},
					Topic:   "events",
					GroupID: "go-base-ms",
				},
			},
			wantErr: false,
		},
		{
			name: "custom values",
			envVars: map[string]string{
				"PORT":                 "9090",
				"DB_HOST":              "db.example.com",
				"DB_PORT":              "5433",
				"DB_USER":              "testuser",
				"DB_PASSWORD":          "testpass",
				"DB_NAME":              "testdb",
				"DB_SSLMODE":           "require",
				"DB_MAX_OPEN_CONNS":    "50",
				"DB_MAX_IDLE_CONNS":    "10",
				"DB_CONN_MAX_LIFETIME": "10",
				"KAFKA_BROKERS":        "kafka1:9092",
				"KAFKA_TOPIC":          "test-events",
				"KAFKA_GROUP_ID":       "test-group",
			},
			want: &Config{
				Port: 9090,
				Database: DatabaseConfig{
					Host:            "db.example.com",
					Port:            5433,
					User:            "testuser",
					Password:        "testpass",
					DBName:          "testdb",
					SSLMode:         "require",
					MaxOpenConns:    50,
					MaxIdleConns:    10,
					ConnMaxLifetime: 10,
				},
				Kafka: KafkaConfig{
					Brokers: []string{"kafka1:9092"},
					Topic:   "test-events",
					GroupID: "test-group",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid port",
			envVars: map[string]string{
				"PORT": "invalid",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid db port",
			envVars: map[string]string{
				"DB_PORT": "invalid",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid max open conns",
			envVars: map[string]string{
				"DB_MAX_OPEN_CONNS": "invalid",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid max idle conns",
			envVars: map[string]string{
				"DB_MAX_IDLE_CONNS": "invalid",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid conn max lifetime",
			envVars: map[string]string{
				"DB_CONN_MAX_LIFETIME": "invalid",
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer func() {
				for k := range tt.envVars {
					os.Unsetenv(k)
				}
			}()

			got, err := Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Port != tt.want.Port {
					t.Errorf("Load() Port = %v, want %v", got.Port, tt.want.Port)
				}
				if got.Database != tt.want.Database {
					t.Errorf("Load() Database = %v, want %v", got.Database, tt.want.Database)
				}
				if len(got.Kafka.Brokers) != len(tt.want.Kafka.Brokers) ||
					got.Kafka.Brokers[0] != tt.want.Kafka.Brokers[0] ||
					got.Kafka.Topic != tt.want.Kafka.Topic ||
					got.Kafka.GroupID != tt.want.Kafka.GroupID {
					t.Errorf("Load() Kafka = %v, want %v", got.Kafka, tt.want.Kafka)
				}
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		want         string
	}{
		{
			name:         "env var exists",
			key:          "TEST_VAR",
			defaultValue: "default",
			envValue:     "custom",
			want:         "custom",
		},
		{
			name:         "env var does not exist",
			key:          "NON_EXISTENT_VAR",
			defaultValue: "default",
			envValue:     "",
			want:         "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			if got := getEnv(tt.key, tt.defaultValue); got != tt.want {
				t.Errorf("getEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}
