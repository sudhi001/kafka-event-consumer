package consumer

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

// MessageHandler is the callback function type for processing Kafka messages
type MessageHandler func(message []byte, topic string, partition int, offset int64) error

// ConsumerConfig holds the configuration for the Kafka consumer
type ConsumerConfig struct {
	Brokers           []string
	Topics            []string // Changed from single Topic to multiple Topics
	GroupID           string
	AutoOffsetReset   string // "earliest" or "latest"
	MaxBytes          int    // Maximum message size in bytes
	CommitInterval    time.Duration
	ReadBatchTimeout  time.Duration
	MaxRetries        int
	RetryBackoff      time.Duration
	HealthCheckPort   int
	EnableHealthCheck bool
}

// KafkaEventConsumer is the main consumer framework
type KafkaEventConsumer struct {
	config       *ConsumerConfig
	readers      []*kafka.Reader
	handler      MessageHandler
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	isRunning    bool
	mu           sync.RWMutex
	healthStatus *HealthStatus
	logger       *logrus.Logger
}

// HealthStatus represents the health status of the consumer
type HealthStatus struct {
	Status            string    `json:"status"`
	LastMessage       time.Time `json:"last_message"`
	MessagesProcessed int64     `json:"messages_processed"`
	Errors            int64     `json:"errors"`
	StartTime         time.Time `json:"start_time"`
	mu                sync.RWMutex
}

// NewConsumerFromConfig creates a consumer from TOML configuration file
func NewConsumerFromConfig(configPath string, handler MessageHandler) (*KafkaEventConsumer, error) {
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return NewKafkaEventConsumer(config.ToConsumerConfig(), handler), nil
}

// NewConsumer creates a consumer with environment variable configuration
func NewConsumer(handler MessageHandler) *KafkaEventConsumer {
	brokers := getEnvOrDefault("KAFKA_BROKERS", "localhost:9092")
	topics := getEnvOrDefault("KAFKA_TOPICS", "default-topic")
	groupID := getEnvOrDefault("KAFKA_GROUP_ID", "default-consumer-group")
	port := getEnvOrDefault("HEALTH_PORT", "8080")

	// Split comma-separated topics
	topicList := splitAndTrim(topics, ",")

	config := &ConsumerConfig{
		Brokers:           []string{brokers},
		Topics:            topicList,
		GroupID:           groupID,
		AutoOffsetReset:   "latest",
		MaxBytes:          1048576,
		CommitInterval:    1 * time.Second,
		ReadBatchTimeout:  10 * time.Second,
		MaxRetries:        3,
		RetryBackoff:      1 * time.Second,
		HealthCheckPort:   parsePort(port),
		EnableHealthCheck: true,
	}

	return NewKafkaEventConsumer(config, handler)
}

// NewSimpleConsumer creates a new simple consumer with minimal configuration
func NewSimpleConsumer(brokers []string, topics []string, groupID string, handler MessageHandler) *KafkaEventConsumer {
	config := &ConsumerConfig{
		Brokers:           brokers,
		Topics:            topics,
		GroupID:           groupID,
		AutoOffsetReset:   "latest",
		MaxBytes:          1048576, // 1MB
		CommitInterval:    1 * time.Second,
		ReadBatchTimeout:  10 * time.Second,
		MaxRetries:        3,
		RetryBackoff:      1 * time.Second,
		HealthCheckPort:   8080,
		EnableHealthCheck: true,
	}

	return NewKafkaEventConsumer(config, handler)
}

