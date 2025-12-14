package chat

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fake operation: fails N times then succeeds
type flakyOperation struct {
	failCount int
	calls     int
}

func (f *flakyOperation) Run() error {
	f.calls++
	if f.calls <= f.failCount {
		return errors.New("temporary failure")
	}
	return nil
}

func TestSimpleExponentialBackoff_Succeeds(t *testing.T) {
	op := &flakyOperation{failCount: 2}

	retry := &SimpleExponentialBackoff{
		MaxRetries: 3,
		BaseDelay:  1 * time.Millisecond,
	}

	ctx := context.Background()
	err := retry.Do(ctx, op.Run)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if op.calls != 3 {
		t.Fatalf("expected 3 calls, got %d", op.calls)
	}
}

func TestSimpleExponentialBackoff_ContextCancelled(t *testing.T) {
	op := &flakyOperation{failCount: 100}

	retry := &SimpleExponentialBackoff{
		MaxRetries: 10,
		BaseDelay:  10 * time.Millisecond,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Millisecond)
	defer cancel()

	err := retry.Do(ctx, op.Run)
	if err == nil {
		t.Fatal("expected error due to context cancellation")
	}
}
