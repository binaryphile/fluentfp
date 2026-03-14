package hof

import (
	"context"
	"math"
	"math/rand/v2"
	"time"
)

// Backoff computes the delay before retry number n (0-indexed).
// Called between attempts: backoff(0) is the delay before the first retry.
type Backoff func(n int) time.Duration

// ConstantBackoff returns a Backoff that always waits delay.
func ConstantBackoff(delay time.Duration) Backoff {
	return func(int) time.Duration { return delay }
}

// ExponentialBackoff returns a Backoff with full jitter: random in [0, initial * 2^n).
// Panics if initial <= 0.
func ExponentialBackoff(initial time.Duration) Backoff {
	if initial <= 0 {
		panic("hof.ExponentialBackoff: initial must be > 0")
	}

	return func(n int) time.Duration {
		max := initial << n
		if max <= 0 {
			max = math.MaxInt64
		}

		return rand.N(max)
	}
}

// Retry wraps fn to retry on error up to maxAttempts total times.
// The first call is immediate; backoff(0) is the delay before the first retry.
// Returns the result and error from the last attempt.
//
// shouldRetry controls which errors trigger a retry. When non-nil, only errors
// for which shouldRetry returns true are retried; non-retryable errors are
// returned immediately without backoff. When nil, all errors are retried.
//
// Context cancellation is checked before each attempt and during backoff waits.
// Panics if maxAttempts < 1, backoff is nil, or fn is nil.
func Retry[T, R any](maxAttempts int, backoff Backoff, shouldRetry func(error) bool, fn func(context.Context, T) (R, error)) func(context.Context, T) (R, error) {
	if maxAttempts < 1 {
		panic("hof.Retry: maxAttempts must be > 0")
	}
	if backoff == nil {
		panic("hof.Retry: backoff must not be nil")
	}
	if fn == nil {
		panic("hof.Retry: fn must not be nil")
	}

	return func(ctx context.Context, t T) (R, error) {
		var lastR R
		var lastErr error

		for attempt := 0; attempt < maxAttempts; attempt++ {
			if err := ctx.Err(); err != nil {
				return lastR, err
			}

			lastR, lastErr = fn(ctx, t)
			if lastErr == nil {
				return lastR, nil
			}
			if shouldRetry != nil && !shouldRetry(lastErr) {
				return lastR, lastErr
			}

			if attempt < maxAttempts-1 {
				delay := backoff(attempt)
				timer := time.NewTimer(delay)

				select {
				case <-timer.C:
				case <-ctx.Done():
					timer.Stop()

					return lastR, ctx.Err()
				}
			}
		}

		return lastR, lastErr
	}
}
