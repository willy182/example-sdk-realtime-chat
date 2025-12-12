// Package chat: SDK client for realtime messaging
package chat

import (
    "context"
    "errors"
    "fmt"
    "sync"
    "sync/atomic"
    "time"
)

type Config struct {
    APIKey     string
    Endpoint   string
    DialTimeout time.Duration

    // Optional behavior
    Retry RetryStrategy
}

type Message struct {
    Channel string
    From    string
    Body    string
    Time    time.Time
}

// Transport defines how the client talks to the realtime backend.
type Transport interface {
    Connect(ctx context.Context) error
    Send(ctx context.Context, msg Message) error
    Close() error
    Subscribe(chan<- Message) error
}

type Client struct {
    cfg       Config
    transport Transport

    mu        sync.RWMutex
    listeners map[int64]chan Message
    nextID    int64

    recvCh    chan Message
    wg        sync.WaitGroup

    closed    uint32 // atomic flag
    closeOnce sync.Once
}

// NewClient creates and initializes the SDK client.
func NewClient(cfg Config, t Transport) (*Client, error) {
    if cfg.APIKey == "" {
        return nil, errors.New("api key required")
    }
    if cfg.Endpoint == "" {
        return nil, errors.New("endpoint required")
    }
    if cfg.DialTimeout == 0 {
        cfg.DialTimeout = 5 * time.Second
    }

    // default retry strategy -> no retry
    if cfg.Retry == nil {
        cfg.Retry = &NoRetry{}
    }

    c := &Client{
        cfg:       cfg,
        transport: t,
        recvCh:    make(chan Message, 128),
        listeners: make(map[int64]chan Message),
    }

    c.wg.Add(1)
    go c.receiver()

    return c, nil
}

func (c *Client) isClosed() bool {
    return atomic.LoadUint32(&c.closed) == 1
}

func (c *Client) Connect(ctx context.Context) error {
    if c.isClosed() {
        return ErrClientClosed
    }

    ctx, cancel := context.WithTimeout(ctx, c.cfg.DialTimeout)
    defer cancel()

    if err := c.transport.Connect(ctx); err != nil {
        return fmt.Errorf("connect: %w", err)
    }

    if err := c.transport.Subscribe(c.recvCh); err != nil {
        return fmt.Errorf("subscribe: %w", err)
    }

    return nil
}

func (c *Client) SendMessage(ctx context.Context, channel, body string) error {
    if c.isClosed() {
        return ErrClientClosed
    }

    msg := Message{Channel: channel, From: "sdk-client", Body: body, Time: time.Now().UTC()}

    // Use RetryStrategy to perform send
    return c.cfg.Retry.Do(ctx, func() error {
        return c.transport.Send(ctx, msg)
    })
}

// Subscribe returns an id and a channel that receives incoming messages.
// Caller must call Unsubscribe(id) to stop and allow resources cleanup.
func (c *Client) Subscribe() (int64, <-chan Message) {
    ch := make(chan Message, 64)
    c.mu.Lock()
    defer c.mu.Unlock()
    if c.isClosed() {
        close(ch)
        return -1, ch
    }
    id := c.nextID
    c.nextID++
    c.listeners[id] = ch
    return id, ch
}

func (c *Client) Unsubscribe(id int64) {
    c.mu.Lock()
    ch, ok := c.listeners[id]
    if ok {
        delete(c.listeners, id)
        close(ch)
    }
    c.mu.Unlock()
}

func (c *Client) receiver() {
    defer c.wg.Done()
    for msg := range c.recvCh {
        c.mu.RLock()
        listeners := make([]chan Message, 0, len(c.listeners))
        for _, l := range c.listeners {
            listeners = append(listeners, l)
        }
        c.mu.RUnlock()

        for _, l := range listeners {
            select {
            case l <- msg:
            default:
                // drop to avoid blocking slow consumers
            }
        }
    }
}

func (c *Client) Close() error {
    c.closeOnce.Do(func() {
        atomic.StoreUint32(&c.closed, 1)
        // Close transport first to stop producing messages
        _ = c.transport.Close()

        // close recvCh to stop receiver
        close(c.recvCh)
        c.wg.Wait()

        // close listener channels
        c.mu.Lock()
        for id, l := range c.listeners {
            delete(c.listeners, id)
            close(l)
        }
        c.mu.Unlock()
    })
    return nil
}
