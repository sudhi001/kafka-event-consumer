# Kafka Event Consumer Library

A simple, high-level Kafka consumer library for Go that abstracts away the complexity of Kafka configuration and provides a clean callback-based interface for processing messages.

## Features

- **Simple 2-3 line setup** - Get a Kafka consumer running with minimal code
- **Multiple topics support** - Consume from multiple Kafka topics simultaneously
- **Health monitoring** - Built-in REST API for health checks and metrics
- **Graceful shutdown** - Proper cleanup and resource management
- **Error handling** - Automatic retries and error recovery
- **Flexible configuration** - Environment variables, TOML files, or in-code configuration
- **Structured logging** - JSON-formatted logs for production environments

## Quick Start

### Environment Variables (Simplest)

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "kafka-event-consumer/internal/consumer"
)

// Your business logic - just implement this function
func handleMessage(message []byte, topic string, partition int, offset int64) error {
    var msg struct{ ID, Data string }
    if err := json.Unmarshal(message, &msg); err != nil {
        return err
    }
    
    fmt.Printf("Processing message from topic '%s': ID=%s, Data=%s\n", topic, msg.ID, msg.Data)
    return nil
}

func main() {
    // 🚀 JUST 2 LINES OF CODE! 🚀
    consumer := consumer.NewConsumer(handleMessage)
    log.Fatal(consumer.Run())
}
```

Set environment variables:
```bash
export KAFKA_BROKERS="localhost:9092"
export KAFKA_TOPICS="topic1,topic2,topic3"  # Multiple topics supported
export KAFKA_GROUP_ID="my-consumer-group"
export HEALTH_PORT="8080"
```

### TOML Configuration

```go
package main

import (
    "log"
    "kafka-event-consumer/internal/consumer"
)

func handleMessage(message []byte, topic string, partition int, offset int64) error {
    // Your business logic here
    fmt.Printf("Processing message from topic '%s': %s\n", topic, string(message))
    return nil
}

func main() {
    consumer, err := consumer.NewConsumerFromConfig("config.toml", handleMessage)
    if err != nil {
        log.Fatalf("Failed to create consumer: %v", err)
    }
    
    log.Fatal(consumer.Run())
}
```

Example `config.toml`:
```toml
[kafka]
brokers = ["localhost:9092"]
topics = ["user-events", "order-events", "system-logs"]  # Multiple topics
group_id = "my-consumer-group"
auto_offset_reset = "latest"
max_bytes = 1048576
commit_interval = "1s"
read_timeout = "10s"
max_retries = 3
retry_backoff = "1s"

[health]
enabled = true
port = 8080

[logging]
level = "info"
format = "json"
```

### In-Code Configuration

```go
package main

import (
    "log"
    "kafka-event-consumer/internal/consumer"
)

func handleMessage(message []byte, topic string, partition int, offset int64) error {
    // Your business logic here
    return nil
}

func main() {
    brokers := []string{"localhost:9092"}
    topics := []string{"topic1", "topic2", "topic3"}  // Multiple topics
    groupID := "my-consumer-group"
    
    consumer := consumer.NewSimpleConsumer(brokers, topics, groupID, handleMessage)
    log.Fatal(consumer.Run())
}
```

## Multiple Topics Support

The library now supports consuming from multiple Kafka topics simultaneously:

- **Environment Variables**: Use `KAFKA_TOPICS="topic1,topic2,topic3"` (comma-separated)
- **TOML Configuration**: Use `topics = ["topic1", "topic2", "topic3"]` (array)
- **In-Code**: Pass `[]string{"topic1", "topic2", "topic3"}` to `NewSimpleConsumer`

Each topic gets its own Kafka reader and goroutine, ensuring parallel processing of messages from different topics.

## Health Monitoring

The library provides a built-in HTTP server with the following endpoints:

- `GET /health` - Basic health status
- `GET /health/detailed` - Detailed health information with configuration
- `GET /metrics` - Prometheus-compatible metrics
- `GET /ready` - Kubernetes readiness probe
- `GET /live` - Kubernetes liveness probe

Example health response:
```json
{
  "status": "running",
  "last_message": "2024-01-15T10:30:00Z",
  "messages_processed": 1234,
  "errors": 5,
  "start_time": "2024-01-15T10:00:00Z",
  "uptime": "30m0s"
}
```

## Configuration Options

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `KAFKA_BROKERS` | `localhost:9092` | Kafka broker addresses |
| `KAFKA_TOPICS` | `default-topic` | Comma-separated list of topics |
| `KAFKA_GROUP_ID` | `default-consumer-group` | Consumer group ID |
| `KAFKA_AUTO_OFFSET_RESET` | `latest` | Offset reset strategy |
| `KAFKA_MAX_BYTES` | `1048576` | Maximum message size |
| `KAFKA_COMMIT_INTERVAL` | `1s` | Commit interval |
| `KAFKA_READ_TIMEOUT` | `10s` | Read timeout |
| `KAFKA_MAX_RETRIES` | `3` | Maximum retries |
| `KAFKA_RETRY_BACKOFF` | `1s` | Retry backoff |
| `HEALTH_ENABLED` | `true` | Enable health server |
| `HEALTH_PORT` | `8080` | Health server port |
| `LOG_LEVEL` | `info` | Log level |
| `LOG_FORMAT` | `json` | Log format |

All TOML configuration options have equivalent environment variable overrides.

## Running the Examples

```bash
# Build the examples
make build

# Run with environment variables
make run

# Run with TOML configuration
make run-toml

# Clean build artifacts
make clean
```

## Production Deployment

1. **Configuration**: Use TOML files for production configuration
2. **Health Checks**: Configure your load balancer to use `/ready` and `/live` endpoints
3. **Monitoring**: Use `/metrics` endpoint with Prometheus
4. **Logging**: Set `LOG_LEVEL=info` and `LOG_FORMAT=json` for structured logging
5. **Multiple Topics**: Configure multiple topics for different event types

Example production TOML:
```toml
[kafka]
brokers = ["kafka-prod-1:9092", "kafka-prod-2:9092", "kafka-prod-3:9092"]
topics = ["user-events", "order-events", "payment-events", "system-events"]
group_id = "production-consumer-group"
auto_offset_reset = "earliest"
max_bytes = 2097152
commit_interval = "500ms"
read_timeout = "5s"
max_retries = 5
retry_backoff = "2s"

[health]
enabled = true
port = 8080

[logging]
level = "info"
format = "json"
```

## Dependencies

- `github.com/gin-gonic/gin` - HTTP framework for health endpoints
- `github.com/segmentio/kafka-go` - Kafka client library
- `github.com/sirupsen/logrus` - Structured logging
- `github.com/BurntSushi/toml` - TOML configuration parsing

## License

MIT License 