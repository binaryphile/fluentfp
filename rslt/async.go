package rslt

import (
	"context"
	"fmt"
	"runtime/debug"
)

// AsyncResult holds the outcome of a background goroutine launched
// by [RunAsync]. Use [AsyncResult.Wait] to block for the result or
// [AsyncResult.Done] to compose with select.
type AsyncResult[R any] struct {
	done chan struct{}
	val  R
	err  error
}

// RunAsync launches fn in a goroutine and returns a handle to wait
// on the result. Panics in fn are recovered and wrapped as
// [*PanicError] with a stack trace — an unrecovered panic in a
// background goroutine would crash the entire process.
//
// The caller owns ctx and is responsible for cancellation.
// RunAsync does not add its own cancel or timeout.
//
// Panics if fn is nil.
func RunAsync[R any](ctx context.Context, fn func(context.Context) (R, error)) *AsyncResult[R] {
	if fn == nil {
		panic("rslt.RunAsync: fn must not be nil")
	}

	a := &AsyncResult[R]{done: make(chan struct{})}
	go func() {
		defer close(a.done)
		defer func() {
			if r := recover(); r != nil {
				a.err = &PanicError{
					Value: r,
					Stack: debug.Stack(),
				}
			}
		}()
		a.val, a.err = fn(ctx)
	}()
	return a
}

// Wait blocks until the goroutine completes and returns the result.
// Safe to call from multiple goroutines and multiple times — always
// returns the same value.
func (a *AsyncResult[R]) Wait() (R, error) {
	<-a.done
	return a.val, a.err
}

// Done returns a channel that closes when the goroutine completes.
// Composable with select:
//
//	select {
//	case <-result.Done():
//	    val, err := result.Wait()
//	case <-ctx.Done():
//	    // timed out waiting
//	}
func (a *AsyncResult[R]) Done() <-chan struct{} {
	return a.done
}

// String returns a human-readable description of the async result's
// state (pending, ok, or error).
func (a *AsyncResult[R]) String() string {
	select {
	case <-a.done:
		if a.err != nil {
			return fmt.Sprintf("AsyncResult(err: %v)", a.err)
		}
		return fmt.Sprintf("AsyncResult(ok: %v)", a.val)
	default:
		return "AsyncResult(pending)"
	}
}
