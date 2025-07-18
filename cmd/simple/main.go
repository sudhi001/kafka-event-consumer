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

	// Your business logic here
	fmt.Printf("Processing message from topic '%s': ID=%s, Data=%s\n", topic, msg.ID, msg.Data)
	return nil
}

func main() {
	// 🚀 JUST 2 LINES OF CODE! 🚀
	// Set KAFKA_TOPICS environment variable to specify multiple topics
	// Example: export KAFKA_TOPICS="topic1,topic2,topic3"
	consumer := consumer.NewConsumer(handleMessage)
	log.Fatal(consumer.Run())
}
