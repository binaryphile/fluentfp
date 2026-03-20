package call_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/binaryphile/fluentfp/call"
)

func TestWithSingle(t *testing.T) {
	// double wraps fn to double the result.
	double := func(fn call.Func[int, int]) call.Func[int, int] {
		return func(ctx context.Context, n int) (int, error) {
			r, err := fn(ctx, n)
			return r * 2, err
		}
	}

	fn := call.Func[int, int](func(_ context.Context, n int) (int, error) {
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

	makeRecorder := func(name string) call.Decorator[int, int] {
		return func(fn call.Func[int, int]) call.Func[int, int] {
			return func(ctx context.Context, n int) (int, error) {
				order = append(order, name)
				return fn(ctx, n)
			}
		}
	}

	fn := call.Func[int, int](func(_ context.Context, n int) (int, error) {
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

	counterA := func(fn call.Func[int, int]) call.Func[int, int] {
		return func(ctx context.Context, n int) (int, error) {
			countA.Add(1)
			return fn(ctx, n)
		}
	}
	counterB := func(fn call.Func[int, int]) call.Func[int, int] {
		return func(ctx context.Context, n int) (int, error) {
			countB.Add(1)
			return fn(ctx, n)
		}
	}

	base := call.Func[int, int](func(_ context.Context, n int) (int, error) {
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
	fn := call.Func[int, int](func(_ context.Context, n int) (int, error) {
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

	fn := call.Func[int, int](func(_ context.Context, n int) (int, error) {
		return n, nil
	})
	fn.With(nil)
}

func TestCircuitBreakerDecorator(t *testing.T) {
	b := call.NewBreaker(call.BreakerConfig{
		ResetTimeout: time.Second,
		ReadyToTrip:  call.ConsecutiveFailures(1),
	})

	fn := call.Func[int, int](func(_ context.Context, n int) (int, error) {
		return 0, errors.New("fail")
	})

	wrapped := fn.With(call.CircuitBreaker[int, int](b))

	// First call fails normally.
	_, err := wrapped(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}

	// Second call should be rejected by breaker.
	_, err = wrapped(context.Background(), 1)
	if !errors.Is(err, call.ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestRetrierDecorator(t *testing.T) {
	var attempts int

	fn := call.Func[int, int](func(_ context.Context, n int) (int, error) {
		attempts++
		if attempts < 3 {
			return 0, errors.New("not yet")
		}
		return n * 2, nil
	})

	wrapped := fn.With(call.Retrier[int, int](3, call.ConstantBackoff(0), nil))

	got, err := wrapped(context.Background(), 5)
	if err != nil || got != 10 {
		t.Errorf("Retrier: got (%d, %v), want (10, nil)", got, err)
	}
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}

func TestErrMapperDecorator(t *testing.T) {
	sentinel := errors.New("mapped")
	mapper := func(err error) error { return sentinel }

	fn := call.Func[int, int](func(_ context.Context, n int) (int, error) {
		return 0, errors.New("original")
	})

	wrapped := fn.With(call.ErrMapper[int, int](mapper))

	_, err := wrapped(context.Background(), 1)
	if err != sentinel {
		t.Errorf("ErrMapper: got %v, want %v", err, sentinel)
	}
}

func TestOnErrorDecorator(t *testing.T) {
	var captured error
	handler := func(err error) { captured = err }

	original := errors.New("fail")
	fn := call.Func[int, int](func(_ context.Context, n int) (int, error) {
		return 0, original
	})

	wrapped := fn.With(call.OnError[int, int](handler))

	wrapped(context.Background(), 1)
	if captured != original {
		t.Errorf("OnError: captured %v, want %v", captured, original)
	}
}
