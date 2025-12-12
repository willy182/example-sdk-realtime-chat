package main

import (
	"context"
	"fmt"
	"time"

	"github.com/willy182/example-sdk-realtime-chat/chat"
)

func main() {
	// Build config with builder
	cfg, err := chat.NewConfigBuilder().
		WithAPIKey("demo-key").
		WithEndpoint("mock://").
		WithDialTimeout(2 * time.Second).
		Build()
	if err != nil {
		panic(err)
	}

	// create transport via factory
	tpt, err := chat.NewTransport(chat.TransportMock, cfg)
	if err != nil {
		panic(err)
	}

	// optional retry
	cfg.Retry = &chat.SimpleExponentialBackoff{MaxRetries: 2, BaseDelay: 20 * time.Millisecond}

	client, err := chat.NewClient(cfg, tpt)
	if err != nil {
		panic(err)
	}

	// Ensure client is closed on exit
	defer func() {
		_ = client.Close()
	}()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		panic(err)
	}

	// subscribe and get id+channel
	id, sub := client.Subscribe()

	// send message
	if err := client.SendMessage(ctx, "general", "Hello from example"); err != nil {
		client.Unsubscribe(id)
		panic(err)
	}

	// Read a single message (or timeout) synchronously — no extra goroutine
	select {
	case m, ok := <-sub:
		if !ok {
			fmt.Println("subscription channel closed")
		} else {
			fmt.Printf("received: %+v\n", m)
		}
	case <-time.After(1 * time.Second):
		fmt.Println("timeout waiting for message")
	}

	// Unsubscribe before closing the client to avoid receiver goroutines stuck
	client.Unsubscribe(id)

	// close will be done by defer
	fmt.Println("shutdown complete")
}
