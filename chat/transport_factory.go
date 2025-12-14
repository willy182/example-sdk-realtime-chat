package chat

type TransportType string

const TransportMock TransportType = "mock"

func NewTransport(t TransportType, cfg Config) (Transport, error) {
	return NewMockTransport(cfg), nil
}
