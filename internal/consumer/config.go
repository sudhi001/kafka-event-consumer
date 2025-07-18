package consumer

import (
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

// Config represents the TOML configuration structure
type Config struct {
	Kafka   KafkaConfig   `toml:"kafka"`
	Health  HealthConfig  `toml:"health"`
	Logging LoggingConfig `toml:"logging"`
}

// KafkaConfig represents Kafka-specific configuration
type KafkaConfig struct {
	Brokers         []string      `toml:"brokers"`
	Topic           string        `toml:"topic"`
	GroupID         string        `toml:"group_id"`
	AutoOffsetReset string        `toml:"auto_offset_reset"`
	MaxBytes        int           `toml:"max_bytes"`
	CommitInterval  time.Duration `toml:"commit_interval"`
	ReadTimeout     time.Duration `toml:"read_timeout"`
	MaxRetries      int           `toml:"max_retries"`
	RetryBackoff    time.Duration `toml:"retry_backoff"`
}

// HealthConfig represents health monitoring configuration
type HealthConfig struct {
	Enabled bool `toml:"enabled"`
	Port    int  `toml:"port"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level  string `toml:"level"`
	Format string `toml:"format"`
}

// LoadConfig loads configuration from TOML file or uses defaults
func LoadConfig(configPath string) (*Config, error) {
	config := &Config{
		Kafka: KafkaConfig{
			Brokers:         []string{"localhost:9092"},
			Topic:           "default-topic",
			GroupID:         "default-consumer-group",
			AutoOffsetReset: "latest",
			MaxBytes:        1048576, // 1MB
			CommitInterval:  1 * time.Second,
			ReadTimeout:     10 * time.Second,
			MaxRetries:      3,
			RetryBackoff:    1 * time.Second,
		},
		Health: HealthConfig{
			Enabled: true,
			Port:    8080,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}

	// Try to load from TOML file if path is provided
	if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			data, err := os.ReadFile(configPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}

			if err := toml.Unmarshal(data, config); err != nil {
				return nil, fmt.Errorf("failed to parse TOML config: %w", err)
			}
		}
	}

	// Override with environment variables if they exist
	config.overrideWithEnvVars()

	return config, nil
}

// overrideWithEnvVars overrides configuration with environment variables
func (c *Config) overrideWithEnvVars() {
	// Kafka configuration
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		c.Kafka.Brokers = []string{brokers}
	}
	if topic := os.Getenv("KAFKA_TOPIC"); topic != "" {
		c.Kafka.Topic = topic
	}
	if groupID := os.Getenv("KAFKA_GROUP_ID"); groupID != "" {
		c.Kafka.GroupID = groupID
	}
	if autoOffsetReset := os.Getenv("KAFKA_AUTO_OFFSET_RESET"); autoOffsetReset != "" {
		c.Kafka.AutoOffsetReset = autoOffsetReset
	}
	if maxBytes := os.Getenv("KAFKA_MAX_BYTES"); maxBytes != "" {
		if parsed, err := parseInt(maxBytes); err == nil {
			c.Kafka.MaxBytes = parsed
		}
	}
	if commitInterval := os.Getenv("KAFKA_COMMIT_INTERVAL"); commitInterval != "" {
		if parsed, err := time.ParseDuration(commitInterval); err == nil {
			c.Kafka.CommitInterval = parsed
		}
	}
	if readTimeout := os.Getenv("KAFKA_READ_TIMEOUT"); readTimeout != "" {
		if parsed, err := time.ParseDuration(readTimeout); err == nil {
			c.Kafka.ReadTimeout = parsed
		}
	}
	if maxRetries := os.Getenv("KAFKA_MAX_RETRIES"); maxRetries != "" {
		if parsed, err := parseInt(maxRetries); err == nil {
			c.Kafka.MaxRetries = parsed
		}
	}
	if retryBackoff := os.Getenv("KAFKA_RETRY_BACKOFF"); retryBackoff != "" {
		if parsed, err := time.ParseDuration(retryBackoff); err == nil {
			c.Kafka.RetryBackoff = parsed
		}
	}

	// Health configuration
	if healthEnabled := os.Getenv("HEALTH_ENABLED"); healthEnabled != "" {
		c.Health.Enabled = healthEnabled == "true"
	}
	if healthPort := os.Getenv("HEALTH_PORT"); healthPort != "" {
		if parsed, err := parseInt(healthPort); err == nil {
			c.Health.Port = parsed
		}
	}

	// Logging configuration
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		c.Logging.Level = logLevel
	}
	if logFormat := os.Getenv("LOG_FORMAT"); logFormat != "" {
		c.Logging.Format = logFormat
	}
}

// ToConsumerConfig converts Config to ConsumerConfig
func (c *Config) ToConsumerConfig() *ConsumerConfig {
	return &ConsumerConfig{
		Brokers:           c.Kafka.Brokers,
		Topic:             c.Kafka.Topic,
		GroupID:           c.Kafka.GroupID,
		AutoOffsetReset:   c.Kafka.AutoOffsetReset,
		MaxBytes:          c.Kafka.MaxBytes,
		CommitInterval:    c.Kafka.CommitInterval,
		ReadBatchTimeout:  c.Kafka.ReadTimeout,
		MaxRetries:        c.Kafka.MaxRetries,
		RetryBackoff:      c.Kafka.RetryBackoff,
		HealthCheckPort:   c.Health.Port,
		EnableHealthCheck: c.Health.Enabled,
	}
}

// Helper function to parse integer from string
func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}
