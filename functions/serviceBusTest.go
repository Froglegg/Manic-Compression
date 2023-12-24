package main

import (
	"fmt"
	serviceBus "manic-compression/pkg/service_bus"
)

func main() {
	sb := serviceBus.NewServiceBus()
	queue := "audiotasks"

	fmt.Println("Sending a single message...")
	sb.SendMessage(serviceBus.Msg{Type: "single", Content: "firstMessage"}, queue)

	fmt.Println("\nSending two messages as a batch...")
	messagesBatch := []serviceBus.Msg{
		{Type: "batch", Content: "secondMessage"},
		{Type: "batch", Content: "thirdMessage"},
	}
	sb.SendMessageBatch(messagesBatch, queue)

	// fmt.Println("\nRetrieving messages...")
	// sb.GetMessage(3, queue)
}
