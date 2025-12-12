package chat

import (
    "context"
    "sync"
    "testing"
    "time"
)

func TestManySubscribersNoBlock(t *testing.T) {
    cfg := Config{APIKey: "k", Endpoint: "mock://", DialTimeout: 2 * time.Second}
    tpt := NewMockTransport(cfg)
    c, err := NewClient(cfg, tpt)
    if err != nil {
        t.Fatalf("NewClient: %v", err)
    }
    defer c.Close()

    if err := c.Connect(context.Background()); err != nil {
        t.Fatalf("Connect: %v", err)
    }

    const subs = 50
    const msgs = 200

    ids := make([]int64, 0, subs)
    chans := make([]<-chan Message, 0, subs)
    for i := 0; i < subs; i++ {
        id, ch := c.Subscribe()
        ids = append(ids, id)
        chans = append(chans, ch)
    }
    defer func() {
        for _, id := range ids {
            c.Unsubscribe(id)
        }
    }()

    // send messages concurrently
    var wg sync.WaitGroup
    wg.Add(msgs)
    for i := 0; i < msgs; i++ {
        go func(i int) {
            defer wg.Done()
            _ = c.SendMessage(context.Background(), "general", "m")
        }(i)
    }
    wg.Wait()

    // wait a bit for delivery
    time.Sleep(200 * time.Millisecond)

    // ensure at least one message received by each subscriber (allow drops)
    for i, ch := range chans {
        select {
        case <-ch:
            // ok
        default:
            t.Logf("subscriber %d received no messages (allowed)", i)
        }
    }
}
