package hof_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/binaryphile/fluentfp/hof"
)

func TestOnErrBasicSuccess(t *testing.T) {
	called := false

	// onErr records whether it was called.
	onErr := func() { called = true }
	// doubleIt doubles the input.
	doubleIt := func(_ context.Context, n int) (int, error) { return n * 2, nil }

	wrapped := hof.OnErr(doubleIt, onErr)
	got, err := wrapped(context.Background(), 5)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 10 {
		t.Fatalf("got %d, want 10", got)
	}
	if called {
		t.Fatal("onErr should not be called on success")
	}
}

func TestOnErrCallsOnError(t *testing.T) {
	called := false

	// onErr records whether it was called.
	onErr := func() { called = true }
	errBoom := errors.New("boom")
	// failingFn always returns an error.
	failingFn := func(_ context.Context, _ int) (int, error) { return 0, errBoom }

	wrapped := hof.OnErr(failingFn, onErr)
	_, err := wrapped(context.Background(), 5)

	if !errors.Is(err, errBoom) {
		t.Fatalf("got error %v, want %v", err, errBoom)
	}
	if !called {
		t.Fatal("onErr should be called on error")
	}
}

func TestOnErrWithContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// onErr cancels the context on first error.
	onErr := func() { cancel() }
	errBoom := errors.New("boom")
	// failingFn always returns an error.
	failingFn := func(_ context.Context, _ int) (int, error) { return 0, errBoom }

	wrapped := hof.OnErr(failingFn, onErr)
	_, err := wrapped(ctx, 5)

	if !errors.Is(err, errBoom) {
		t.Fatalf("got error %v, want %v", err, errBoom)
	}
	if ctx.Err() == nil {
		t.Fatal("context should be cancelled after error")
	}
}

func TestOnErrComposesWithThrottle(t *testing.T) {
	var errCount atomic.Int32

	// onErr increments the error counter (concurrency-safe).
	onErr := func() { errCount.Add(1) }
	// doubleOrFail doubles positive inputs, errors on negative.
	doubleOrFail := func(_ context.Context, n int) (int, error) {
		if n < 0 {
			return 0, errors.New("negative")
		}

		return n * 2, nil
	}

	// Compose: Throttle wrapping OnErr.
	throttled := hof.Throttle(2, hof.OnErr(doubleOrFail, onErr))

	// Success path.
	got, err := throttled(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 10 {
		t.Fatalf("got %d, want 10", got)
	}
	if errCount.Load() != 0 {
		t.Fatalf("errCount = %d, want 0", errCount.Load())
	}

	// Error path.
	_, err = throttled(context.Background(), -1)
	if err == nil {
		t.Fatal("expected error for negative input")
	}
	if errCount.Load() != 1 {
		t.Fatalf("errCount = %d, want 1", errCount.Load())
	}
}

func TestOnErrValidationPanics(t *testing.T) {
	// dummyFn is a placeholder function.
	dummyFn := func(_ context.Context, _ int) (int, error) { return 0, nil }
	// dummyOnErr is a placeholder side-effect.
	dummyOnErr := func() {}

	tests := []struct {
		name string
		fn   func()
	}{
		{
			name: "nil_fn",
			fn:   func() { hof.OnErr[int, int](nil, dummyOnErr) },
		},
		{
			name: "nil_onErr",
			fn:   func() { hof.OnErr(dummyFn, nil) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected panic")
				}
			}()
			tt.fn()
		})
	}
}
