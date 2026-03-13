package slice

import (
	"context"
	"runtime/debug"
	"sync"

	"github.com/binaryphile/fluentfp/rslt"
)

// FanOut applies fn to each element of ts concurrently with at most n goroutines.
// Each element gets its own goroutine (semaphore-bounded), enabling per-item scheduling
// suited for I/O-bound workloads with variable latency.
//
// Returns Mapper[rslt.Result[R]] with len == len(ts), where output[i] corresponds to ts[i].
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
func FanOut[T, R any](ctx context.Context, n int, ts []T, fn func(context.Context, T) (R, error)) Mapper[rslt.Result[R]] {
	return fanOut(ctx, n, ts, fn)
}

// FanOutEach applies fn to each element of ts concurrently with at most n goroutines.
// It is the side-effect variant of FanOut for operations that don't produce values.
//
// Returns []error with len == len(ts). Nil entries indicate success.
// Panics from fn are wrapped as *rslt.PanicError in the error slice,
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
func fanOut[T, R any](ctx context.Context, n int, ts Mapper[T], fn func(context.Context, T) (R, error)) Mapper[rslt.Result[R]] {
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
		return Mapper[rslt.Result[R]]{}
	}

	results := make([]rslt.Result[R], len(ts))

	// Already-cancelled ctx: fill all slots, no callbacks run.
	if err := ctx.Err(); err != nil {
		for i := range results {
			results[i] = rslt.Err[R](err)
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
					results[j] = rslt.Err[R](err)
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
				results[j] = rslt.Err[R](err)
			}

			break
		}

		select {
		case <-ctx.Done():
			for j := i; j < len(results); j++ {
				results[j] = rslt.Err[R](ctx.Err())
			}

			break loop
		case sem <- struct{}{}:
		}

		// Post-select cancellation check — narrow the race window.
		if err := ctx.Err(); err != nil {
			<-sem

			for j := i; j < len(results); j++ {
				results[j] = rslt.Err[R](err)
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

// FanOutWeighted applies fn to each element of ts concurrently, bounded by a total
// cost budget rather than a fixed item count. Each item's cost is determined by the
// cost function, and at most capacity units of cost run concurrently.
//
// Returns Mapper[rslt.Result[R]] with len == len(ts), where output[i] corresponds to ts[i].
//
// Same cancellation guarantees as FanOut. Partial acquire rollback: if ctx cancels
// after acquiring some tokens for an item, the scheduler releases them and fills
// remaining items with ctx.Err().
//
// Panics if capacity <= 0, cost is nil, ctx is nil, or fn is nil.
// Per-item: panics if cost(t) <= 0 or cost(t) > capacity.
func FanOutWeighted[T, R any](ctx context.Context, capacity int, ts []T, cost func(T) int, fn func(context.Context, T) (R, error)) Mapper[rslt.Result[R]] {
	return fanOutWeighted(ctx, capacity, ts, cost, fn)
}

// FanOutEachWeighted applies fn to each element of ts concurrently, bounded by a
// total cost budget. It is the side-effect variant of FanOutWeighted.
//
// Returns []error with len == len(ts). Nil entries indicate success.
//
// Panics if capacity <= 0, cost is nil, ctx is nil, or fn is nil.
// Per-item: panics if cost(t) <= 0 or cost(t) > capacity.
func FanOutEachWeighted[T any](ctx context.Context, capacity int, ts []T, cost func(T) int, fn func(context.Context, T) error) []error {
	if capacity <= 0 {
		panic("slice.FanOutEachWeighted: capacity must be > 0")
	}
	if cost == nil {
		panic("slice.FanOutEachWeighted: cost must not be nil")
	}
	if ctx == nil {
		panic("slice.FanOutEachWeighted: ctx must not be nil")
	}
	if fn == nil {
		panic("slice.FanOutEachWeighted: fn must not be nil")
	}

	// wrapFn adapts a side-effect fn to the engine's (R, error) signature.
	wrapFn := func(ctx context.Context, t T) (struct{}, error) {
		return struct{}{}, fn(ctx, t)
	}

	results := fanOutWeighted(ctx, capacity, ts, cost, wrapFn)

	errs := make([]error, len(results))
	for i, r := range results {
		if err, ok := r.GetErr(); ok {
			errs[i] = err
		}
	}

	return errs
}

// fanOutWeighted is the internal engine for FanOutWeighted and FanOutEachWeighted.
func fanOutWeighted[T, R any](ctx context.Context, capacity int, ts Mapper[T], cost func(T) int, fn func(context.Context, T) (R, error)) Mapper[rslt.Result[R]] {
	if capacity <= 0 {
		panic("slice.FanOutWeighted: capacity must be > 0")
	}
	if cost == nil {
		panic("slice.FanOutWeighted: cost must not be nil")
	}
	if ctx == nil {
		panic("slice.FanOutWeighted: ctx must not be nil")
	}
	if fn == nil {
		panic("slice.FanOutWeighted: fn must not be nil")
	}

	if len(ts) == 0 {
		return Mapper[rslt.Result[R]]{}
	}

	results := make([]rslt.Result[R], len(ts))

	// Already-cancelled ctx: fill all slots, no callbacks run.
	if err := ctx.Err(); err != nil {
		for i := range results {
			results[i] = rslt.Err[R](err)
		}

		return results
	}

	sem := make(chan struct{}, capacity)
	var wg sync.WaitGroup

loop:
	for i, t := range ts {
		itemCost := cost(t)
		if itemCost <= 0 {
			panic("slice.FanOutWeighted: cost must be > 0")
		}
		if itemCost > capacity {
			panic("slice.FanOutWeighted: cost must be <= capacity")
		}

		// Pre-select cancellation check.
		if err := ctx.Err(); err != nil {
			for j := i; j < len(results); j++ {
				results[j] = rslt.Err[R](err)
			}

			break
		}

		// Multi-token acquire.
		acquired := 0
		cancelled := false

		for acquired < itemCost {
			select {
			case <-ctx.Done():
				cancelled = true
			case sem <- struct{}{}:
				acquired++
			}

			if cancelled {
				break
			}
		}

		if cancelled {
			// Rollback: release already-acquired tokens.
			for k := 0; k < acquired; k++ {
				<-sem
			}

			for j := i; j < len(results); j++ {
				results[j] = rslt.Err[R](ctx.Err())
			}

			break loop
		}

		// Post-select cancellation check — narrow the race window.
		if err := ctx.Err(); err != nil {
			for k := 0; k < itemCost; k++ {
				<-sem
			}

			for j := i; j < len(results); j++ {
				results[j] = rslt.Err[R](err)
			}

			break
		}

		wg.Add(1)

		go func(i int, t T, itemCost int) {
			defer wg.Done()
			defer func() {
				for k := 0; k < itemCost; k++ {
					<-sem
				}
			}()

			results[i] = runItem(ctx, t, fn)
		}(i, t, itemCost)
	}

	wg.Wait()

	return results
}

// FanOutAll applies fn to each element concurrently (at most n goroutines),
// returning all values if every call succeeds, or the first observed failure otherwise.
// On first failure (error or panic), the derived context is cancelled so cooperative
// work can stop early. Already-running callbacks continue until fn returns.
//
// The returned error is the first failure observed (by time, not index).
// Sibling context.Canceled errors from cancellation do not mask the root cause.
// Panics in fn are captured as *[rslt.PanicError] with the original stack trace.
//
// Derives a child context internally — the caller's context is never cancelled.
//
// Panics if n <= 0, ctx is nil, or fn is nil.
func FanOutAll[T, R any](ctx context.Context, n int, ts []T, fn func(context.Context, T) (R, error)) ([]R, error) {
	if ctx == nil {
		panic("slice.FanOutAll: ctx must not be nil")
	}

	if fn == nil {
		panic("slice.FanOutAll: fn must not be nil")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var causeOnce sync.Once
	var cause error

	// recordCause captures the first observed failure. Only the first call wins.
	recordCause := func(err error) {
		causeOnce.Do(func() { cause = err })
	}

	// cancelOnFail wraps fn to record the first failure, cancel remaining work,
	// and convert panics to errors with the original stack trace.
	cancelOnFail := func(ctx context.Context, t T) (r R, err error) {
		defer func() {
			if v := recover(); v != nil {
				err = &rslt.PanicError{Value: v, Stack: debug.Stack()}
				recordCause(err)
				cancel()
			}
		}()

		r, err = fn(ctx, t)
		if err != nil {
			recordCause(err)
			cancel()
		}

		return
	}

	results := fanOut(ctx, n, ts, cancelOnFail)

	if cause != nil {
		return nil, cause
	}

	return rslt.CollectAll([]rslt.Result[R](results))
}

// FanOutWeightedAll applies fn to each element concurrently, bounded by a total
// cost budget, returning all values if every call succeeds, or the first observed
// failure otherwise. On first failure (error or panic), the derived context is
// cancelled so cooperative work can stop early.
//
// Same semantics as [FanOutAll] — see its documentation for error selection,
// panic handling, and cancellation guarantees.
//
// Derives a child context internally — the caller's context is never cancelled.
//
// Panics if capacity <= 0, cost is nil, ctx is nil, or fn is nil.
// Per-item: panics if cost(t) <= 0 or cost(t) > capacity.
func FanOutWeightedAll[T, R any](ctx context.Context, capacity int, ts []T, cost func(T) int, fn func(context.Context, T) (R, error)) ([]R, error) {
	if ctx == nil {
		panic("slice.FanOutWeightedAll: ctx must not be nil")
	}

	if fn == nil {
		panic("slice.FanOutWeightedAll: fn must not be nil")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var causeOnce sync.Once
	var cause error

	// recordCause captures the first observed failure. Only the first call wins.
	recordCause := func(err error) {
		causeOnce.Do(func() { cause = err })
	}

	// cancelOnFail wraps fn to record the first failure, cancel remaining work,
	// and convert panics to errors with the original stack trace.
	cancelOnFail := func(ctx context.Context, t T) (r R, err error) {
		defer func() {
			if v := recover(); v != nil {
				err = &rslt.PanicError{Value: v, Stack: debug.Stack()}
				recordCause(err)
				cancel()
			}
		}()

		r, err = fn(ctx, t)
		if err != nil {
			recordCause(err)
			cancel()
		}

		return
	}

	results := fanOutWeighted(ctx, capacity, ts, cost, cancelOnFail)

	if cause != nil {
		return nil, cause
	}

	return rslt.CollectAll([]rslt.Result[R](results))
}

// runItem calls fn with panic recovery. Named return enables defer/recover to set the rslt.
func runItem[T, R any](ctx context.Context, t T, fn func(context.Context, T) (R, error)) (res rslt.Result[R]) {
	defer func() {
		if v := recover(); v != nil {
			res = rslt.Err[R](&rslt.PanicError{Value: v, Stack: debug.Stack()})
		}
	}()

	r, err := fn(ctx, t)
	if err != nil {
		return rslt.Err[R](err)
	}

	return rslt.Ok(r)
}
