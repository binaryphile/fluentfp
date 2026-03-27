package wrap_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/binaryphile/fluentfp/wrap"
)

func TestWithEmptyFeatures(t *testing.T) {
	fn := wrap.Func(func(_ context.Context, n int) (int, error) {
		return n * 3, nil
	})

	got, err := fn.With(wrap.Features{})(context.Background(), 5)
	if err != nil || got != 15 {
		t.Errorf("With(Features{})(5) = (%d, %v), want (15, nil)", got, err)
	}
}

func TestFn_WithBreaker(t *testing.T) {
	b := wrap.NewBreaker(wrap.BreakerConfig{
		ResetTimeout: time.Second,
		ReadyToTrip:  wrap.ConsecutiveFailures(1),
	})

	fn := wrap.Func(func(_ context.Context, n int) (int, error) {
		return 0, errors.New("fail")
	})

	wrapped := fn.With(wrap.Features{Breaker: b})

	// First call fails normally.
	_, err := wrapped(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}

	// Second call should be rejected by breaker.
	_, err = wrapped(context.Background(), 1)
	if !errors.Is(err, wrap.ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestWithRetry(t *testing.T) {
	var attempts int

	// constBackoff always waits 0 for testing.
	constBackoff := wrap.Backoff(func(int) time.Duration { return 0 })

	fn := wrap.Func(func(_ context.Context, n int) (int, error) {
		attempts++
		if attempts < 3 {
			return 0, errors.New("not yet")
		}
		return n * 2, nil
	})

	wrapped := fn.With(wrap.Features{Retry: wrap.Retry(3, constBackoff, nil)})

	got, err := wrapped(context.Background(), 5)
	if err != nil || got != 10 {
		t.Errorf("WithRetry: got (%d, %v), want (10, nil)", got, err)
	}
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}

func TestWithMapError(t *testing.T) {
	sentinel := errors.New("mapped")
	mapper := func(err error) error { return sentinel }

	fn := wrap.Func(func(_ context.Context, n int) (int, error) {
		return 0, errors.New("original")
	})

	wrapped := fn.With(wrap.Features{MapError: mapper})

	_, err := wrapped(context.Background(), 1)
	if err != sentinel {
		t.Errorf("WithMapError: got %v, want %v", err, sentinel)
	}
}

func TestWithOnError(t *testing.T) {
	var captured error
	handler := func(err error) { captured = err }

	original := errors.New("fail")
	fn := wrap.Func(func(_ context.Context, n int) (int, error) {
		return 0, original
	})

	wrapped := fn.With(wrap.Features{OnError: handler})

	wrapped(context.Background(), 1)
	if captured != original {
		t.Errorf("WithOnError: captured %v, want %v", captured, original)
	}
}
