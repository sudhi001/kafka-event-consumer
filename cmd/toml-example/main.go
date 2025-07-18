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
	// 🚀 TOML Configuration Example with Multiple Topics 🚀
	// Update config.toml to specify multiple topics in the [kafka] section
	consumer, err := consumer.NewConsumerFromConfig("config.toml", handleMessage)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}

	log.Fatal(consumer.Run())
}
