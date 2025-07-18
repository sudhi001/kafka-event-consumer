package consumer

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// startHealthServer starts the HTTP health check server
func (kec *KafkaEventConsumer) startHealthServer() {
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		status := kec.GetHealthStatus()

		// Determine HTTP status code based on consumer status
		httpStatus := http.StatusOK
		if status.Status != "running" {
			httpStatus = http.StatusServiceUnavailable
		}

		c.JSON(httpStatus, gin.H{
			"status":             status.Status,
			"last_message":       status.LastMessage,
			"messages_processed": status.MessagesProcessed,
			"errors":             status.Errors,
			"start_time":         status.StartTime,
			"uptime":             time.Since(status.StartTime).String(),
		})
	})

	// Detailed health endpoint
	router.GET("/health/detailed", func(c *gin.Context) {
		status := kec.GetHealthStatus()

		response := gin.H{
			"consumer": gin.H{
				"status":             status.Status,
				"is_running":         kec.IsRunning(),
				"last_message":       status.LastMessage,
				"messages_processed": status.MessagesProcessed,
				"errors":             status.Errors,
				"start_time":         status.StartTime,
				"uptime":             time.Since(status.StartTime).String(),
			},
			"configuration": gin.H{
				"topic":             kec.config.Topic,
				"group_id":          kec.config.GroupID,
				"brokers":           kec.config.Brokers,
				"auto_offset_reset": kec.config.AutoOffsetReset,
				"max_bytes":         kec.config.MaxBytes,
				"commit_interval":   kec.config.CommitInterval.String(),
				"read_timeout":      kec.config.ReadBatchTimeout.String(),
			},
		}

		c.JSON(http.StatusOK, response)
	})

	// Metrics endpoint (for monitoring systems)
	router.GET("/metrics", func(c *gin.Context) {
		status := kec.GetHealthStatus()

		// Format suitable for Prometheus or similar monitoring systems
		metrics := []string{
			"# HELP kafka_consumer_status Consumer status (1=running, 0=stopped)",
			"# TYPE kafka_consumer_status gauge",
			"kafka_consumer_status " + strconv.Itoa(func() int {
				if status.Status == "running" {
					return 1
				}
				return 0
			}()),
			"",
			"# HELP kafka_consumer_messages_processed_total Total messages processed",
			"# TYPE kafka_consumer_messages_processed_total counter",
			"kafka_consumer_messages_processed_total " + strconv.FormatInt(status.MessagesProcessed, 10),
			"",
			"# HELP kafka_consumer_errors_total Total errors encountered",
			"# TYPE kafka_consumer_errors_total counter",
			"kafka_consumer_errors_total " + strconv.FormatInt(status.Errors, 10),
			"",
			"# HELP kafka_consumer_uptime_seconds Consumer uptime in seconds",
			"# TYPE kafka_consumer_uptime_seconds gauge",
			"kafka_consumer_uptime_seconds " + strconv.FormatFloat(time.Since(status.StartTime).Seconds(), 'f', 2, 64),
		}

		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.String(http.StatusOK, strings.Join(metrics, "\n"))
	})

	// Readiness probe endpoint
	router.GET("/ready", func(c *gin.Context) {
		if kec.IsRunning() {
			c.JSON(http.StatusOK, gin.H{"status": "ready"})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
		}
	})

	// Liveness probe endpoint
	router.GET("/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "alive"})
	})

	// Start server
	addr := ":" + strconv.Itoa(kec.config.HealthCheckPort)
	kec.logger.Info("Starting health check server", map[string]interface{}{
		"port": kec.config.HealthCheckPort,
		"endpoints": []string{
			"GET /health",
			"GET /health/detailed",
			"GET /metrics",
			"GET /ready",
			"GET /live",
		},
	})

	if err := router.Run(addr); err != nil {
		kec.logger.Error("Health server error: ", err)
	}
}
