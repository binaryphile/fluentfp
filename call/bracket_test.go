package call_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/binaryphile/fluentfp/call"
)

func TestBracket_happyPath(t *testing.T) {
	var order []string

	// trackAcquire records "acquire" and returns a release that records "release".
	trackAcquire := func(_ context.Context, _ int) (func(), error) {
		order = append(order, "acquire")
		return func() { order = append(order, "release") }, nil
	}

	// trackFn records "fn" and doubles the input.
	trackFn := func(_ context.Context, n int) (int, error) {
		order = append(order, "fn")
		return n * 2, nil
	}

	wrapped := call.From(trackFn).With(call.Bracket[int, int](trackAcquire))
	got, err := wrapped(context.Background(), 5)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 10 {
		t.Fatalf("got %d, want 10", got)
	}
	if len(order) != 3 || order[0] != "acquire" || order[1] != "fn" || order[2] != "release" {
		t.Fatalf("order = %v, want [acquire fn release]", order)
	}
}

func TestBracket_acquireError(t *testing.T) {
	fnCalled := false
	releaseCalled := false

	// failAcquire always returns an error.
	failAcquire := func(_ context.Context, _ int) (func(), error) {
		return nil, fmt.Errorf("acquire failed")
	}

	fn := func(_ context.Context, n int) (int, error) {
		fnCalled = true
		return n, nil
	}

	wrapped := call.From(fn).With(call.Bracket[int, int](failAcquire))
	_, err := wrapped(context.Background(), 1)

	if err == nil || err.Error() != "acquire failed" {
		t.Fatalf("expected 'acquire failed', got %v", err)
	}
	if fnCalled {
		t.Fatal("fn should not be called when acquire fails")
	}
	if releaseCalled {
		t.Fatal("release should not be called when acquire fails")
	}
}

func TestBracket_fnError(t *testing.T) {
	releaseCalled := false

	// okAcquire succeeds and returns a release that records the call.
	okAcquire := func(_ context.Context, _ int) (func(), error) {
		return func() { releaseCalled = true }, nil
	}

	// failFn always returns an error.
	failFn := func(_ context.Context, _ int) (int, error) {
		return 0, fmt.Errorf("fn failed")
	}

	wrapped := call.From(failFn).With(call.Bracket[int, int](okAcquire))
	_, err := wrapped(context.Background(), 1)

	if err == nil || err.Error() != "fn failed" {
		t.Fatalf("expected 'fn failed', got %v", err)
	}
	if !releaseCalled {
		t.Fatal("release must be called even when fn errors")
	}
}

func TestBracket_fnPanic(t *testing.T) {
	releaseCalled := false

	okAcquire := func(_ context.Context, _ int) (func(), error) {
		return func() { releaseCalled = true }, nil
	}

	// panicFn always panics.
	panicFn := func(_ context.Context, _ int) (int, error) {
		panic("boom")
	}

	wrapped := call.From(panicFn).With(call.Bracket[int, int](okAcquire))

	func() {
		defer func() {
			v := recover()
			if v == nil {
				t.Fatal("expected panic to propagate")
			}
			if v != "boom" {
				t.Fatalf("got panic %v, want 'boom'", v)
			}
		}()
		wrapped(context.Background(), 1) //nolint:errcheck
	}()

	if !releaseCalled {
		t.Fatal("release must be called even when fn panics")
	}
}

func TestBracket_nilAcquirePanics(t *testing.T) {
	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("expected panic for nil acquire")
		}
		if s, ok := v.(string); !ok || s != "call.Bracket: acquire must not be nil" {
			t.Fatalf("unexpected panic: %v", v)
		}
	}()

	call.Bracket[int, int](nil)
}

func TestBracket_concurrent(t *testing.T) {
	var active atomic.Int32
	var maxSeen atomic.Int32

	// countAcquire tracks concurrent acquisitions.
	countAcquire := func(_ context.Context, _ int) (func(), error) {
		cur := active.Add(1)
		for {
			old := maxSeen.Load()
			if cur <= old || maxSeen.CompareAndSwap(old, cur) {
				break
			}
		}
		return func() { active.Add(-1) }, nil
	}

	// slowFn simulates work.
	slowFn := func(_ context.Context, n int) (int, error) {
		return n, nil
	}

	wrapped := call.From(slowFn).With(call.Bracket[int, int](countAcquire))

	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			wrapped(context.Background(), 1) //nolint:errcheck
			done <- struct{}{}
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}

	if active.Load() != 0 {
		t.Fatalf("active count should be 0 after all calls, got %d", active.Load())
	}
}

func TestBracket_acquireCapturesState(t *testing.T) {
	var acquired, released atomic.Int32

	// countingAcquire increments acquired, returns release that increments released.
	countingAcquire := func(_ context.Context, _ int) (func(), error) {
		acquired.Add(1)
		return func() { released.Add(1) }, nil
	}

	fn := func(_ context.Context, n int) (int, error) { return n, nil }
	wrapped := call.From(fn).With(call.Bracket[int, int](countingAcquire))

	for i := 0; i < 5; i++ {
		wrapped(context.Background(), i) //nolint:errcheck
	}

	if acquired.Load() != 5 {
		t.Fatalf("acquired = %d, want 5", acquired.Load())
	}
	if released.Load() != 5 {
		t.Fatalf("released = %d, want 5", released.Load())
	}
}

func TestBracket_acquireContextError(t *testing.T) {
	// acquire checks context
	ctxAcquire := func(ctx context.Context, _ int) (func(), error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		return func() {}, nil
	}

	fn := func(_ context.Context, n int) (int, error) { return n, nil }
	wrapped := call.From(fn).With(call.Bracket[int, int](ctxAcquire))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := wrapped(ctx, 1)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
