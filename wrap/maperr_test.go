package wrap_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/binaryphile/fluentfp/wrap"
)

func TestMapErrSuccess(t *testing.T) {
	called := false

	// trackMapper records whether it was called.
	trackMapper := func(err error) error {
		called = true

		return err
	}
	// doubleIt doubles the input.
	doubleIt := func(_ context.Context, n int) (int, error) { return n * 2, nil }

	wrapped := wrap.Func(doubleIt).WithMapError(trackMapper)
	got, err := wrapped(context.Background(), 5)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 10 {
		t.Fatalf("got %d, want 10", got)
	}
	if called {
		t.Fatal("mapper should not be called on success")
	}
}

func TestMapErrTransformsError(t *testing.T) {
	errOriginal := errors.New("original")

	// addPrefix wraps err with a prefix.
	addPrefix := func(err error) error { return fmt.Errorf("wrapped: %w", err) }
	// alwaysFail always returns errOriginal.
	alwaysFail := func(_ context.Context, _ int) (int, error) { return 0, errOriginal }

	wrapped := wrap.Func(alwaysFail).WithMapError(addPrefix)
	_, err := wrapped(context.Background(), 0)

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "wrapped: original" {
		t.Fatalf("got error %q, want %q", err.Error(), "wrapped: original")
	}
}

func TestMapErrPassesExactErrorToMapper(t *testing.T) {
	errOriginal := errors.New("original")
	var received error

	// captureMapper records the error it receives.
	captureMapper := func(err error) error {
		received = err

		return err
	}
	// alwaysFail always returns errOriginal.
	alwaysFail := func(_ context.Context, _ int) (int, error) { return 0, errOriginal }

	wrapped := wrap.Func(alwaysFail).WithMapError(captureMapper)
	wrapped(context.Background(), 0)

	if received != errOriginal {
		t.Fatalf("mapper received %p, want %p (must be exact same error)", received, errOriginal)
	}
}

func TestMapErrPreservesResultOnError(t *testing.T) {
	errBoom := errors.New("boom")

	// failWithValue returns a non-zero value alongside an error.
	failWithValue := func(_ context.Context, _ int) (int, error) { return 42, errBoom }
	// identity returns err unchanged.
	identity := func(err error) error { return err }

	wrapped := wrap.Func(failWithValue).WithMapError(identity)
	got, _ := wrapped(context.Background(), 0)

	if got != 42 {
		t.Fatalf("result value = %d, want 42 (must be preserved on error path)", got)
	}
}

func TestMapErrPreservesErrorIdentity(t *testing.T) {
	errSentinel := errors.New("sentinel")

	// wrapWithContext wraps err preserving identity via %w.
	wrapWithContext := func(err error) error { return fmt.Errorf("context: %w", err) }
	// alwaysFail always returns errSentinel.
	alwaysFail := func(_ context.Context, _ int) (int, error) { return 0, errSentinel }

	wrapped := wrap.Func(alwaysFail).WithMapError(wrapWithContext)
	_, err := wrapped(context.Background(), 0)

	if !errors.Is(err, errSentinel) {
		t.Fatalf("errors.Is failed: wrapped error %v does not match sentinel %v", err, errSentinel)
	}
}

func TestMapErrOuterSeesRetryFinalError(t *testing.T) {
	calls := 0
	errBoom := errors.New("boom")

	// constBackoff always waits 0 for testing.
	constBackoff := wrap.Backoff(func(int) time.Duration { return 0 })

	// alwaysFail always returns an error and counts invocations.
	alwaysFail := func(_ context.Context, _ int) (int, error) {
		calls++

		return 0, errBoom
	}

	mapperCalls := 0

	// countingWrapper wraps err and counts invocations.
	countingWrapper := func(err error) error {
		mapperCalls++

		return fmt.Errorf("mapped: %w", err)
	}

	// Retry first (inner), then map error (outer).
	composed := wrap.Func(alwaysFail).WithRetry(3, constBackoff, nil).WithMapError(countingWrapper)
	_, err := composed(context.Background(), 0)

	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 3 {
		t.Fatalf("fn called %d times, want 3", calls)
	}
	if mapperCalls != 1 {
		t.Fatalf("mapper called %d times, want 1 (should run once after retries exhaust)", mapperCalls)
	}
}

