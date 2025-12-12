package chat

import (
    "context"
    "testing"
    "time"
)

func TestSendReceive(t *testing.T) {
    cfg := Config{APIKey: "key", Endpoint: "mock://", DialTimeout: 2 * time.Second}
    tpt := NewMockTransport(cfg)
    c, err := NewClient(cfg, tpt)
    if err != nil {
        t.Fatalf("NewClient: %v", err)
    }
    defer c.Close()

    ctx := context.Background()
    if err := c.Connect(ctx); err != nil {
        t.Fatalf("Connect: %v", err)
    }

    id, sub := c.Subscribe()
    defer c.Unsubscribe(id)

    if err := c.SendMessage(ctx, "general", "Hello, world!"); err != nil {
        t.Fatalf("SendMessage: %v", err)
    }

    select {
    case msg := <-sub:
        if msg.Body != "Hello, world!" {
            t.Fatalf("unexpected body: %s", msg.Body)
        }
    case <-time.After(1 * time.Second):
        t.Fatal("timeout waiting for message")
    }
}

func TestCloseIdempotent(t *testing.T) {
    cfg := Config{APIKey: "key", Endpoint: "mock://"}
    tpt := NewMockTransport(cfg)
    c, err := NewClient(cfg, tpt)
    if err != nil {
        t.Fatalf("NewClient: %v", err)
    }

    if err := c.Close(); err != nil {
        t.Fatalf("Close: %v", err)
    }
    if err := c.Close(); err != nil {
        t.Fatalf("Close2: %v", err)
    }
}

func TestSendWhenClosed(t *testing.T) {
    cfg := Config{APIKey: "key", Endpoint: "mock://"}
    tpt := NewMockTransport(cfg)
    c, err := NewClient(cfg, tpt)
    if err != nil {
        t.Fatalf("NewClient: %v", err)
    }

    c.Close()
    ctx := context.Background()
    if err := c.SendMessage(ctx, "general", "hi"); err == nil {
        t.Fatalf("expected error sending on closed client")
    }
}
