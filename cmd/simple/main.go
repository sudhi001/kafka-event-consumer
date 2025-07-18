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
	fmt.Printf("Processing message: ID=%s, Data=%s\n", msg.ID, msg.Data)
	return nil
}

func main() {
	// 🚀 JUST 2 LINES OF CODE! 🚀
	consumer := consumer.NewConsumer(handleMessage)
	log.Fatal(consumer.Run())
}