// NewKafkaEventConsumer creates a new KafkaEventConsumer instance
func NewKafkaEventConsumer(config *ConsumerConfig, handler MessageHandler) *KafkaEventConsumer {
	if config == nil {
		config = &ConsumerConfig{
			AutoOffsetReset:   "latest",
			MaxBytes:          1048576, // 1MB
			CommitInterval:    1 * time.Second,
			ReadBatchTimeout:  10 * time.Second,
			MaxRetries:        3,
			RetryBackoff:      1 * time.Second,
			HealthCheckPort:   8080,
			EnableHealthCheck: true,
		}
	}

	if handler == nil {
		handler = func(message []byte, topic string, partition int, offset int64) error {
			fmt.Printf("Default handler: Received message from topic %s, partition %d, offset %d: %s\n",
				topic, partition, offset, string(message))
			return nil
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	consumer := &KafkaEventConsumer{
		config:  config,
		handler: handler,
		ctx:     ctx,
		cancel:  cancel,
		healthStatus: &HealthStatus{
			Status:            "stopped",
			LastMessage:       time.Time{},
			MessagesProcessed: 0,
			Errors:            0,
			StartTime:         time.Time{},
		},
		logger: logrus.New(),
	}

	// Configure logger
	consumer.logger.SetFormatter(&logrus.JSONFormatter{})
	consumer.logger.SetLevel(logrus.InfoLevel)

	return consumer
}

// StartAndWait starts the consumer and waits for shutdown signal
func (kec *KafkaEventConsumer) StartAndWait() error {
	if err := kec.Start(); err != nil {
		return fmt.Errorf("failed to start consumer: %w", err)
	}

	fmt.Printf("✅ Kafka consumer started successfully!\n")
	fmt.Printf("📊 Health check: http://localhost:%d/health\n", kec.config.HealthCheckPort)
	fmt.Printf("📈 Metrics: http://localhost:%d/metrics\n", kec.config.HealthCheckPort)
	fmt.Printf("🔄 Topics: %v, Group: %s\n", kec.config.Topics, kec.config.GroupID)
	fmt.Printf("⏹️  Press Ctrl+C to stop\n\n")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n🛑 Shutting down Kafka consumer...")
	return kec.Stop()
}

// Run starts the consumer and waits for shutdown (one-liner)
func (kec *KafkaEventConsumer) Run() error {
	return kec.StartAndWait()
}

// Start begins consuming messages from Kafka
func (kec *KafkaEventConsumer) Start() error {
	kec.mu.Lock()
	defer kec.mu.Unlock()

	if kec.isRunning {
		return fmt.Errorf("consumer is already running")
	}

	// Create Kafka readers for each topic
	kec.readers = make([]*kafka.Reader, 0, len(kec.config.Topics))

	for _, topic := range kec.config.Topics {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:          kec.config.Brokers,
			Topic:            topic,
			GroupID:          kec.config.GroupID,
			MinBytes:         10e3, // 10KB
			MaxBytes:         kec.config.MaxBytes,
			CommitInterval:   kec.config.CommitInterval,
			ReadBatchTimeout: kec.config.ReadBatchTimeout,
		})

		// Set offset based on configuration
		switch kec.config.AutoOffsetReset {
		case "earliest":
			reader.SetOffset(kafka.FirstOffset)
		case "latest":
			reader.SetOffset(kafka.LastOffset)
		default:
			reader.SetOffset(kafka.LastOffset)
		}

		kec.readers = append(kec.readers, reader)
	}

	kec.isRunning = true
	kec.healthStatus.mu.Lock()
	kec.healthStatus.Status = "running"
	kec.healthStatus.StartTime = time.Now()
	kec.healthStatus.mu.Unlock()

	kec.logger.Info("Kafka consumer started", map[string]interface{}{
		"topics":  kec.config.Topics,
		"brokers": kec.config.Brokers,
		"groupID": kec.config.GroupID,
	})

	// Start health check server if enabled
	if kec.config.EnableHealthCheck {
		kec.wg.Add(1)
		go func() {
			defer kec.wg.Done()
			kec.startHealthServer()
		}()
	}

	// Start consuming messages from all topics
	for i, reader := range kec.readers {
		topic := kec.config.Topics[i]
		kec.wg.Add(1)
		go func(r *kafka.Reader, t string) {
			defer kec.wg.Done()
			kec.consumeMessagesFromReader(r, t)
		}(reader, topic)
	}

	return nil
}

