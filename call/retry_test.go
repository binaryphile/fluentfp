package call_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/binaryphile/fluentfp/call"
)

func TestRetry(t *testing.T) {
	t.Run("succeeds on first attempt", func(t *testing.T) {
		calls := 0
		// doubleIt doubles the input.
		doubleIt := func(_ context.Context, n int) (int, error) {
			calls++
			return n * 2, nil
		}

		retried := call.Retry(3, call.ConstantBackoff(0), nil, doubleIt)
		got, err := retried(context.Background(), 5)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 10 {
			t.Fatalf("got %d, want 10", got)
		}
		if calls != 1 {
			t.Fatalf("fn called %d times, want 1", calls)
		}
	})

	t.Run("succeeds after retries", func(t *testing.T) {
		calls := 0
		// failTwiceThenDouble fails twice, then doubles.
		failTwiceThenDouble := func(_ context.Context, n int) (int, error) {
			calls++
			if calls < 3 {
				return 0, fmt.Errorf("attempt %d", calls)
			}
			return n * 2, nil
		}

		retried := call.Retry(3, call.ConstantBackoff(0), nil, failTwiceThenDouble)
		got, err := retried(context.Background(), 5)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 10 {
			t.Fatalf("got %d, want 10", got)
		}
		if calls != 3 {
			t.Fatalf("fn called %d times, want 3", calls)
		}
	})

	t.Run("exhausts all attempts", func(t *testing.T) {
		calls := 0
		// alwaysFail always returns an error.
		alwaysFail := func(_ context.Context, _ int) (int, error) {
			calls++
			return 0, fmt.Errorf("fail %d", calls)
		}

		retried := call.Retry(3, call.ConstantBackoff(0), nil, alwaysFail)
		_, err := retried(context.Background(), 1)

		if err == nil {
			t.Fatal("expected error")
		}
		if err.Error() != "fail 3" {
			t.Fatalf("got error %q, want %q", err.Error(), "fail 3")
		}
		if calls != 3 {
			t.Fatalf("fn called %d times, want 3", calls)
		}
	})

	t.Run("maxAttempts 1 no retries", func(t *testing.T) {
		calls := 0
		// alwaysFail always returns an error.
		alwaysFail := func(_ context.Context, _ int) (int, error) {
			calls++
			return 0, fmt.Errorf("fail")
		}

		retried := call.Retry(1, call.ConstantBackoff(0), nil, alwaysFail)
		_, err := retried(context.Background(), 1)

		if err == nil {
			t.Fatal("expected error")
		}
		if calls != 1 {
			t.Fatalf("fn called %d times, want 1", calls)
		}
	})

	t.Run("context cancelled before first attempt", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		calls := 0
		// doubleIt doubles the input.
		doubleIt := func(_ context.Context, n int) (int, error) {
			calls++
			return n * 2, nil
		}

		retried := call.Retry(3, call.ConstantBackoff(0), nil, doubleIt)
		_, err := retried(ctx, 5)

		if err != context.Canceled {
			t.Fatalf("got error %v, want context.Canceled", err)
		}
		if calls != 0 {
			t.Fatalf("fn called %d times, want 0", calls)
		}
	})

	t.Run("context cancelled during backoff", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		calls := 0
		// failAndCancel fails and cancels context on first call.
		failAndCancel := func(_ context.Context, _ int) (int, error) {
			calls++
			if calls == 1 {
				cancel()
			}
			return 0, fmt.Errorf("fail")
		}

		retried := call.Retry(3, call.ConstantBackoff(10*time.Second), nil, failAndCancel)
		_, err := retried(ctx, 1)

		if err != context.Canceled {
			t.Fatalf("got error %v, want context.Canceled", err)
		}
		if calls != 1 {
			t.Fatalf("fn called %d times, want 1", calls)
		}
	})
}

