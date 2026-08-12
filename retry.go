package main

import (
	"context"
	"errors"
	"time"
)

func retryWithDelays(ctx context.Context, delays []time.Duration, operation func(attempt int) error) error {
	if len(delays) == 0 {
		return errors.New("retry schedule is empty")
	}
	var lastErr error
	for index, delay := range delays {
		if delay > 0 {
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				return ctx.Err()
			case <-timer.C:
			}
		}
		if err := operation(index + 1); err != nil {
			lastErr = err
			if ctx.Err() != nil {
				return ctx.Err()
			}
			continue
		}
		return nil
	}
	return lastErr
}
