package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/avro"
	"github.com/sksmith/go-base-ms/internal/config"
)

type Client struct {
	producer         *kafka.Producer
	consumer         *kafka.Consumer
	schemaRegistry   schemaregistry.Client
	avroSerializer   *avro.GenericSerializer
	avroDeserializer *avro.GenericDeserializer
	logger           *slog.Logger
	cfg              config.KafkaConfig
	srCfg            config.SchemaRegistryConfig
	mu               sync.RWMutex
	closed           bool
}

type Message struct {
	Key     []byte
	Value   []byte
	Headers map[string][]byte
	Topic   string
}

type MessageHandler func(Message) error

func New(kafkaCfg config.KafkaConfig, srCfg config.SchemaRegistryConfig, logger *slog.Logger) (*Client, error) {
	client := &Client{
		logger: logger,
		cfg:    kafkaCfg,
		srCfg:  srCfg,
	}

	// Initialize Schema Registry client
	if err := client.initSchemaRegistry(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema registry: %w", err)
	}

	// Initialize Kafka producer
	if err := client.initProducer(); err != nil {
		return nil, fmt.Errorf("failed to initialize producer: %w", err)
	}

	// Initialize Kafka consumer
	if err := client.initConsumer(); err != nil {
		return nil, fmt.Errorf("failed to initialize consumer: %w", err)
	}

	return client, nil
}

func (c *Client) initSchemaRegistry() error {
	if c.srCfg.URL == "" {
		c.logger.Warn("schema registry URL not configured, skipping initialization")
		return nil
	}

	srConfig := schemaregistry.NewConfig(c.srCfg.URL)

	// Configure authentication
	if c.srCfg.Username != "" && c.srCfg.Password != "" {
		srConfig.BasicAuthUserInfo = fmt.Sprintf("%s:%s", c.srCfg.Username, c.srCfg.Password)
	}
	if c.srCfg.APIKey != "" && c.srCfg.APISecret != "" {
		srConfig.BasicAuthUserInfo = fmt.Sprintf("%s:%s", c.srCfg.APIKey, c.srCfg.APISecret)
	}

	var err error
	c.schemaRegistry, err = schemaregistry.NewClient(srConfig)
	if err != nil {
		return fmt.Errorf("failed to create schema registry client: %w", err)
	}

	// Initialize Avro serializer/deserializer
	serConfig := avro.NewSerializerConfig()
	c.avroSerializer, err = avro.NewGenericSerializer(c.schemaRegistry, serde.ValueSerde, serConfig)
	if err != nil {
		return fmt.Errorf("failed to create avro serializer: %w", err)
	}

	deserConfig := avro.NewDeserializerConfig()
	c.avroDeserializer, err = avro.NewGenericDeserializer(c.schemaRegistry, serde.ValueSerde, deserConfig)
	if err != nil {
		return fmt.Errorf("failed to create avro deserializer: %w", err)
	}

	c.logger.Info("schema registry initialized", "url", c.srCfg.URL)
	return nil
}

func (c *Client) initProducer() error {
	configMap := kafka.ConfigMap{
		"bootstrap.servers":                     strings.Join(c.cfg.Brokers, ","),
		"client.id":                             "go-base-ms-producer",
		"acks":                                  "all",
		"retries":                               2147483647,
		"max.in.flight.requests.per.connection": 5,
		"enable.idempotence":                    true,
	}

	// Add security configuration
	if c.cfg.SecurityProtocol != "PLAINTEXT" {
		configMap["security.protocol"] = c.cfg.SecurityProtocol
		if c.cfg.SaslMechanism != "" {
			configMap["sasl.mechanism"] = c.cfg.SaslMechanism
			if c.cfg.SaslUsername != "" && c.cfg.SaslPassword != "" {
				configMap["sasl.username"] = c.cfg.SaslUsername
				configMap["sasl.password"] = c.cfg.SaslPassword
			}
		}
	}

	var err error
	c.producer, err = kafka.NewProducer(&configMap)
	if err != nil {
		return fmt.Errorf("failed to create producer: %w", err)
	}

	// Start delivery report goroutine
	go c.handleDeliveryReports()

	c.logger.Info("kafka producer initialized", "brokers", c.cfg.Brokers)
	return nil
}

func (c *Client) initConsumer() error {
	configMap := kafka.ConfigMap{
		"bootstrap.servers":  strings.Join(c.cfg.Brokers, ","),
		"client.id":          "go-base-ms-consumer",
		"group.id":           c.cfg.GroupID,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
	}

	// Add security configuration
	if c.cfg.SecurityProtocol != "PLAINTEXT" {
		configMap["security.protocol"] = c.cfg.SecurityProtocol
		if c.cfg.SaslMechanism != "" {
			configMap["sasl.mechanism"] = c.cfg.SaslMechanism
			if c.cfg.SaslUsername != "" && c.cfg.SaslPassword != "" {
				configMap["sasl.username"] = c.cfg.SaslUsername
				configMap["sasl.password"] = c.cfg.SaslPassword
			}
		}
	}

	var err error
	c.consumer, err = kafka.NewConsumer(&configMap)
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}

	c.logger.Info("kafka consumer initialized", "group_id", c.cfg.GroupID)
	return nil
}

