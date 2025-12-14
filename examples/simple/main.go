package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/willy182/example-sdk-realtime-chat/chat"
)

func main() {
	cfg, err := chat.NewConfigBuilder().
		WithAPIKey("demo").
		WithEndpoint("mock://").
		WithDialTimeout(10 * time.Second).
		WithRetry(&chat.SimpleExponentialBackoff{
			MaxRetries: 3,
			BaseDelay:  1 * time.Second,
		}).
		Build()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	tpt, err := chat.NewTransport(chat.TransportMock, cfg)
	if err != nil {
		log.Fatalf("transport error: %v", err)
	}

	client, err := chat.NewClient(cfg, tpt)
	if err != nil {
		log.Fatalf("client error: %v", err)
	}

	if err := client.Connect(context.Background()); err != nil {
		log.Fatalf("connect error: %v", err)
	}

	id, sub := client.Subscribe()

	// HANDLE SendMessage error
	if err := client.SendMessage(context.Background(), "general", "Hello from example"); err != nil {
		log.Fatalf("send message error: %v", err)
	}

	select {
	case m := <-sub:
		fmt.Println("received:", m)
	case <-time.After(1 * time.Second):
		log.Println("timeout waiting for message")
	}

	client.Unsubscribe(id)
	if err := client.Close(); err != nil {
		log.Printf("close error: %v", err)
	}

	log.Println("shutdown complete")
}
