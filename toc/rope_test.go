package toc_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/binaryphile/fluentfp/toc"
)

func slowFn(ctx context.Context, n int) (int, error) {
	select {
	case <-time.After(50 * time.Millisecond):
		return n * 10, nil
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

func identityFn(_ context.Context, n int) (int, error) {
	return n, nil
}

func TestMaxWIPDefault(t *testing.T) {
	// Default MaxWIP = Capacity + Workers.
	ctx := context.Background()
	s := toc.Start(ctx, identityFn, toc.Options[int]{
		Capacity: 5,
		Workers:  2,
	})

	if got := s.MaxWIP(); got != 7 {
		t.Errorf("default MaxWIP = %d, want 7", got)
	}

	s.CloseInput()
	s.DiscardAndWait()
}

func TestMaxWIPStatic(t *testing.T) {
	// Static MaxWIP limits admission.
	ctx := context.Background()
	s := toc.Start(ctx, identityFn, toc.Options[int]{
		Capacity: 5,
		Workers:  2,
		MaxWIP:   3,
	})

	if got := s.MaxWIP(); got != 3 {
		t.Errorf("MaxWIP = %d, want 3", got)
	}

	s.CloseInput()
	s.DiscardAndWait()
}

func TestMaxWIPClampedToCeiling(t *testing.T) {
	ctx := context.Background()
	s := toc.Start(ctx, identityFn, toc.Options[int]{
		Capacity: 3,
		Workers:  1,
		MaxWIP:   100, // exceeds ceiling
	})

	if got := s.MaxWIP(); got != 4 {
		t.Errorf("MaxWIP = %d, want 4 (capacity+workers)", got)
	}

	s.CloseInput()
	s.DiscardAndWait()
}

func TestMaxWIPFloorOfOne(t *testing.T) {
	ctx := context.Background()
	s := toc.Start(ctx, identityFn, toc.Options[int]{
		Capacity: 3,
		Workers:  1,
	})

	applied := s.SetMaxWIP(0)
	if applied != 1 {
		t.Errorf("SetMaxWIP(0) = %d, want 1", applied)
	}
	if got := s.MaxWIP(); got != 1 {
		t.Errorf("MaxWIP = %d, want 1", got)
	}

	s.CloseInput()
	s.DiscardAndWait()
}

func TestSetMaxWIPClampsToCeiling(t *testing.T) {
	ctx := context.Background()
	s := toc.Start(ctx, identityFn, toc.Options[int]{
		Capacity: 2,
		Workers:  1,
	})

	applied := s.SetMaxWIP(100)
	if applied != 3 {
		t.Errorf("SetMaxWIP(100) = %d, want 3", applied)
	}

	s.CloseInput()
	s.DiscardAndWait()
}

func TestRopeBlocksAtLimit(t *testing.T) {
	// With MaxWIP=1 and a slow worker, the second Submit should block.
	ctx := context.Background()
	s := toc.Start(ctx, slowFn, toc.Options[int]{
		Capacity: 5,
		Workers:  1,
		MaxWIP:   1,
	})

	// First submit should succeed immediately.
	if err := s.Submit(ctx, 1); err != nil {
		t.Fatalf("first Submit: %v", err)
	}

	// Second submit should block (MaxWIP=1, first is in-flight).
	blocked := make(chan struct{})
	go func() {
		s.Submit(ctx, 2)
		close(blocked)
	}()

	select {
	case <-blocked:
		t.Fatal("second Submit should have blocked")
	case <-time.After(20 * time.Millisecond):
		// Expected: still blocked.
	}

	// Drain output to let workers complete and unblock.
	go func() {
		for range s.Out() {
		}
	}()

	// Wait for the second submit to complete.
	select {
	case <-blocked:
	case <-time.After(time.Second):
		t.Fatal("second Submit never unblocked")
	}

	s.CloseInput()
	s.DiscardAndWait()
}

func TestSetMaxWIPIncrease(t *testing.T) {
	// Start with MaxWIP=1, submit 2, increase to 2, second should unblock.
	ctx := context.Background()
	s := toc.Start(ctx, slowFn, toc.Options[int]{
		Capacity: 5,
		Workers:  2,
		MaxWIP:   1,
	})

	if err := s.Submit(ctx, 1); err != nil {
		t.Fatalf("first Submit: %v", err)
	}

	unblocked := make(chan struct{})
	go func() {
		s.Submit(ctx, 2)
		close(unblocked)
	}()

	// Give the goroutine time to block.
	time.Sleep(10 * time.Millisecond)

	select {
	case <-unblocked:
		t.Fatal("should be blocked before SetMaxWIP")
	default:
	}

	// Increase limit — should wake the blocked Submit.
	s.SetMaxWIP(2)

	select {
	case <-unblocked:
	case <-time.After(time.Second):
		t.Fatal("SetMaxWIP(2) did not unblock second Submit")
	}

	s.CloseInput()

	for range s.Out() {
	}

	s.Wait()
}

func TestSetMaxWIPDecrease(t *testing.T) {
	// Decrease doesn't affect already-admitted items, but blocks new ones.
	ctx := context.Background()
	s := toc.Start(ctx, slowFn, toc.Options[int]{
		Capacity: 5,
		Workers:  2,
		MaxWIP:   3,
	})

	// Submit 2 items.
	s.Submit(ctx, 1)
	s.Submit(ctx, 2)

	// Decrease to 1 — third submit should block.
	s.SetMaxWIP(1)

	blocked := make(chan struct{})
	go func() {
		s.Submit(ctx, 3)
		close(blocked)
	}()

	select {
	case <-blocked:
		t.Fatal("third Submit should block after SetMaxWIP(1)")
	case <-time.After(20 * time.Millisecond):
	}

	// Drain and finish.
	go func() {
		for range s.Out() {
		}
	}()

	select {
	case <-blocked:
	case <-time.After(2 * time.Second):
		t.Fatal("third Submit never unblocked")
	}

	s.CloseInput()
	s.DiscardAndWait()
}

func TestRopeCancelledSubmit(t *testing.T) {
	// Cancelled context should return ctx.Err from rope wait.
	ctx := context.Background()
	s := toc.Start(ctx, slowFn, toc.Options[int]{
		Capacity: 5,
		Workers:  1,
		MaxWIP:   1,
	})

	// Fill the slot.
	s.Submit(ctx, 1)

	// Try to submit with a cancelled context.
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()

	err := s.Submit(cancelCtx, 2)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}

	s.CloseInput()

	for range s.Out() {
	}

	s.Wait()
}

func TestRopeStats(t *testing.T) {
	ctx := context.Background()
	s := toc.Start(ctx, slowFn, toc.Options[int]{
		Capacity: 5,
		Workers:  1,
		MaxWIP:   1,
	})

	// First submit goes through.
	s.Submit(ctx, 1)

	// Second submit will block on rope.
	done := make(chan struct{})
	go func() {
		s.Submit(ctx, 2)
		close(done)
	}()

	// Give time to block.
	time.Sleep(10 * time.Millisecond)

	stats := s.Stats()
	if stats.MaxWIP != 1 {
		t.Errorf("Stats.MaxWIP = %d, want 1", stats.MaxWIP)
	}
	if stats.RopeWaitCount < 1 {
		t.Errorf("Stats.RopeWaitCount = %d, want >= 1", stats.RopeWaitCount)
	}

	// Drain.
	go func() {
		for range s.Out() {
		}
	}()

	<-done
	s.CloseInput()
	s.DiscardAndWait()
}

func TestRopeConcurrentSubmitAndSetMaxWIP(t *testing.T) {
	// Race test: concurrent Submit + SetMaxWIP.
	ctx := context.Background()
	s := toc.Start(ctx, identityFn, toc.Options[int]{
		Capacity:        10,
		Workers:         4,
		MaxWIP:          2,
		ContinueOnError: true,
	})

	var wg sync.WaitGroup
	var submitted atomic.Int64

	// Submitters.
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 50 {
				if err := s.Submit(ctx, i); err == nil {
					submitted.Add(1)
				}
			}
		}()
	}

	// Dynamic adjuster.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range 20 {
			s.SetMaxWIP(1)
			time.Sleep(time.Millisecond)
			s.SetMaxWIP(14) // capacity + workers
			time.Sleep(time.Millisecond)
		}
	}()

	// Drain output in background.
	go func() {
		for range s.Out() {
		}
	}()

	wg.Wait()
	s.CloseInput()
	s.DiscardAndWait()

	if submitted.Load() != 200 {
		t.Errorf("submitted = %d, want 200", submitted.Load())
	}
}

func TestRopeBackwardCompatible(t *testing.T) {
	// MaxWIP=0 means default (Capacity+Workers) — existing behavior unchanged.
	ctx := context.Background()
	s := toc.Start(ctx, identityFn, toc.Options[int]{
		Capacity: 10,
		Workers:  2,
	})

	// Should be able to submit up to capacity without blocking.
	for i := range 10 {
		if err := s.Submit(ctx, i); err != nil {
			t.Fatalf("Submit(%d): %v", i, err)
		}
	}

	s.CloseInput()

	for range s.Out() {
	}

	s.Wait()
}
