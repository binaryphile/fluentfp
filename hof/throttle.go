package hof

import (
	"context"
	"sync"
)

// Throttle wraps fn with count-based concurrency control.
// At most n calls to fn execute concurrently. The returned function
// blocks until a slot is available, then calls fn.
// The returned function is safe for concurrent use from multiple goroutines.
// Panics if n <= 0 or fn is nil.
func Throttle[T, R any](n int, fn func(context.Context, T) (R, error)) func(context.Context, T) (R, error) {
	if n <= 0 {
		panic("hof.Throttle: n must be > 0")
	}
	if fn == nil {
		panic("hof.Throttle: fn must not be nil")
	}

	sem := make(chan struct{}, n)

	return func(ctx context.Context, t T) (R, error) {
		var zero R

		if err := ctx.Err(); err != nil {
			return zero, err
		}

		select {
		case sem <- struct{}{}:
			defer func() { <-sem }()

			return fn(ctx, t)
		case <-ctx.Done():
			return zero, ctx.Err()
		}
	}
}

// ThrottleWeighted wraps fn with cost-based concurrency control.
// The total cost of concurrently-executing calls never exceeds capacity.
// The returned function blocks until enough budget is available.
// The returned function is safe for concurrent use from multiple goroutines.
// Panics if capacity <= 0, cost is nil, or fn is nil.
// Per-call: panics if cost(t) <= 0 or cost(t) > capacity.
func ThrottleWeighted[T, R any](capacity int, cost func(T) int, fn func(context.Context, T) (R, error)) func(context.Context, T) (R, error) {
	if capacity <= 0 {
		panic("hof.ThrottleWeighted: capacity must be > 0")
	}
	if cost == nil {
		panic("hof.ThrottleWeighted: cost must not be nil")
	}
	if fn == nil {
		panic("hof.ThrottleWeighted: fn must not be nil")
	}

	sem := make(chan struct{}, capacity)

	// acquireMu serializes multi-token acquire to prevent deadlock.
	// Without it, N goroutines each partially acquiring tokens can fill
	// the channel, leaving all goroutines unable to complete acquisition.
	// This mirrors FanOutWeighted's sequential scheduling loop.
	var acquireMu sync.Mutex

	return func(ctx context.Context, t T) (R, error) {
		var zero R

		if err := ctx.Err(); err != nil {
			return zero, err
		}

		itemCost := cost(t)
		if itemCost <= 0 {
			panic("hof.ThrottleWeighted: cost must be > 0")
		}
		if itemCost > capacity {
			panic("hof.ThrottleWeighted: cost must be <= capacity")
		}

		// Serialize multi-token acquire to prevent deadlock.
		acquireMu.Lock()

		acquired := 0

		for acquired < itemCost {
			select {
			case sem <- struct{}{}:
				acquired++
			case <-ctx.Done():
				// Rollback: release already-acquired tokens.
				for k := 0; k < acquired; k++ {
					<-sem
				}

				acquireMu.Unlock()

				return zero, ctx.Err()
			}
		}

		acquireMu.Unlock()

		defer func() {
			for k := 0; k < itemCost; k++ {
				<-sem
			}
		}()

		return fn(ctx, t)
	}
}