// Stop gracefully stops the consumer
func (kec *KafkaEventConsumer) Stop() error {
	kec.mu.Lock()
	defer kec.mu.Unlock()

	if !kec.isRunning {
		return fmt.Errorf("consumer is not running")
	}

	kec.logger.Info("Stopping Kafka consumer...")

	// Cancel context to stop all goroutines
	kec.cancel()

	// Close all readers
	for _, reader := range kec.readers {
		if err := reader.Close(); err != nil {
			kec.logger.Error("Error closing Kafka reader: ", err)
		}
	}

	kec.isRunning = false
	kec.healthStatus.mu.Lock()
	kec.healthStatus.Status = "stopped"
	kec.healthStatus.mu.Unlock()

	// Wait for all goroutines to finish
	kec.wg.Wait()

	kec.logger.Info("Kafka consumer stopped")
	return nil
}

// consumeMessagesFromReader handles the main message consumption loop for a specific reader
func (kec *KafkaEventConsumer) consumeMessagesFromReader(reader *kafka.Reader, topic string) {
	for {
		select {
		case <-kec.ctx.Done():
			kec.logger.Info("Consumer context cancelled, stopping message consumption for topic: ", topic)
			return
		default:
			// Read message with timeout
			ctx, cancel := context.WithTimeout(kec.ctx, kec.config.ReadBatchTimeout)
			message, err := reader.ReadMessage(ctx)
			cancel()

			if err != nil {
				if err == context.DeadlineExceeded || err == context.Canceled {
					continue
				}

				kec.logger.Error("Error reading message from topic ", topic, ": ", err)
				kec.healthStatus.mu.Lock()
				kec.healthStatus.Errors++
				kec.healthStatus.mu.Unlock()

				// Retry logic
				time.Sleep(kec.config.RetryBackoff)
				continue
			}

			// Process message
			if err := kec.processMessage(message); err != nil {
				kec.logger.Error("Error processing message from topic ", topic, ": ", err)
				kec.healthStatus.mu.Lock()
				kec.healthStatus.Errors++
				kec.healthStatus.mu.Unlock()
			}
		}
	}
}

// processMessage processes a single Kafka message
func (kec *KafkaEventConsumer) processMessage(message kafka.Message) error {
	start := time.Now()

	// Update health status
	kec.healthStatus.mu.Lock()
	kec.healthStatus.LastMessage = time.Now()
	kec.healthStatus.MessagesProcessed++
	kec.healthStatus.mu.Unlock()

	// Call the message handler
	err := kec.handler(message.Value, message.Topic, message.Partition, message.Offset)

	if err != nil {
		kec.logger.Error("Handler error", map[string]interface{}{
			"error":     err,
			"topic":     message.Topic,
			"partition": message.Partition,
			"offset":    message.Offset,
		})
		return err
	}

	kec.logger.Debug("Message processed successfully", map[string]interface{}{
		"topic":     message.Topic,
		"partition": message.Partition,
		"offset":    message.Offset,
		"duration":  time.Since(start),
	})

	return nil
}

// GetHealthStatus returns the current health status
func (kec *KafkaEventConsumer) GetHealthStatus() *HealthStatus {
	kec.healthStatus.mu.RLock()
	defer kec.healthStatus.mu.RUnlock()

	// Return a copy to avoid race conditions
	return &HealthStatus{
		Status:            kec.healthStatus.Status,
		LastMessage:       kec.healthStatus.LastMessage,
		MessagesProcessed: kec.healthStatus.MessagesProcessed,
		Errors:            kec.healthStatus.Errors,
		StartTime:         kec.healthStatus.StartTime,
	}
}

// IsRunning returns whether the consumer is currently running
func (kec *KafkaEventConsumer) IsRunning() bool {
	kec.mu.RLock()
	defer kec.mu.RUnlock()
	return kec.isRunning
}

// Helper function to get environment variable with default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Helper function to parse port
func parsePort(port string) int {
	if port == "" {
		return 8080
	}

	var portNum int
	fmt.Sscanf(port, "%d", &portNum)
	if portNum == 0 {
		return 8080
	}
	return portNum
}
