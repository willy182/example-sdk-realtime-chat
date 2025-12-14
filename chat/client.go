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
	APIKey      string
	Endpoint    string
	DialTimeout time.Duration
	Retry       RetryStrategy
}

type Message struct {
	Channel string
	From    string
	Body    string
	Time    time.Time
}

type Transport interface {
	Connect(ctx context.Context) error
	Send(ctx context.Context, msg Message) error
	Subscribe(chan<- Message) error
	Close() error
}

type Client struct {
	cfg       Config
	transport Transport

	recvCh chan Message
	wg     sync.WaitGroup

	mu        sync.RWMutex
	listeners map[int64]chan Message
	nextID    int64

	closed    uint32
	closeOnce sync.Once
}

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

	msg := Message{
		Channel: channel,
		From:    "sdk-client",
		Body:    body,
		Time:    time.Now().UTC(),
	}

	return c.cfg.Retry.Do(ctx, func() error {
		return c.transport.Send(ctx, msg)
	})
}

func (c *Client) Subscribe() (int64, <-chan Message) {
	ch := make(chan Message, 64)

	c.mu.Lock()
	defer c.mu.Unlock()

	id := c.nextID
	c.nextID++
	c.listeners[id] = ch

	return id, ch
}

func (c *Client) Unsubscribe(id int64) {
	c.mu.Lock()
	if ch, ok := c.listeners[id]; ok {
		delete(c.listeners, id)
		close(ch)
	}
	c.mu.Unlock()
}

func (c *Client) receiver() {
	defer c.wg.Done()

	for msg := range c.recvCh {
		c.mu.RLock()
		for _, ch := range c.listeners {
			select {
			case ch <- msg:
			default:
			}
		}
		c.mu.RUnlock()
	}
}

func (c *Client) Close() error {
	c.closeOnce.Do(func() {
		atomic.StoreUint32(&c.closed, 1)
		_ = c.transport.Close()
		close(c.recvCh)
		c.wg.Wait()

		c.mu.Lock()
		for id, ch := range c.listeners {
			delete(c.listeners, id)
			close(ch)
		}
		c.mu.Unlock()
	})
	return nil
}
