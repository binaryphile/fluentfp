package wrap_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/binaryphile/fluentfp/wrap"
)

func TestRetry(t *testing.T) {
	t.Run("succeeds on first attempt", func(t *testing.T) {
		calls := 0
		// doubleIt doubles the input.
		doubleIt := func(_ context.Context, n int) (int, error) {
			calls++
			return n * 2, nil
		}

		retried := wrap.Func(doubleIt).WithRetry(3, wrap.Backoff(func(int) time.Duration { return 0 }), nil)
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

		retried := wrap.Func(failTwiceThenDouble).WithRetry(3, wrap.Backoff(func(int) time.Duration { return 0 }), nil)
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

		retried := wrap.Func(alwaysFail).WithRetry(3, wrap.Backoff(func(int) time.Duration { return 0 }), nil)
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

		retried := wrap.Func(alwaysFail).WithRetry(1, wrap.Backoff(func(int) time.Duration { return 0 }), nil)
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

		retried := wrap.Func(doubleIt).WithRetry(3, wrap.Backoff(func(int) time.Duration { return 0 }), nil)
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
		// failAndCancel fails and cancels context on first wrap.
		failAndCancel := func(_ context.Context, _ int) (int, error) {
			calls++
			if calls == 1 {
				cancel()
			}
			return 0, fmt.Errorf("fail")
		}

		retried := wrap.Func(failAndCancel).WithRetry(3, wrap.Backoff(func(int) time.Duration { return 10*time.Second }), nil)
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

		retried := wrap.Func(alwaysFatal).WithRetry(3, wrap.Backoff(func(int) time.Duration { return 0 }), isRetryable)
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

		retried := wrap.Func(retryableThenFatal).WithRetry(5, wrap.Backoff(func(int) time.Duration { return 0 }), isRetryable)
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

		retried := wrap.Func(alwaysFail).WithRetry(3, wrap.Backoff(func(int) time.Duration { return 0 }), nil)
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
		wrap.Func[int, int](nil).WithRetry(3, wrap.Backoff(func(int) time.Duration { return 0 }), nil)
	})

	t.Run("nil backoff", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()
		wrap.Func(doubleIt).WithRetry(3, nil, nil)
	})

	t.Run("maxAttempts zero", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()
		wrap.Func(doubleIt).WithRetry(0, wrap.Backoff(func(int) time.Duration { return 0 }), nil)
	})

	t.Run("maxAttempts negative", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()
		wrap.Func(doubleIt).WithRetry(-1, wrap.Backoff(func(int) time.Duration { return 0 }), nil)
	})
}

func TestExpBackoff(t *testing.T) {
	t.Run("returns values in expected range", func(t *testing.T) {
		backoff := wrap.ExpBackoff(100 * time.Millisecond)

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
		backoff := wrap.ExpBackoff(time.Second)

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
		wrap.ExpBackoff(0)
	})

	t.Run("panics on negative initial", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()
		wrap.ExpBackoff(-time.Second)
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

	// constBackoff always waits 0 for testing.
	constBackoff := wrap.Backoff(func(int) time.Duration { return 0 })

	composed := wrap.Func(failOnceThenDouble).WithThrottle(1).WithRetry(3, constBackoff, nil)
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
