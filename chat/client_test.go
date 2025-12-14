package chat

import (
	"context"
	"testing"
	"time"
)

func TestSendReceive(t *testing.T) {
	cfg := Config{APIKey: "k", Endpoint: "mock://"}
	tpt := NewMockTransport(cfg)
	c, _ := NewClient(cfg, tpt)

	c.Connect(context.Background())

	id, sub := c.Subscribe()
	defer c.Unsubscribe(id)

	c.SendMessage(context.Background(), "general", "hi")

	select {
	case m := <-sub:
		if m.Body != "hi" {
			t.Fatal("unexpected message")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}
