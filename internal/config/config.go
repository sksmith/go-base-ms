package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port           int
	Database       DatabaseConfig
	Kafka          KafkaConfig
	SchemaRegistry SchemaRegistryConfig
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int // in minutes
}

type KafkaConfig struct {
	Brokers          []string
	Topic            string
	GroupID          string
	SecurityProtocol string
	SaslMechanism    string
	SaslUsername     string
	SaslPassword     string
}

type SchemaRegistryConfig struct {
	URL       string
	Username  string
	Password  string
	APIKey    string
	APISecret string
}

func Load() (*Config, error) {
	port, err := strconv.Atoi(getEnv("PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid PORT: %w", err)
	}

	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	maxOpenConns, err := strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS", "25"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_OPEN_CONNS: %w", err)
	}

	maxIdleConns, err := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS", "5"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_IDLE_CONNS: %w", err)
	}

	connMaxLifetime, err := strconv.Atoi(getEnv("DB_CONN_MAX_LIFETIME", "5"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_CONN_MAX_LIFETIME: %w", err)
	}

	return &Config{
		Port: port,
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            dbPort,
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", ""),
			DBName:          getEnv("DB_NAME", "gobase"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    maxOpenConns,
			MaxIdleConns:    maxIdleConns,
			ConnMaxLifetime: connMaxLifetime,
		},
		Kafka: KafkaConfig{
			Brokers:          []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
			Topic:            getEnv("KAFKA_TOPIC", "events"),
			GroupID:          getEnv("KAFKA_GROUP_ID", "go-base-ms"),
			SecurityProtocol: getEnv("KAFKA_SECURITY_PROTOCOL", "PLAINTEXT"),
			SaslMechanism:    getEnv("KAFKA_SASL_MECHANISM", ""),
			SaslUsername:     getEnv("KAFKA_SASL_USERNAME", ""),
			SaslPassword:     getEnv("KAFKA_SASL_PASSWORD", ""),
		},
		SchemaRegistry: SchemaRegistryConfig{
			URL:       getEnv("SCHEMA_REGISTRY_URL", "http://localhost:8081"),
			Username:  getEnv("SCHEMA_REGISTRY_USERNAME", ""),
			Password:  getEnv("SCHEMA_REGISTRY_PASSWORD", ""),
			APIKey:    getEnv("SCHEMA_REGISTRY_API_KEY", ""),
			APISecret: getEnv("SCHEMA_REGISTRY_API_SECRET", ""),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
