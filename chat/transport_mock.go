package chat

import (
	"context"
	"errors"
	"sync"
)

type MockTransport struct {
	in  chan Message
	out chan Message

	mu      sync.Mutex
	running bool
	wg      sync.WaitGroup
}

func NewMockTransport(cfg Config) *MockTransport {
	return &MockTransport{
		in:  make(chan Message, 256),
		out: make(chan Message, 256),
	}
}

func (m *MockTransport) Connect(ctx context.Context) error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return nil
	}
	m.running = true
	m.mu.Unlock()

	started := make(chan struct{})

	m.wg.Add(1)
	go func() {
		close(started)
		defer m.wg.Done()
		for msg := range m.in {
			msg.From = "remote-user"
			select {
			case m.out <- msg:
			default:
			}
		}
		close(m.out)
	}()

	<-started
	return nil
}

func (m *MockTransport) Send(ctx context.Context, msg Message) error {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return errors.New("not connected")
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
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		for msg := range m.out {
			select {
			case dst <- msg:
			default:
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
	return nil
}
