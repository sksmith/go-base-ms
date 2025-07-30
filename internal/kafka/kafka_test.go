package kafka

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/sksmith/go-base-ms/internal/config"
)

func TestNew_InvalidBrokers(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	kafkaCfg := config.KafkaConfig{
		Brokers:          []string{"invalid:9999"},
		Topic:            "test-topic",
		GroupID:          "test-group",
		SecurityProtocol: "PLAINTEXT",
	}

	srCfg := config.SchemaRegistryConfig{
		URL: "", // Skip schema registry for this test
	}

	// This should not fail immediately as Confluent's client doesn't validate brokers on creation
	client, err := New(kafkaCfg, srCfg, logger)
	if err != nil {
		t.Errorf("expected New() to succeed with invalid brokers, got error: %v", err)
	}

	if client != nil {
		defer client.Close()

		// However, Ping should fail
		ctx := context.Background()
		if err := client.Ping(ctx); err == nil {
			t.Error("expected Ping() to fail with invalid brokers")
		}
	}
}

func TestClient_CloseIdempotent(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	kafkaCfg := config.KafkaConfig{
		Brokers:          []string{"localhost:9092"},
		Topic:            "test-topic",
		GroupID:          "test-group",
		SecurityProtocol: "PLAINTEXT",
	}

	srCfg := config.SchemaRegistryConfig{
		URL: "", // Skip schema registry for this test
	}

	client, err := New(kafkaCfg, srCfg, logger)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Close multiple times should not panic
	err1 := client.Close()
	err2 := client.Close()

	if err1 != nil {
		t.Errorf("first Close() returned error: %v", err1)
	}
	if err2 != nil {
		t.Errorf("second Close() returned error: %v", err2)
	}
}

func TestClient_ClosedOperations(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	kafkaCfg := config.KafkaConfig{
		Brokers:          []string{"localhost:9092"},
		Topic:            "test-topic",
		GroupID:          "test-group",
		SecurityProtocol: "PLAINTEXT",
	}

	srCfg := config.SchemaRegistryConfig{
		URL: "", // Skip schema registry for this test
	}

	client, err := New(kafkaCfg, srCfg, logger)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Close the client
	client.Close()

	ctx := context.Background()

	// Operations on closed client should fail
	if err := client.Ping(ctx); err == nil {
		t.Error("expected Ping() to fail on closed client")
	}

	msg := Message{
		Key:   []byte("test-key"),
		Value: []byte("test-value"),
		Topic: "test-topic",
	}
	if err := client.SendMessage(ctx, msg); err == nil {
		t.Error("expected SendMessage() to fail on closed client")
	}
}

func TestMessage_Headers(t *testing.T) {
	msg := Message{
		Key:   []byte("test-key"),
		Value: []byte("test-value"),
		Topic: "test-topic",
		Headers: map[string][]byte{
			"header1": []byte("value1"),
			"header2": []byte("value2"),
		},
	}

	if len(msg.Headers) != 2 {
		t.Errorf("expected 2 headers, got %d", len(msg.Headers))
	}

	if string(msg.Headers["header1"]) != "value1" {
		t.Errorf("expected header1 to be 'value1', got %s", string(msg.Headers["header1"]))
	}
}

func TestKafkaConfig_SecuritySettings(t *testing.T) {
	tests := []struct {
		name             string
		securityProtocol string
		saslMechanism    string
		username         string
		password         string
		expectValid      bool
	}{
		{
			name:             "plaintext",
			securityProtocol: "PLAINTEXT",
			expectValid:      true,
		},
		{
			name:             "sasl_ssl with credentials",
			securityProtocol: "SASL_SSL",
			saslMechanism:    "PLAIN",
			username:         "user",
			password:         "pass",
			expectValid:      true,
		},
		{
			name:             "sasl_ssl without credentials",
			securityProtocol: "SASL_SSL",
			saslMechanism:    "PLAIN",
			expectValid:      false, // Client creation fails without credentials
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	srCfg := config.SchemaRegistryConfig{URL: ""}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kafkaCfg := config.KafkaConfig{
				Brokers:          []string{"localhost:9092"},
				Topic:            "test-topic",
				GroupID:          "test-group",
				SecurityProtocol: tt.securityProtocol,
				SaslMechanism:    tt.saslMechanism,
				SaslUsername:     tt.username,
				SaslPassword:     tt.password,
			}

			client, err := New(kafkaCfg, srCfg, logger)
			if (err == nil) != tt.expectValid {
				t.Errorf("expected valid=%v, got error=%v", tt.expectValid, err)
			}

			if client != nil {
				defer client.Close()
			}
		})
	}
}

func TestSchemaRegistryConfig(t *testing.T) {
	tests := []struct {
		name   string
		config config.SchemaRegistryConfig
		expect bool // true if schema registry should be initialized
	}{
		{
			name: "empty URL",
			config: config.SchemaRegistryConfig{
				URL: "",
			},
			expect: false,
		},
		{
			name: "with URL",
			config: config.SchemaRegistryConfig{
				URL: "http://localhost:8081",
			},
			expect: true,
		},
		{
			name: "with basic auth",
			config: config.SchemaRegistryConfig{
				URL:      "http://localhost:8081",
				Username: "user",
				Password: "pass",
			},
			expect: true,
		},
		{
			name: "with API key",
			config: config.SchemaRegistryConfig{
				URL:       "http://localhost:8081",
				APIKey:    "key",
				APISecret: "secret",
			},
			expect: true,
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	kafkaCfg := config.KafkaConfig{
		Brokers:          []string{"localhost:9092"},
		Topic:            "test-topic",
		GroupID:          "test-group",
		SecurityProtocol: "PLAINTEXT",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(kafkaCfg, tt.config, logger)
			if err != nil {
				if tt.expect {
					t.Errorf("expected schema registry to be initialized, got error: %v", err)
				}
				return
			}
			defer client.Close()

			hasSchemaRegistry := client.GetSchemaRegistry() != nil
			if hasSchemaRegistry != tt.expect {
				t.Errorf("expected schema registry initialized=%v, got=%v", tt.expect, hasSchemaRegistry)
			}

			hasSerializer := client.GetAvroSerializer() != nil
			if tt.expect && !hasSerializer {
				t.Error("expected avro serializer to be initialized when schema registry is configured")
			}

			hasDeserializer := client.GetAvroDeserializer() != nil
			if tt.expect && !hasDeserializer {
				t.Error("expected avro deserializer to be initialized when schema registry is configured")
			}
		})
	}
}