func TestRetryShouldRetry(t *testing.T) {
	errRetryable := fmt.Errorf("retryable")
	errFatal := fmt.Errorf("fatal")

	t.Run("stops on non-retryable error", func(t *testing.T) {
		calls := 0
		// alwaysFatal always returns a non-retryable error.
		alwaysFatal := func(_ context.Context, _ int) (int, error) {
			calls++
			return 0, errFatal
		}
		// isRetryable returns true only for retryable errors.
		isRetryable := func(err error) bool { return err == errRetryable }

		retried := call.Retry(3, call.ConstantBackoff(0), isRetryable, alwaysFatal)
		_, err := retried(context.Background(), 1)

		if err != errFatal {
			t.Fatalf("got error %v, want %v", err, errFatal)
		}
		if calls != 1 {
			t.Fatalf("fn called %d times, want 1", calls)
		}
	})

	t.Run("retries retryable then stops on fatal", func(t *testing.T) {
		calls := 0
		// retryableThenFatal returns retryable twice, then fatal.
		retryableThenFatal := func(_ context.Context, _ int) (int, error) {
			calls++
			if calls < 3 {
				return 0, errRetryable
			}
			return 0, errFatal
		}
		// isRetryable returns true only for retryable errors.
		isRetryable := func(err error) bool { return err == errRetryable }

		retried := call.Retry(5, call.ConstantBackoff(0), isRetryable, retryableThenFatal)
		_, err := retried(context.Background(), 1)

		if err != errFatal {
			t.Fatalf("got error %v, want %v", err, errFatal)
		}
		if calls != 3 {
			t.Fatalf("fn called %d times, want 3", calls)
		}
	})

	t.Run("nil predicate retries all errors", func(t *testing.T) {
		calls := 0
		// alwaysFail always returns an error.
		alwaysFail := func(_ context.Context, _ int) (int, error) {
			calls++
			return 0, fmt.Errorf("fail %d", calls)
		}

		retried := call.Retry(3, call.ConstantBackoff(0), nil, alwaysFail)
		_, err := retried(context.Background(), 1)

		if err == nil {
			t.Fatal("expected error")
		}
		if calls != 3 {
			t.Fatalf("fn called %d times, want 3", calls)
		}
	})
}

func TestRetryPanics(t *testing.T) {
	// doubleIt doubles the input.
	doubleIt := func(_ context.Context, n int) (int, error) { return n * 2, nil }

	t.Run("nil fn", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()
		call.Retry[int, int](3, call.ConstantBackoff(0), nil, nil)
	})

	t.Run("nil backoff", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()
		call.Retry(3, nil, nil, doubleIt)
	})

	t.Run("maxAttempts zero", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()
		call.Retry(0, call.ConstantBackoff(0), nil, doubleIt)
	})

	t.Run("maxAttempts negative", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()
		call.Retry(-1, call.ConstantBackoff(0), nil, doubleIt)
	})
}

func TestConstantBackoff(t *testing.T) {
	backoff := call.ConstantBackoff(500 * time.Millisecond)

	for _, n := range []int{0, 1, 5, 10} {
		got := backoff(n)
		if got != 500*time.Millisecond {
			t.Fatalf("backoff(%d) = %v, want 500ms", n, got)
		}
	}
}

func TestExponentialBackoff(t *testing.T) {
	t.Run("returns values in expected range", func(t *testing.T) {
		backoff := call.ExponentialBackoff(100 * time.Millisecond)

		for n := 0; n < 10; n++ {
			max := 100 * time.Millisecond << n

			for range 100 {
				got := backoff(n)
				if got < 0 || got >= max {
					t.Fatalf("backoff(%d) = %v, want [0, %v)", n, got, max)
				}
			}
		}
	})

	t.Run("overflow guard returns non-negative", func(t *testing.T) {
		backoff := call.ExponentialBackoff(time.Second)

		for range 100 {
			got := backoff(62)
			if got < 0 {
				t.Fatalf("backoff(62) = %v, want non-negative", got)
			}
		}
	})

	t.Run("panics on zero initial", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()
		call.ExponentialBackoff(0)
	})

	t.Run("panics on negative initial", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()
		call.ExponentialBackoff(-time.Second)
	})
}

func TestRetryComposesWithThrottle(t *testing.T) {
	calls := 0
	// failOnceThenDouble fails once, then doubles.
	failOnceThenDouble := func(_ context.Context, n int) (int, error) {
		calls++
		if calls < 2 {
			return 0, fmt.Errorf("fail")
		}
		return n * 2, nil
	}

	composed := call.Retry(3, call.ConstantBackoff(0), nil, call.Throttle(1, failOnceThenDouble))
	got, err := composed(context.Background(), 5)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 10 {
		t.Fatalf("got %d, want 10", got)
	}
	if calls != 2 {
		t.Fatalf("fn called %d times, want 2", calls)
	}
}
