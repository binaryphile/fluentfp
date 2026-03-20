package rslt

import (
	"context"
	"fmt"
	"runtime/debug"
)

// asyncState holds the mutable state shared between the launched
// goroutine and all AsyncResult handles. AsyncResult wraps a
// pointer to this, so copies of AsyncResult are safe.
type asyncState[R any] struct {
	done chan struct{}
	val  R
	err  error
}

// AsyncResult holds the outcome of a background goroutine launched
// by [RunAsync]. Use [AsyncResult.Wait] to block for the result or
// [AsyncResult.Done] to compose with select.
//
// AsyncResult is safe to copy — copies share the same underlying
// state. The zero value is not valid; use [RunAsync] to create.
type AsyncResult[R any] struct {
	s *asyncState[R]
}

// RunAsync launches fn in a goroutine and returns a handle to wait
// on the result. Panics in fn are recovered and wrapped as
// [*PanicError] with a stack trace — an unrecovered panic in a
// background goroutine would crash the entire process.
//
// Only panics on fn's goroutine are recovered. If fn launches
// additional goroutines, panics in those are not caught.
//
// The caller owns ctx and is responsible for cancellation.
// RunAsync does not add its own cancel or timeout.
//
// Panics if fn or ctx is nil.
func RunAsync[R any](ctx context.Context, fn func(context.Context) (R, error)) AsyncResult[R] {
	if ctx == nil {
		panic("rslt.RunAsync: ctx must not be nil")
	}
	if fn == nil {
		panic("rslt.RunAsync: fn must not be nil")
	}

	st := &asyncState[R]{done: make(chan struct{})}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				st.err = &PanicError{
					Value: r,
					Stack: debug.Stack(),
				}
			}
			close(st.done) // always last — establishes happens-before for Wait
		}()
		st.val, st.err = fn(ctx)
	}()
	return AsyncResult[R]{s: st}
}

// Wait blocks until the goroutine completes and returns the result.
// Safe to call from multiple goroutines and multiple times — always
// returns the same value.
func (a AsyncResult[R]) Wait() (R, error) {
	<-a.s.done
	return a.s.val, a.s.err
}

// Done returns a channel that closes when the goroutine completes.
// Composable with select:
//
//	select {
//	case <-resultOption.Done():
//	    val, err := resultOption.Wait()
//	case <-ctx.Done():
//	    // timed out waiting
//	}
func (a AsyncResult[R]) Done() <-chan struct{} {
	return a.s.done
}

// String returns a human-readable description of the async result's
// state (pending, ok, or error).
func (a AsyncResult[R]) String() string {
	select {
	case <-a.s.done:
		if a.s.err != nil {
			return fmt.Sprintf("AsyncResult(err: %v)", a.s.err)
		}
		return fmt.Sprintf("AsyncResult(ok: %v)", a.s.val)
	default:
		return "AsyncResult(pending)"
	}
}
