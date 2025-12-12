package chat

import (
    "fmt"
)

type TransportType string

const (
    TransportMock TransportType = "mock"
    TransportWS   TransportType = "websocket"
)

// NewTransport creates a transport implementation based on type.
// For now, only Mock is implemented. WebSocket placeholder returns error.
func NewTransport(t TransportType, cfg Config) (Transport, error) {
    switch t {
    case TransportMock:
        return NewMockTransport(cfg), nil
    case TransportWS:
        // Implement real websocket transport later.
        return nil, fmt.Errorf("websocket transport not implemented")
    default:
        return nil, fmt.Errorf("unknown transport type %s", t)
    }
}