func (c *Client) handleDeliveryReports() {
	for e := range c.producer.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				c.logger.Error("delivery failed",
					"topic", *ev.TopicPartition.Topic,
					"partition", ev.TopicPartition.Partition,
					"error", ev.TopicPartition.Error)
			} else {
				c.logger.Debug("message delivered",
					"topic", *ev.TopicPartition.Topic,
					"partition", ev.TopicPartition.Partition,
					"offset", ev.TopicPartition.Offset)
			}
		}
	}
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true

	if c.producer != nil {
		c.producer.Close()
	}
	if c.consumer != nil {
		c.consumer.Close()
	}

	c.logger.Info("kafka client closed")
	return nil
}

func (c *Client) Ping(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return fmt.Errorf("client is closed")
	}

	if c.producer == nil {
		return fmt.Errorf("producer not initialized")
	}

	// Get metadata to check connection
	metadata, err := c.producer.GetMetadata(nil, false, 5000)
	if err != nil {
		return fmt.Errorf("failed to get metadata: %w", err)
	}

	if len(metadata.Brokers) == 0 {
		return fmt.Errorf("no brokers available")
	}

	return nil
}

func (c *Client) SendMessage(ctx context.Context, msg Message) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return fmt.Errorf("client is closed")
	}

	if c.producer == nil {
		return fmt.Errorf("producer not initialized")
	}

	topic := msg.Topic
	if topic == "" {
		topic = c.cfg.Topic
	}

	kafkaMsg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            msg.Key,
		Value:          msg.Value,
	}

	// Add headers if provided
	if msg.Headers != nil {
		kafkaMsg.Headers = make([]kafka.Header, 0, len(msg.Headers))
		for key, value := range msg.Headers {
			kafkaMsg.Headers = append(kafkaMsg.Headers, kafka.Header{
				Key:   key,
				Value: value,
			})
		}
	}

	// Send message
	deliveryChan := make(chan kafka.Event)
	err := c.producer.Produce(kafkaMsg, deliveryChan)
	if err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	// Wait for delivery report with timeout
	select {
	case e := <-deliveryChan:
		if m, ok := e.(*kafka.Message); ok {
			if m.TopicPartition.Error != nil {
				return fmt.Errorf("message delivery failed: %w", m.TopicPartition.Error)
			}
			c.logger.Debug("message sent successfully",
				"topic", topic,
				"partition", m.TopicPartition.Partition,
				"offset", m.TopicPartition.Offset)
		}
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(30 * time.Second):
		return fmt.Errorf("message delivery timeout")
	}

	return nil
}

func (c *Client) SendAvroMessage(ctx context.Context, topic string, key []byte, value interface{}, subject string) error {
	if c.avroSerializer == nil {
		return fmt.Errorf("avro serializer not initialized")
	}

	serializedValue, err := c.avroSerializer.Serialize(subject, value)
	if err != nil {
		return fmt.Errorf("failed to serialize avro message: %w", err)
	}

	return c.SendMessage(ctx, Message{
		Topic: topic,
		Key:   key,
		Value: serializedValue,
	})
}

func (c *Client) ConsumeMessages(ctx context.Context, handler MessageHandler) error {
	c.mu.RLock()
	consumer := c.consumer
	topic := c.cfg.Topic
	c.mu.RUnlock()

	if consumer == nil {
		return fmt.Errorf("consumer not initialized")
	}

	// Subscribe to topic
	err := consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
	}

	c.logger.Info("started consuming messages", "topic", topic, "group_id", c.cfg.GroupID)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("stopping message consumption")
			return ctx.Err()
		default:
			msg, err := consumer.ReadMessage(1000) // 1 second timeout
			if err != nil {
				if kafkaErr, ok := err.(kafka.Error); ok && kafkaErr.Code() == kafka.ErrTimedOut {
					continue // Timeout is expected, continue polling
				}
				c.logger.Error("failed to read message", "error", err)
				continue
			}

			// Convert kafka message to our Message type
			ourMsg := Message{
				Topic: *msg.TopicPartition.Topic,
				Key:   msg.Key,
				Value: msg.Value,
			}

			// Add headers if present
			if len(msg.Headers) > 0 {
				ourMsg.Headers = make(map[string][]byte)
				for _, header := range msg.Headers {
					ourMsg.Headers[header.Key] = header.Value
				}
			}

			// Process message
			if err := handler(ourMsg); err != nil {
				c.logger.Error("message handler failed",
					"topic", *msg.TopicPartition.Topic,
					"partition", msg.TopicPartition.Partition,
					"offset", msg.TopicPartition.Offset,
					"error", err)
				continue
			}

			// Commit message
			if _, err := consumer.CommitMessage(msg); err != nil {
				c.logger.Error("failed to commit message",
					"topic", *msg.TopicPartition.Topic,
					"partition", msg.TopicPartition.Partition,
					"offset", msg.TopicPartition.Offset,
					"error", err)
			}
		}
	}
}

func (c *Client) GetSchemaRegistry() schemaregistry.Client {
	return c.schemaRegistry
}

func (c *Client) GetAvroSerializer() *avro.GenericSerializer {
	return c.avroSerializer
}

func (c *Client) GetAvroDeserializer() *avro.GenericDeserializer {
	return c.avroDeserializer
}
