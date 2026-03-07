package slice

import (
	"context"
	"runtime/debug"
	"sync"

	"github.com/binaryphile/fluentfp/result"
)

// FanOut applies fn to each element of ts concurrently with at most n goroutines.
// Each element gets its own goroutine (semaphore-bounded), enabling per-item scheduling
// suited for I/O-bound workloads with variable latency.
//
// Returns Mapper[result.Result[R]] with len == len(ts), where output[i] corresponds to ts[i].
//
// Cancellation guarantees:
//  1. FanOut returns only after all started callbacks have returned. No goroutine leaks.
//  2. At most n callbacks execute concurrently. Semaphore enforced.
//  3. When ctx is cancelled, the scheduler stops launching new work promptly.
//     Unscheduled items get Err(ctx.Err()). Due to check-then-act races,
//     at most one additional callback may start after cancellation occurs.
//  4. In-flight callbacks continue until fn returns.
//  5. Callbacks may observe an already-cancelled context.
//  6. An already-cancelled ctx before FanOut entry: zero callbacks run.
//  7. fn errors do NOT cancel siblings. The caller controls fail-fast by cancelling ctx in fn.
//
// Panics if n <= 0, ctx is nil, or fn is nil.
func FanOut[T, R any](ctx context.Context, n int, ts []T, fn func(context.Context, T) (R, error)) Mapper[result.Result[R]] {
	return fanOut(ctx, n, ts, fn)
}

// FanOutEach applies fn to each element of ts concurrently with at most n goroutines.
// It is the side-effect variant of FanOut for operations that don't produce values.
//
// Returns []error with len == len(ts). Nil entries indicate success.
// Panics from fn are wrapped as *result.PanicError in the error slice,
// detectable via errors.As.
//
// Panics if n <= 0, ctx is nil, or fn is nil.
func FanOutEach[T any](ctx context.Context, n int, ts []T, fn func(context.Context, T) error) []error {
	if n <= 0 {
		panic("slice.FanOutEach: n must be > 0")
	}
	if ctx == nil {
		panic("slice.FanOutEach: ctx must not be nil")
	}
	if fn == nil {
		panic("slice.FanOutEach: fn must not be nil")
	}

	// wrapFn adapts a side-effect fn to the engine's (R, error) signature.
	wrapFn := func(ctx context.Context, t T) (struct{}, error) {
		return struct{}{}, fn(ctx, t)
	}

	results := fanOut(ctx, n, ts, wrapFn)

	errs := make([]error, len(results))
	for i, r := range results {
		if err, ok := r.GetErr(); ok {
			errs[i] = err
		}
	}

	return errs
}

// fanOut is the internal engine shared by FanOut and FanOutEach.
func fanOut[T, R any](ctx context.Context, n int, ts []T, fn func(context.Context, T) (R, error)) Mapper[result.Result[R]] {
	if n <= 0 {
		panic("slice.FanOut: n must be > 0")
	}
	if ctx == nil {
		panic("slice.FanOut: ctx must not be nil")
	}
	if fn == nil {
		panic("slice.FanOut: fn must not be nil")
	}

	if len(ts) == 0 {
		return Mapper[result.Result[R]]{}
	}

	results := make([]result.Result[R], len(ts))

	// Already-cancelled ctx: fill all slots, no callbacks run.
	if err := ctx.Err(); err != nil {
		for i := range results {
			results[i] = result.Err[R](err)
		}

		return results
	}

	if n > len(ts) {
		n = len(ts)
	}

	if n == 1 {
		for i, t := range ts {
			if err := ctx.Err(); err != nil {
				for j := i; j < len(results); j++ {
					results[j] = result.Err[R](err)
				}

				break
			}

			results[i] = runItem(ctx, t, fn)
		}

		return results
	}

	sem := make(chan struct{}, n)
	var wg sync.WaitGroup

loop:
	for i, t := range ts {
		// Pre-select cancellation check.
		if err := ctx.Err(); err != nil {
			for j := i; j < len(results); j++ {
				results[j] = result.Err[R](err)
			}

			break
		}

		select {
		case <-ctx.Done():
			for j := i; j < len(results); j++ {
				results[j] = result.Err[R](ctx.Err())
			}

			break loop
		case sem <- struct{}{}:
		}

		// Post-select cancellation check — narrow the race window.
		if err := ctx.Err(); err != nil {
			<-sem

			for j := i; j < len(results); j++ {
				results[j] = result.Err[R](err)
			}

			break
		}

		wg.Add(1)

		go func(i int, t T) {
			defer wg.Done()
			defer func() { <-sem }()

			results[i] = runItem(ctx, t, fn)
		}(i, t)
	}

	wg.Wait()

	return results
}

// runItem calls fn with panic recovery. Named return enables defer/recover to set the result.
func runItem[T, R any](ctx context.Context, t T, fn func(context.Context, T) (R, error)) (res result.Result[R]) {
	defer func() {
		if v := recover(); v != nil {
			res = result.Err[R](&result.PanicError{Value: v, Stack: debug.Stack()})
		}
	}()

	r, err := fn(ctx, t)
	if err != nil {
		return result.Err[R](err)
	}

	return result.Ok(r)
}
