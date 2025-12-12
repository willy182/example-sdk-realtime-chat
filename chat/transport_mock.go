package chat

import (
    "context"
    "errors"
    "sync"
    "time"
)

// MockTransport simulates backend; it can be configured to fail initial sends for retry tests.
type MockTransport struct {
    endpoint string
    apiKey   string

    in  chan Message
    out chan Message

    mu      sync.RWMutex
    running bool
    wg      sync.WaitGroup

    // failFirstSend count: how many first Send calls should return error
    failFirstSends int
    sendCounter    int
    failErr        error
}

func NewMockTransport(cfg Config) *MockTransport {
    return &MockTransport{
        endpoint: cfg.Endpoint,
        apiKey:   cfg.APIKey,
        in:       make(chan Message, 512),
        out:      make(chan Message, 512),
        failErr:  errors.New("simulated send failure"),
    }
}

// WithFailFirstSends configures mock to fail first n sends.
func (m *MockTransport) WithFailFirstSends(n int) *MockTransport {
    m.failFirstSends = n
    return m
}

func (m *MockTransport) Connect(ctx context.Context) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    if m.running {
        return nil
    }
    m.running = true
    m.wg.Add(1)
    go m.loop()
    return nil
}

func (m *MockTransport) loop() {
    defer m.wg.Done()
    for msg := range m.in {
        // simulate processing delay
        time.Sleep(5 * time.Millisecond)
        msg.From = "remote-user"
        // non-blocking write to out
        select {
        case m.out <- msg:
        default:
            // drop if full
        }
    }
}

func (m *MockTransport) Send(ctx context.Context, msg Message) error {
    m.mu.Lock()
    running := m.running
    if !running {
        m.mu.Unlock()
        return ErrTransportNotConnected
    }
    // check fail-first behavior
    if m.sendCounter < m.failFirstSends {
        m.sendCounter++
        m.mu.Unlock()
        return m.failErr
    }
    m.mu.Unlock()

    select {
    case m.in <- msg:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

func (m *MockTransport) Subscribe(dst chan<- Message) error {
    m.mu.RLock()
    if !m.running {
        m.mu.RUnlock()
        return ErrTransportNotConnected
    }
    m.mu.RUnlock()

    m.wg.Add(1)
    go func() {
        defer m.wg.Done()
        for msg := range m.out {
            select {
            case dst <- msg:
            default:
                // drop
            }
        }
    }()
    return nil
}

func (m *MockTransport) Close() error {
    m.mu.Lock()
    if !m.running {
        m.mu.Unlock()
        return nil
    }
    m.running = false
    close(m.in)
    m.mu.Unlock()

    m.wg.Wait()
    // close out only after loop finished
    close(m.out)
    return nil
}
