package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetryWithDelaysSucceedsOnThirdAttempt(t *testing.T) {
	attempts := 0
	err := retryWithDelays(context.Background(), []time.Duration{0, 0, 0}, func(attempt int) error {
		attempts++
		if attempt < 3 {
			return errors.New("temporary failure")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("retryWithDelays returned an error: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("attempt count = %d, want 3", attempts)
	}
}

func TestRetryWithDelaysStopsWhenCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0
	err := retryWithDelays(ctx, []time.Duration{0, time.Hour}, func(attempt int) error {
		attempts++
		cancel()
		return errors.New("failure")
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("retryWithDelays error = %v, want context.Canceled", err)
	}
	if attempts != 1 {
		t.Fatalf("attempt count = %d, want 1", attempts)
	}
}
