package wrap_test

import (
	"context"
	"errors"
	"testing"

	"github.com/binaryphile/fluentfp/wrap"
)

func TestOnErrBasicSuccess(t *testing.T) {
	called := false

	// onErr records whether it was called.
	onErr := func(_ error) { called = true }
	// doubleIt doubles the input.
	doubleIt := func(_ context.Context, n int) (int, error) { return n * 2, nil }

	wrapped := wrap.Func(doubleIt).OnError(onErr)
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
	var gotErr error

	// onErr captures the error it receives.
	onErr := func(err error) { gotErr = err }
	errBoom := errors.New("boom")
	// failingFn always returns an error.
	failingFn := func(_ context.Context, _ int) (int, error) { return 0, errBoom }

	wrapped := wrap.Func(failingFn).OnError(onErr)
	_, err := wrapped(context.Background(), 5)

	if !errors.Is(err, errBoom) {
		t.Fatalf("got error %v, want %v", err, errBoom)
	}
	if !errors.Is(gotErr, errBoom) {
		t.Fatalf("onErr received %v, want %v", gotErr, errBoom)
	}
}

func TestOnErrWithContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// onErr cancels the context on first error.
	onErr := func(_ error) { cancel() }
	errBoom := errors.New("boom")
	// failingFn always returns an error.
	failingFn := func(_ context.Context, _ int) (int, error) { return 0, errBoom }

	wrapped := wrap.Func(failingFn).OnError(onErr)
	_, err := wrapped(ctx, 5)

	if !errors.Is(err, errBoom) {
		t.Fatalf("got error %v, want %v", err, errBoom)
	}
	if ctx.Err() == nil {
		t.Fatal("context should be cancelled after error")
	}
}


func TestOnErrValidationPanics(t *testing.T) {
	// dummyOnErr is a placeholder side-effect.
	dummyOnErr := func(error) {}

	tests := []struct {
		name string
		fn   func()
	}{
		{
			name: "nil_fn",
			fn:   func() { wrap.Func[int, int](nil).OnError(dummyOnErr) },
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
