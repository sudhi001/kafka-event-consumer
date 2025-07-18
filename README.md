# KafkaEventConsumer Library

A simple, powerful Kafka consumer library for Go applications. This library abstracts away all Kafka complexity and provides a clean, minimal interface for consuming Kafka messages with built-in health monitoring.

## Features

- **Ultra-Simple**: Just 2 lines of code to get started
- **Automatic Health Monitoring**: Built-in REST API for health checks and metrics
- **Environment Configuration**: Configure via environment variables
- **Graceful Shutdown**: Proper cleanup and shutdown handling
- **Error Handling**: Robust error handling with retry mechanisms
- **Production Ready**: Ready for production deployment

## Quick Start

### 🚀 Ultra-Simple Usage (2 Lines of Code!)

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
    json.Unmarshal(message, &msg)
    fmt.Printf("Processing: %s\n", msg.Data)
    return nil
}

func main() {
    // 🚀 JUST 2 LINES OF CODE! 🚀
    consumer := consumer.NewConsumer(handleMessage)
    log.Fatal(consumer.Run())
}
```

### Configuration via Environment Variables

```bash
export KAFKA_BROKERS=localhost:9092
export KAFKA_TOPIC=my-topic
export KAFKA_GROUP_ID=my-consumer-group
export HEALTH_PORT=8080
```

### Alternative: In-Code Configuration

```go
func main() {
    consumer := consumer.NewSimpleConsumer(
        []string{"localhost:9092"}, 
        "my-topic", 
        "my-consumer-group", 
        handleMessage,
    )
    log.Fatal(consumer.Run())
}
```

### TOML Configuration File

```go
func main() {
    // Load configuration from TOML file
    consumer, err := consumer.NewConsumerFromConfig("config.toml", handleMessage)
    if err != nil {
        log.Fatal(err)
    }
    log.Fatal(consumer.Run())
}
```

**Example TOML configuration (`config.toml`):**
```toml
[kafka]
brokers = ["localhost:9092", "localhost:9093"]
topic = "my-events"
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

## Health Monitoring

The library automatically provides health monitoring endpoints:

- **Health Check**: `http://localhost:8080/health`
- **Detailed Health**: `http://localhost:8080/health/detailed`
- **Metrics**: `http://localhost:8080/metrics`
- **Ready Probe**: `http://localhost:8080/ready`
- **Live Probe**: `http://localhost:8080/live`

### Example Health Response

```json
{
  "status": "running",
  "last_message": "2024-01-15T10:30:00Z",
  "messages_processed": 1234,
  "errors": 5,
  "start_time": "2024-01-15T09:00:00Z",
  "uptime": "1h30m0s"
}
```

## Message Handler Function

Your message handler function should have this signature:

```go
func(message []byte, topic string, partition int, offset int64) error
```

**Parameters:**
- `message`: Raw message bytes from Kafka
- `topic`: Topic name the message came from
- `partition`: Partition number
- `offset`: Message offset

**Return Value:**
- Return `nil` for successful processing
- Return an error to indicate processing failure (will be logged and counted in metrics)

## Running the Examples

### Environment Variables Example
1. **Install dependencies**:
   ```bash
   go mod tidy
   ```

2. **Set environment variables**:
   ```bash
   export KAFKA_BROKERS=localhost:9092
   export KAFKA_TOPIC=my-topic
   export KAFKA_GROUP_ID=my-consumer-group
   ```

3. **Run the example**:
   ```bash
   go run cmd/simple/main.go
   ```

### TOML Configuration Example
1. **Copy the example configuration**:
   ```bash
   cp config.toml my-config.toml
   # Edit my-config.toml with your settings
   ```

2. **Run with TOML config**:
   ```bash
   go run cmd/toml-example/main.go
   ```

## Docker Support

### Build and Run
```bash
docker build -t kafka-consumer .
docker run -p 8080:8080 kafka-consumer
```

### With Docker Compose
```bash
docker-compose up
```

## Production Deployment

### Environment Variables
```bash
KAFKA_BROKERS=kafka1:9092,kafka2:9092,kafka3:9092
KAFKA_TOPIC=production-events
KAFKA_GROUP_ID=production-consumer-group
HEALTH_PORT=8080
```

### Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kafka-consumer
spec:
  replicas: 3
  selector:
    matchLabels:
      app: kafka-consumer
  template:
    metadata:
      labels:
        app: kafka-consumer
    spec:
      containers:
      - name: consumer
        image: kafka-consumer:latest
        ports:
        - containerPort: 8080
        env:
        - name: KAFKA_BROKERS
          value: "kafka1:9092,kafka2:9092,kafka3:9092"
        - name: KAFKA_TOPIC
          value: "production-events"
        - name: KAFKA_GROUP_ID
          value: "production-consumer-group"
        livenessProbe:
          httpGet:
            path: /live
            port: 8080
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
```

## Dependencies

- `github.com/segmentio/kafka-go` - Kafka client library
- `github.com/gin-gonic/gin` - HTTP framework for health endpoints
- `github.com/sirupsen/logrus` - Structured logging

## License

This project is licensed under the MIT License. 