func TestMapErrInnerMapsPerRetryAttempt(t *testing.T) {
	mapperCalls := 0
	errBoom := errors.New("boom")

	// constBackoff always waits 0 for testing.
	constBackoff := wrap.Backoff(func(int) time.Duration { return 0 })

	// countingWrapper wraps err and counts invocations.
	countingWrapper := func(err error) error {
		mapperCalls++

		return fmt.Errorf("mapped: %w", err)
	}
	// alwaysFail always returns errBoom.
	alwaysFail := func(_ context.Context, _ int) (int, error) { return 0, errBoom }

	// Map error first (inner), then retry (outer).
	composed := wrap.Func(alwaysFail).WithMapError(countingWrapper).WithRetry(3, constBackoff, nil)
	_, err := composed(context.Background(), 0)

	if err == nil {
		t.Fatal("expected error")
	}
	if mapperCalls != 3 {
		t.Fatalf("mapper called %d times, want 3 (should run per attempt)", mapperCalls)
	}
}

func TestMapErrOuterOnErrSeesMapping(t *testing.T) {
	errOriginal := errors.New("original")
	var observed error

	// observeErr captures the error for inspection.
	observeErr := func(err error) { observed = err }
	// addPrefix wraps err with a prefix.
	addPrefix := func(err error) error { return fmt.Errorf("prefix: %w", err) }
	// alwaysFail always returns errOriginal.
	alwaysFail := func(_ context.Context, _ int) (int, error) { return 0, errOriginal }

	// MapError first (inner), then OnError (outer) — observer sees mapped error.
	composed := wrap.Func(alwaysFail).WithMapError(addPrefix).WithOnError(observeErr)
	composed(context.Background(), 0)

	if observed == nil {
		t.Fatal("observer should have been called")
	}
	if observed.Error() != "prefix: original" {
		t.Fatalf("observer saw %q, want %q", observed.Error(), "prefix: original")
	}
}

func TestMapErrInnerOnErrSeesOriginal(t *testing.T) {
	errOriginal := errors.New("original")
	var observed error

	// observeErr captures the error for inspection.
	observeErr := func(err error) { observed = err }
	// addPrefix wraps err with a prefix.
	addPrefix := func(err error) error { return fmt.Errorf("prefix: %w", err) }
	// alwaysFail always returns errOriginal.
	alwaysFail := func(_ context.Context, _ int) (int, error) { return 0, errOriginal }

	// OnError first (inner), then MapError (outer) — observer sees original error.
	composed := wrap.Func(alwaysFail).WithOnError(observeErr).WithMapError(addPrefix)
	composed(context.Background(), 0)

	if observed != errOriginal {
		t.Fatalf("observer saw %p, want %p (should see original, not mapped)", observed, errOriginal)
	}
}

func TestMapErrValidationPanics(t *testing.T) {
	// dummyMapper is a placeholder mapper.
	dummyMapper := func(err error) error { return err }

	tests := []struct {
		name string
		fn   func()
	}{
		{
			name: "nil_fn",
			fn:   func() { wrap.Func[int, int](nil).WithMapError(dummyMapper) },
		},
		{
			name: "nil_mapper",
			fn: func() {
				// dummyFn is a placeholder function.
				dummyFn := func(_ context.Context, _ int) (int, error) { return 0, nil }
				wrap.Func(dummyFn).WithMapError(nil)
			},
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

func TestMapErrMapperReturnsNilPanics(t *testing.T) {
	// nilMapper always returns nil, violating the contract.
	nilMapper := func(_ error) error { return nil }
	// alwaysFail always returns an error.
	alwaysFail := func(_ context.Context, _ int) (int, error) { return 0, errors.New("fail") }

	wrapped := wrap.Func(alwaysFail).WithMapError(nilMapper)

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic when mapper returns nil")
		}
	}()

	wrapped(context.Background(), 0)
}
