package chat

import (
    "context"
    "time"
)

// RetryStrategy abstracts retry/backoff behavior.
type RetryStrategy interface {
    Do(ctx context.Context, op func() error) error
}

// NoRetry performs the operation once, returning its error.
type NoRetry struct{}

func (r *NoRetry) Do(ctx context.Context, op func() error) error {
    return op()
}

// SimpleExponentialBackoff retries a few times with backoff.
type SimpleExponentialBackoff struct {
    MaxRetries int
    BaseDelay  time.Duration
}

func (s *SimpleExponentialBackoff) Do(ctx context.Context, op func() error) error {
    if s.MaxRetries <= 0 {
        s.MaxRetries = 3
    }
    if s.BaseDelay <= 0 {
        s.BaseDelay = 50 * time.Millisecond
    }
    var err error
    delay := s.BaseDelay
    for i := 0; i <= s.MaxRetries; i++ {
        err = op()
        if err == nil {
            return nil
        }
        // use timer to be cancelable
        t := time.NewTimer(delay)
        select {
        case <-t.C:
            // increase delay
            delay = delay * 2
            t.Stop()
            continue
        case <-ctx.Done():
            t.Stop()
            return ctx.Err()
        }
    }
    return err
}
