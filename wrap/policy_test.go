package wrap_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/binaryphile/fluentfp/wrap"
)

func TestWithSingle(t *testing.T) {
	// double wraps fn to double the result.
	double := func(fn wrap.Fn[int, int]) wrap.Fn[int, int] {
		return func(ctx context.Context, n int) (int, error) {
			r, err := fn(ctx, n)
			return r * 2, err
		}
	}

	fn := wrap.Func(func(_ context.Context, n int) (int, error) {
		return n, nil
	})

	got, err := fn.With(double)(context.Background(), 5)
	if err != nil || got != 10 {
		t.Errorf("With(double)(5) = (%d, %v), want (10, nil)", got, err)
	}
}

func TestWithOrder(t *testing.T) {
	// Record the order decorators run on the way in.
	var order []string

	makeRecorder := func(name string) wrap.Decorator[int, int] {
		return func(fn wrap.Fn[int, int]) wrap.Fn[int, int] {
			return func(ctx context.Context, n int) (int, error) {
				order = append(order, name)
				return fn(ctx, n)
			}
		}
	}

	fn := wrap.Func(func(_ context.Context, n int) (int, error) {
		order = append(order, "base")
		return n, nil
	})

	// With(A, B, C): A is innermost, C is outermost.
	// Execution order on the way in: C, B, A, base.
	wrapped := fn.With(
		makeRecorder("A"),
		makeRecorder("B"),
		makeRecorder("C"),
	)

	wrapped(context.Background(), 1)

	want := []string{"C", "B", "A", "base"}
	if fmt.Sprint(order) != fmt.Sprint(want) {
		t.Errorf("order = %v, want %v", order, want)
	}
}

func TestWithBranching(t *testing.T) {
	var countA, countB atomic.Int64

	counterA := func(fn wrap.Fn[int, int]) wrap.Fn[int, int] {
		return func(ctx context.Context, n int) (int, error) {
			countA.Add(1)
			return fn(ctx, n)
		}
	}
	counterB := func(fn wrap.Fn[int, int]) wrap.Fn[int, int] {
		return func(ctx context.Context, n int) (int, error) {
			countB.Add(1)
			return fn(ctx, n)
		}
	}

	base := wrap.Func(func(_ context.Context, n int) (int, error) {
		return n, nil
	})

	branchA := base.With(counterA)
	branchB := base.With(counterB)

	branchA(context.Background(), 1)
	branchB(context.Background(), 1)
	branchA(context.Background(), 1)

	if countA.Load() != 2 {
		t.Errorf("countA = %d, want 2", countA.Load())
	}
	if countB.Load() != 1 {
		t.Errorf("countB = %d, want 1", countB.Load())
	}
}

func TestWithEmpty(t *testing.T) {
	fn := wrap.Func(func(_ context.Context, n int) (int, error) {
		return n * 3, nil
	})

	got, err := fn.With()(context.Background(), 5)
	if err != nil || got != 15 {
		t.Errorf("With()(5) = (%d, %v), want (15, nil)", got, err)
	}
}

func TestWithNilPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for nil decorator")
		}
	}()

	fn := wrap.Func(func(_ context.Context, n int) (int, error) {
		return n, nil
	})
	fn.With(nil)
}

func TestFn_WithBreaker(t *testing.T) {
	b := wrap.NewBreaker(wrap.BreakerConfig{
		ResetTimeout: time.Second,
		ReadyToTrip:  wrap.ConsecutiveFailures(1),
	})

	fn := wrap.Func(func(_ context.Context, n int) (int, error) {
		return 0, errors.New("fail")
	})

	wrapped := fn.WithBreaker(b)

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

	wrapped := fn.WithRetry(3, constBackoff, nil)

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

	wrapped := fn.WithMapError(mapper)

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

	wrapped := fn.WithOnError(handler)

	wrapped(context.Background(), 1)
	if captured != original {
		t.Errorf("WithOnError: captured %v, want %v", captured, original)
	}
}
