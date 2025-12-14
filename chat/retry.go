package chat

import (
	"context"
	"time"
)

// RetryStrategy defines how retry logic is applied.
type RetryStrategy interface {
	Do(ctx context.Context, op func() error) error
}

// NoRetry executes the operation once without retry.
type NoRetry struct{}

func (r *NoRetry) Do(ctx context.Context, op func() error) error {
	return op()
}

// SimpleExponentialBackoff retries an operation with exponential backoff.
// It is deterministic, context-aware, and safe for unit testing.
type SimpleExponentialBackoff struct {
	MaxRetries int
	BaseDelay  time.Duration
}

// Do executes the operation with exponential backoff.
func (s *SimpleExponentialBackoff) Do(ctx context.Context, op func() error) error {
	if s.MaxRetries <= 0 {
		s.MaxRetries = 3
	}
	if s.BaseDelay <= 0 {
		s.BaseDelay = 50 * time.Millisecond
	}

	var err error
	delay := s.BaseDelay

	for attempt := 0; attempt <= s.MaxRetries; attempt++ {
		err = op()
		if err == nil {
			return nil
		}

		// wait before next retry, but respect context cancellation
		timer := time.NewTimer(delay)
		select {
		case <-timer.C:
			delay *= 2
			timer.Stop()
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		}
	}

	return err
}
