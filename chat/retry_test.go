package chat

import (
    "context"
    "testing"
    "time"
)

func TestRetrySucceedsAfterFailures(t *testing.T) {
    cfg := Config{APIKey: "k", Endpoint: "mock://", DialTimeout: 2 * time.Second}
    // configure mock to fail first 2 sends
    tpt := NewMockTransport(cfg).WithFailFirstSends(2)
    // set retry strategy to allow retries
    cfg.Retry = &SimpleExponentialBackoff{MaxRetries: 3, BaseDelay: 10 * time.Millisecond}

    c, err := NewClient(cfg, tpt)
    if err != nil {
        t.Fatalf("NewClient: %v", err)
    }
    defer c.Close()

    if err := c.Connect(context.Background()); err != nil {
        t.Fatalf("Connect: %v", err)
    }

    id, ch := c.Subscribe()
    defer c.Unsubscribe(id)

    if err := c.SendMessage(context.Background(), "general", "hello"); err != nil {
        t.Fatalf("SendMessage with retry failed: %v", err)
    }

    select {
    case msg := <-ch:
        if msg.Body != "hello" {
            t.Fatalf("unexpected body: %s", msg.Body)
        }
    case <-time.After(1 * time.Second):
        t.Fatal("timeout waiting for message after retry")
    }
}

func TestSendContextCancel(t *testing.T) {
    cfg := Config{APIKey: "k", Endpoint: "mock://", DialTimeout: 2 * time.Second}
    tpt := NewMockTransport(cfg).WithFailFirstSends(5)
    // small retries but will be canceled by context
    cfg.Retry = &SimpleExponentialBackoff{MaxRetries: 5, BaseDelay: 50 * time.Millisecond}

    c, err := NewClient(cfg, tpt)
    if err != nil {
        t.Fatalf("NewClient: %v", err)
    }
    defer c.Close()

    if err := c.Connect(context.Background()); err != nil {
        t.Fatalf("Connect: %v", err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
    defer cancel()

    err = c.SendMessage(ctx, "general", "x")
    if err == nil {
        t.Fatalf("expected error due to context cancel, got nil")
    }
}
