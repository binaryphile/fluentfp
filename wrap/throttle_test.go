package wrap_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/binaryphile/fluentfp/wrap"
)

// --- Throttle ---

func TestThrottleBasic(t *testing.T) {
	// doubleIt doubles the input.
	doubleIt := func(_ context.Context, n int) (int, error) { return n * 2, nil }

	throttled := wrap.Func(doubleIt).With(wrap.Features{Throttle: wrap.Throttle(3)})
	got, err := throttled(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 10 {
		t.Fatalf("got %d, want 10", got)
	}
}

func TestThrottleError(t *testing.T) {
	// failAlways returns an error.
	failAlways := func(_ context.Context, _ int) (int, error) {
		return 0, fmt.Errorf("fail")
	}

	throttled := wrap.Func(failAlways).With(wrap.Features{Throttle: wrap.Throttle(1)})
	_, err := throttled(context.Background(), 1)
	if err == nil || err.Error() != "fail" {
		t.Fatalf("expected 'fail' error, got %v", err)
	}
}

func TestThrottleConcurrencyBound(t *testing.T) {
	const maxConcurrent = 3
	var active atomic.Int32
	var maxSeen atomic.Int32

	// trackConcurrency increments active count, records max, sleeps, then decrements.
	trackConcurrency := func(_ context.Context, _ int) (int, error) {
		cur := active.Add(1)
		for {
			old := maxSeen.Load()
			if cur <= old || maxSeen.CompareAndSwap(old, cur) {
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
		active.Add(-1)
		return 0, nil
	}

	throttled := wrap.Func(trackConcurrency).With(wrap.Features{Throttle: wrap.Throttle(maxConcurrent)})

	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			throttled(context.Background(), 0) //nolint:errcheck
			done <- struct{}{}
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	if maxSeen.Load() > int32(maxConcurrent) {
		t.Fatalf("max concurrent %d exceeded limit %d", maxSeen.Load(), maxConcurrent)
	}
}

func TestThrottleContextCancelled(t *testing.T) {
	// blockForever blocks until context is cancelled.
	blockForever := func(ctx context.Context, _ int) (int, error) {
		<-ctx.Done()
		return 0, ctx.Err()
	}

	// Fill all slots.
	throttled := wrap.Func(blockForever).With(wrap.Features{Throttle: wrap.Throttle(1)})
	ctx, cancel := context.WithCancel(context.Background())

	go throttled(ctx, 0) //nolint:errcheck
	time.Sleep(10 * time.Millisecond)

	// Second call should block and return error when cancelled.
	ctx2, cancel2 := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel2()

	_, err := throttled(ctx2, 0)
	if err == nil {
		t.Fatal("expected context error")
	}

	cancel()
}

func TestThrottlePreCancelledContext(t *testing.T) {
	// neverCalled should not be invoked.
	neverCalled := func(_ context.Context, _ int) (int, error) {
		t.Fatal("fn should not be called with cancelled context")
		return 0, nil
	}

	throttled := wrap.Func(neverCalled).With(wrap.Features{Throttle: wrap.Throttle(1)})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := throttled(ctx, 0)
	if err == nil {
		t.Fatal("expected context error")
	}
}

func TestThrottlePanicReleasesSlot(t *testing.T) {
	calls := 0

	// maybePanic panics on first call, succeeds on second.
	maybePanic := func(_ context.Context, _ int) (int, error) {
		calls++
		if calls == 1 {
			panic("boom")
		}

		return 42, nil
	}

	throttled := wrap.Func(maybePanic).With(wrap.Features{Throttle: wrap.Throttle(1)})

	// First call: panics. Recover it.
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
		throttled(context.Background(), 0) //nolint:errcheck
	}()

	// Second call: must succeed (slot was released despite panic).
	got, err := throttled(context.Background(), 0)
	if err != nil {
		t.Fatalf("unexpected error after panic recovery: %v", err)
	}
	if got != 42 {
		t.Fatalf("got %d, want 42", got)
	}
}

func TestThrottleValidationPanics(t *testing.T) {
	// dummyFn is a placeholder function.
	dummyFn := func(_ context.Context, _ int) (int, error) { return 0, nil }

	tests := []struct {
		name string
		fn   func()
	}{
		{
			name: "n_zero",
			fn:   func() { wrap.Func(dummyFn).With(wrap.Features{Throttle: wrap.Throttle(0)}) },
		},
		{
			name: "n_negative",
			fn:   func() { wrap.Func(dummyFn).With(wrap.Features{Throttle: wrap.Throttle(-1)}) },
		},
		{
			name: "nil_fn",
			fn:   func() { wrap.Func[int, int](nil).With(wrap.Features{Throttle: wrap.Throttle(1)}) },
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

// --- ThrottleWeighted ---

func TestThrottleWeightedBasic(t *testing.T) {
	// doubleIt doubles the input.
	doubleIt := func(_ context.Context, n int) (int, error) { return n * 2, nil }
	// unitCost returns 1 for any input.
	unitCost := func(_ int) int { return 1 }

	throttled := wrap.Func(doubleIt).Apply(wrap.WithThrottleWeighted[int, int](10, unitCost))
	got, err := throttled(context.Background(), 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 14 {
		t.Fatalf("got %d, want 14", got)
	}
}

func TestThrottleWeightedConcurrencyBound(t *testing.T) {
	const capacity = 10
	const costPerItem = 3 // max concurrent = 10/3 = 3

	var activeCost atomic.Int32
	var maxCost atomic.Int32

	// trackCost tracks active cost, records max, sleeps, then releases.
	trackCost := func(_ context.Context, _ int) (int, error) {
		cur := activeCost.Add(int32(costPerItem))
		for {
			old := maxCost.Load()
			if cur <= old || maxCost.CompareAndSwap(old, cur) {
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
		activeCost.Add(-int32(costPerItem))
		return 0, nil
	}

	// fixedCost returns costPerItem for any input.
	fixedCost := func(_ int) int { return costPerItem }

	throttled := wrap.Func(trackCost).Apply(wrap.WithThrottleWeighted[int, int](capacity, fixedCost))

	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			throttled(context.Background(), 0) //nolint:errcheck
			done <- struct{}{}
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	if maxCost.Load() > int32(capacity) {
		t.Fatalf("max concurrent cost %d exceeded capacity %d", maxCost.Load(), capacity)
	}
}

func TestThrottleWeightedVariableCosts(t *testing.T) {
	const capacity = 6
	var activeCost atomic.Int32
	var maxCost atomic.Int32

	// trackCost adds item cost, records max, sleeps, then releases.
	trackCost := func(_ context.Context, cost int) (int, error) {
		cur := activeCost.Add(int32(cost))
		for {
			old := maxCost.Load()
			if cur <= old || maxCost.CompareAndSwap(old, cur) {
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
		activeCost.Add(-int32(cost))
		return cost, nil
	}

	// itemCost returns the item itself as cost.
	itemCost := func(n int) int { return n }

	throttled := wrap.Func(trackCost).Apply(wrap.WithThrottleWeighted[int, int](capacity, itemCost))

	costs := []int{1, 2, 3, 1, 2, 3, 1, 2}
	done := make(chan struct{})
	for _, c := range costs {
		c := c
		go func() {
			throttled(context.Background(), c) //nolint:errcheck
			done <- struct{}{}
		}()
	}

	for range costs {
		<-done
	}

	if maxCost.Load() > int32(capacity) {
		t.Fatalf("max concurrent cost %d exceeded capacity %d", maxCost.Load(), capacity)
	}
}

func TestThrottleWeightedPartialAcquireRollback(t *testing.T) {
	// Capacity 5, variable cost via the input value.
	// Occupy 4 tokens with cost-1 holders, then attempt a cost-3 acquire
	// which can only get 1 token before blocking. Cancel to trigger rollback.
	// After rollback, verify the token was returned by running a cost-1 wrap.
	const capacity = 5

	started := make(chan struct{}, 4)

	// holderOrNoop holds a slot when input is 0, returns immediately otherwise.
	holderOrNoop := func(ctx context.Context, n int) (int, error) {
		if n == 0 {
			started <- struct{}{}
			<-ctx.Done()
			return 0, ctx.Err()
		}

		return n, nil
	}

	// variableCost uses input 0 → cost 1, input 3 → cost 3, input 1 → cost 1.
	variableCost := func(n int) int {
		if n == 0 {
			return 1
		}

		return n
	}

	throttled := wrap.Func(holderOrNoop).Apply(wrap.WithThrottleWeighted[int, int](capacity, variableCost))
	holderCtx, holderCancel := context.WithCancel(context.Background())

	// Occupy 4 tokens (cost-1 holders).
	for i := 0; i < 4; i++ {
		go throttled(holderCtx, 0) //nolint:errcheck
	}
	for i := 0; i < 4; i++ {
		<-started
	}

	// Try cost-3 acquire (only 1 token available). Should acquire 1, block on 2nd, timeout.
	acquireCtx, acquireCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer acquireCancel()

	_, err := throttled(acquireCtx, 3)
	if err == nil {
		t.Fatal("expected context error from partial acquire")
	}

	// Release holders.
	holderCancel()
	time.Sleep(20 * time.Millisecond)

	// After rollback, all 5 tokens should be free.
	// A cost-5 call proves all tokens were returned (cost-1 would succeed
	// even with a leaked token, since only 1 of 5 would be busy).
	checkCtx, checkCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer checkCancel()

	got, err := throttled(checkCtx, 5)
	if err != nil {
		t.Fatalf("expected success after rollback, got error: %v", err)
	}
	if got != 5 {
		t.Fatalf("got %d, want 5", got)
	}
}

func TestThrottleWeightedContextCancelled(t *testing.T) {
	// blockForever blocks until context is cancelled.
	blockForever := func(ctx context.Context, _ int) (int, error) {
		<-ctx.Done()
		return 0, ctx.Err()
	}
	// unitCost returns 1 for any input.
	unitCost := func(_ int) int { return 1 }

	throttled := wrap.Func(blockForever).Apply(wrap.WithThrottleWeighted[int, int](1, unitCost))
	ctx, cancel := context.WithCancel(context.Background())

	go throttled(ctx, 0) //nolint:errcheck
	time.Sleep(10 * time.Millisecond)

	ctx2, cancel2 := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel2()

	_, err := throttled(ctx2, 0)
	if err == nil {
		t.Fatal("expected context error")
	}

	cancel()
}

func TestThrottleWeightedPreCancelledContext(t *testing.T) {
	// neverCalled should not be invoked.
	neverCalled := func(_ context.Context, _ int) (int, error) {
		t.Fatal("fn should not be called with cancelled context")
		return 0, nil
	}
	// unitCost returns 1 for any input.
	unitCost := func(_ int) int { return 1 }

	throttled := wrap.Func(neverCalled).Apply(wrap.WithThrottleWeighted[int, int](1, unitCost))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := throttled(ctx, 0)
	if err == nil {
		t.Fatal("expected context error")
	}
}

func TestThrottleWeightedPanicReleasesSlot(t *testing.T) {
	const capacity = 3
	calls := 0

	// maybePanic panics on first call, succeeds on second.
	maybePanic := func(_ context.Context, _ int) (int, error) {
		calls++
		if calls == 1 {
			panic("boom")
		}

		return 42, nil
	}

	// fixedCost returns 3 (full capacity) for any input.
	fixedCost := func(_ int) int { return capacity }

	throttled := wrap.Func(maybePanic).Apply(wrap.WithThrottleWeighted[int, int](capacity, fixedCost))

	// First call: panics with all 3 tokens acquired. Recover it.
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
		throttled(context.Background(), 0) //nolint:errcheck
	}()

	// Second call: must succeed (all 3 tokens were released despite panic).
	got, err := throttled(context.Background(), 0)
	if err != nil {
		t.Fatalf("unexpected error after panic recovery: %v", err)
	}
	if got != 42 {
		t.Fatalf("got %d, want 42", got)
	}
}

func TestThrottleWeightedValidationPanics(t *testing.T) {
	// dummyFn is a placeholder function.
	dummyFn := func(_ context.Context, _ int) (int, error) { return 0, nil }
	// unitCost returns 1 for any input.
	unitCost := func(_ int) int { return 1 }

	tests := []struct {
		name string
		fn   func()
	}{
		{
			name: "capacity_zero",
			fn:   func() { wrap.Func(dummyFn).Apply(wrap.WithThrottleWeighted[int, int](0, unitCost)) },
		},
		{
			name: "capacity_negative",
			fn:   func() { wrap.Func(dummyFn).Apply(wrap.WithThrottleWeighted[int, int](-1, unitCost)) },
		},
		{
			name: "nil_cost",
			fn:   func() { wrap.Func(dummyFn).Apply(wrap.WithThrottleWeighted[int, int](1, nil)) },
		},
		{
			name: "nil_fn",
			fn:   func() { wrap.Func((func(context.Context, int) (int, error))(nil)).Apply(wrap.WithThrottleWeighted[int, int](1, unitCost)) },
		},
		{
			name: "cost_zero",
			fn: func() {
				// zeroCost returns 0 for any input.
				zeroCost := func(_ int) int { return 0 }
				throttled := wrap.Func(dummyFn).Apply(wrap.WithThrottleWeighted[int, int](10, zeroCost))
				throttled(context.Background(), 0) //nolint:errcheck
			},
		},
		{
			name: "cost_negative",
			fn: func() {
				// negativeCost returns -1 for any input.
				negativeCost := func(_ int) int { return -1 }
				throttled := wrap.Func(dummyFn).Apply(wrap.WithThrottleWeighted[int, int](10, negativeCost))
				throttled(context.Background(), 0) //nolint:errcheck
			},
		},
		{
			name: "cost_exceeds_capacity",
			fn: func() {
				// bigCost returns 11 for any input.
				bigCost := func(_ int) int { return 11 }
				throttled := wrap.Func(dummyFn).Apply(wrap.WithThrottleWeighted[int, int](10, bigCost))
				throttled(context.Background(), 0) //nolint:errcheck
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
