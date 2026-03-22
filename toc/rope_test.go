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

	applied = s.SetMaxWIP(-5)
	if applied != 1 {
		t.Errorf("SetMaxWIP(-5) = %d, want 1", applied)
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
	if stats.WIPWaitCount < 1 {
		t.Errorf("Stats.WIPWaitCount = %d, want >= 1", stats.WIPWaitCount)
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

func TestRopeStageDoneUnblocksWaiter(t *testing.T) {
	// A waiter blocked in acquireAdmission must wake when stageDone fires.
	ctx, cancel := context.WithCancel(context.Background())
	s := toc.Start(ctx, slowFn, toc.Options[int]{
		Capacity: 5,
		Workers:  1,
		MaxWIP:   1,
	})

	s.Submit(ctx, 1)

	done := make(chan error, 1)
	go func() {
		done <- s.Submit(ctx, 2)
	}()

	time.Sleep(10 * time.Millisecond)

	// Cancel parent context — stageDone fires.
	cancel()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected error after stage cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("waiter not unblocked by stageDone")
	}

	for range s.Out() {
	}

	s.Wait()
}

func TestRopeCloseInputRejectsWaiters(t *testing.T) {
	// After CloseInput, blocked waiters should be rejected, not granted.
	ctx := context.Background()
	s := toc.Start(ctx, slowFn, toc.Options[int]{
		Capacity: 5,
		Workers:  1,
		MaxWIP:   1,
	})

	s.Submit(ctx, 1)

	errs := make(chan error, 3)
	for range 3 {
		go func() {
			errs <- s.Submit(ctx, 99)
		}()
	}

	time.Sleep(10 * time.Millisecond)
	s.CloseInput()

	for range 3 {
		select {
		case err := <-errs:
			if err == nil {
				t.Error("expected ErrClosed after CloseInput, got nil")
			}
		case <-time.After(time.Second):
			t.Fatal("waiter not rejected after CloseInput")
		}
	}

	for range s.Out() {
	}

	s.Wait()
}

func TestRopeAdmittedZeroAtQuiescence(t *testing.T) {
	ctx := context.Background()
	s := toc.Start(ctx, identityFn, toc.Options[int]{
		Capacity: 5,
		Workers:  2,
		MaxWIP:   3,
	})

	go func() {
		for range s.Out() {
		}
	}()

	for i := range 20 {
		s.Submit(ctx, i)
	}

	s.CloseInput()
	s.Wait()

	stats := s.Stats()
	if stats.Admitted != 0 {
		t.Errorf("Admitted = %d after quiescence, want 0", stats.Admitted)
	}
}

func TestRopePanicReleasesPermit(t *testing.T) {
	ctx := context.Background()
	panicFn := func(_ context.Context, n int) (int, error) {
		if n == 1 {
			panic("boom")
		}
		return n, nil
	}

	s := toc.Start(ctx, panicFn, toc.Options[int]{
		Capacity:        5,
		Workers:         1,
		MaxWIP:          2,
		ContinueOnError: true,
	})

	go func() {
		for range s.Out() {
		}
	}()

	s.Submit(ctx, 1)
	s.Submit(ctx, 2)
	s.Submit(ctx, 3)

	s.CloseInput()
	s.Wait()

	stats := s.Stats()
	if stats.Admitted != 0 {
		t.Errorf("Admitted = %d after panic + quiescence, want 0", stats.Admitted)
	}
	if stats.Completed != 3 {
		t.Errorf("Completed = %d, want 3", stats.Completed)
	}
}

func TestPauseAdmission(t *testing.T) {
	// Pause blocks Submit, resume unblocks.
	ctx := context.Background()
	s := toc.Start(ctx, identityFn, toc.Options[int]{Capacity: 5, Workers: 1, MaxWIP: 5})

	go func() { for range s.Out() {} }()

	s.PauseAdmission()

	if !s.Paused() {
		t.Fatal("expected Paused() == true")
	}

	blocked := make(chan struct{})
	go func() {
		s.Submit(ctx, 1)
		close(blocked)
	}()

	select {
	case <-blocked:
		t.Fatal("Submit should block when paused")
	case <-time.After(20 * time.Millisecond):
	}

	s.ResumeAdmission()

	select {
	case <-blocked:
	case <-time.After(time.Second):
		t.Fatal("Submit not unblocked after ResumeAdmission")
	}

	if s.Paused() {
		t.Fatal("expected Paused() == false after resume")
	}

	s.CloseInput()
	s.Wait()
}

func TestPauseIdempotent(t *testing.T) {
	ctx := context.Background()
	s := toc.Start(ctx, identityFn, toc.Options[int]{Capacity: 5, Workers: 1})

	// Double pause/resume should not panic or misbehave.
	s.PauseAdmission()
	s.PauseAdmission()
	s.ResumeAdmission()
	s.ResumeAdmission()

	// Should still work normally.
	go func() { for range s.Out() {} }()

	if err := s.Submit(ctx, 1); err != nil {
		t.Fatalf("Submit after double resume: %v", err)
	}

	s.CloseInput()
	s.Wait()
}

func TestPauseDoesNotRejectWaiters(t *testing.T) {
	// Paused waiters resume when unpaused, not rejected.
	ctx := context.Background()
	s := toc.Start(ctx, identityFn, toc.Options[int]{Capacity: 5, Workers: 1, MaxWIP: 5})

	go func() { for range s.Out() {} }()

	s.PauseAdmission()

	const N = 3
	errs := make(chan error, N)
	for i := range N {
		go func() { errs <- s.Submit(ctx, i) }()
	}

	time.Sleep(20 * time.Millisecond)

	s.ResumeAdmission()

	for range N {
		select {
		case err := <-errs:
			if err != nil {
				t.Errorf("waiter returned %v after resume, want nil", err)
			}
		case <-time.After(time.Second):
			t.Fatal("waiter not unblocked after resume")
		}
	}

	s.CloseInput()
	s.Wait()
}

func TestPauseStats(t *testing.T) {
	ctx := context.Background()
	s := toc.Start(ctx, identityFn, toc.Options[int]{Capacity: 5, Workers: 1})

	if s.Stats().Paused {
		t.Fatal("expected Paused=false initially")
	}

	s.PauseAdmission()

	if !s.Stats().Paused {
		t.Fatal("expected Paused=true after PauseAdmission")
	}

	s.ResumeAdmission()

	if s.Stats().Paused {
		t.Fatal("expected Paused=false after ResumeAdmission")
	}

	s.CloseInput()
	s.DiscardAndWait()
}

func TestPauseThenClose(t *testing.T) {
	// Close while paused rejects waiters with ErrClosed.
	ctx := context.Background()
	s := toc.Start(ctx, identityFn, toc.Options[int]{Capacity: 5, Workers: 1, MaxWIP: 5})

	go func() { for range s.Out() {} }()

	s.PauseAdmission()

	errCh := make(chan error, 1)
	go func() { errCh <- s.Submit(ctx, 1) }()

	time.Sleep(20 * time.Millisecond)

	s.CloseInput()

	select {
	case err := <-errCh:
		if err != toc.ErrClosed {
			t.Errorf("got %v, want ErrClosed", err)
		}
	case <-time.After(time.Second):
		t.Fatal("waiter not rejected after close")
	}

	s.Wait()
}

// ---------------------------------------------------------------------------
// Weighted admission tests
// ---------------------------------------------------------------------------

func TestMaxWIPWeightDefault(t *testing.T) {
	ctx := context.Background()
	s := toc.Start(ctx, identityFn, toc.Options[int]{
		Capacity: 5,
		Workers:  1,
		Weight:   func(n int) int64 { return int64(n * 100) },
	})

	go func() { for range s.Out() {} }()

	for i := 1; i <= 5; i++ {
		if err := s.Submit(ctx, i); err != nil {
			t.Fatalf("Submit(%d): %v", i, err)
		}
	}

	s.CloseInput()
	s.Wait()
}

func TestMaxWIPWeightBlocksHeavyItem(t *testing.T) {
	ctx := context.Background()
	s := toc.Start(ctx, slowFn, toc.Options[int]{
		Capacity:     5,
		Workers:      1,
		MaxWIP:       10,
		MaxWIPWeight: 100,
		Weight:       func(n int) int64 { return int64(n) },
	})

	go func() { for range s.Out() {} }()

	s.Submit(ctx, 80)

	blocked := make(chan struct{})
	go func() {
		s.Submit(ctx, 30)
		close(blocked)
	}()

	select {
	case <-blocked:
		t.Fatal("second Submit should block (weight 80+30 > 100)")
	case <-time.After(20 * time.Millisecond):
	}

	select {
	case <-blocked:
	case <-time.After(time.Second):
		t.Fatal("second Submit never unblocked")
	}

	s.CloseInput()
	s.Wait()
}

func TestMaxWIPWeightRejectsOversize(t *testing.T) {
	ctx := context.Background()
	s := toc.Start(ctx, identityFn, toc.Options[int]{
		Capacity:     5,
		Workers:      1,
		MaxWIPWeight: 50,
		Weight:       func(n int) int64 { return int64(n) },
	})

	go func() { for range s.Out() {} }()

	err := s.Submit(ctx, 51)
	if err != toc.ErrWeightExceedsLimit {
		t.Errorf("Submit(51): got %v, want ErrWeightExceedsLimit", err)
	}

	if err := s.Submit(ctx, 10); err != nil {
		t.Errorf("Submit(10): %v", err)
	}

	s.CloseInput()
	s.Wait()
}

func TestMaxWIPWeightAndCountBothEnforced(t *testing.T) {
	ctx := context.Background()
	s := toc.Start(ctx, slowFn, toc.Options[int]{
		Capacity:     5,
		Workers:      1,
		MaxWIP:       2,
		MaxWIPWeight: 1000,
		Weight:       func(n int) int64 { return 1 },
	})

	go func() { for range s.Out() {} }()

	s.Submit(ctx, 1)
	s.Submit(ctx, 2)

	blocked := make(chan struct{})
	go func() {
		s.Submit(ctx, 3)
		close(blocked)
	}()

	select {
	case <-blocked:
		t.Fatal("third Submit should block (count limit)")
	case <-time.After(20 * time.Millisecond):
	}

	select {
	case <-blocked:
	case <-time.After(2 * time.Second):
		t.Fatal("third Submit never unblocked")
	}

	s.CloseInput()
	s.Wait()
}

func TestMaxWIPWeightHOLBlocking(t *testing.T) {
	// Heavy item at queue head blocks lighter items behind it.
	workerGate := make(chan struct{})
	ctx := context.Background()
	s := toc.Start(ctx, func(ctx context.Context, n int) (int, error) {
		select {
		case <-workerGate:
			return n, nil
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	}, toc.Options[int]{
		Capacity:     5,
		Workers:      1,
		MaxWIP:       10,
		MaxWIPWeight: 50,
		Weight:       func(n int) int64 { return int64(n) },
	})

	go func() { for range s.Out() {} }()

	// Fill to weight 48. Worker blocks on gate.
	s.Submit(ctx, 48)

	// Heavy: weight 40. 48+40=88 > 50 → blocked.
	heavyDone := make(chan struct{})
	go func() {
		s.Submit(ctx, 40)
		close(heavyDone)
	}()
	time.Sleep(30 * time.Millisecond)

	// Light: weight 5. 48+5=53 > 50 → also blocked. Queues behind heavy.
	lightDone := make(chan struct{})
	go func() {
		s.Submit(ctx, 5)
		close(lightDone)
	}()
	time.Sleep(30 * time.Millisecond)

	// Neither should be done — worker is blocked.
	select {
	case <-lightDone:
		t.Fatal("light item should be blocked")
	case <-heavyDone:
		t.Fatal("heavy item should be blocked")
	default:
	}

	// Release worker → first item (48) completes → admittedWeight=0.
	// grantWaitersLocked: heavy (40) fits → granted. light (5) fits → granted.
	close(workerGate)

	select {
	case <-heavyDone:
	case <-time.After(time.Second):
		t.Fatal("heavy item never unblocked")
	}

	select {
	case <-lightDone:
	case <-time.After(time.Second):
		t.Fatal("light item never unblocked after heavy")
	}

	s.CloseInput()
	s.Wait()
}

func TestSetMaxWIPWeight(t *testing.T) {
	ctx := context.Background()
	s := toc.Start(ctx, slowFn, toc.Options[int]{
		Capacity:     5,
		Workers:      1,
		MaxWIP:       10,
		MaxWIPWeight: 30,
		Weight:       func(n int) int64 { return int64(n) },
	})

	go func() { for range s.Out() {} }()

	s.Submit(ctx, 25)

	blocked := make(chan struct{})
	go func() {
		s.Submit(ctx, 10)
		close(blocked)
	}()

	time.Sleep(10 * time.Millisecond)

	s.SetMaxWIPWeight(50)

	select {
	case <-blocked:
	case <-time.After(time.Second):
		t.Fatal("SetMaxWIPWeight(50) did not unblock")
	}

	s.CloseInput()
	s.Wait()
}

func TestMaxWIPWeightQuiescence(t *testing.T) {
	ctx := context.Background()
	s := toc.Start(ctx, identityFn, toc.Options[int]{
		Capacity:     5,
		Workers:      2,
		MaxWIPWeight: 100,
		Weight:       func(n int) int64 { return int64(n) },
	})

	go func() { for range s.Out() {} }()

	for i := 1; i <= 10; i++ {
		s.Submit(ctx, i)
	}

	s.CloseInput()
	s.Wait()

	stats := s.Stats()
	if stats.AdmittedWeight != 0 {
		t.Errorf("AdmittedWeight = %d, want 0", stats.AdmittedWeight)
	}
	if stats.Admitted != 0 {
		t.Errorf("Admitted = %d, want 0", stats.Admitted)
	}
}

func TestMaxWIPWeightStats(t *testing.T) {
	ctx := context.Background()
	s := toc.Start(ctx, slowFn, toc.Options[int]{
		Capacity:     5,
		Workers:      1,
		MaxWIPWeight: 100,
		Weight:       func(n int) int64 { return int64(n) },
	})

	go func() { for range s.Out() {} }()

	s.Submit(ctx, 30)

	stats := s.Stats()
	if stats.MaxWIPWeight != 100 {
		t.Errorf("MaxWIPWeight = %d, want 100", stats.MaxWIPWeight)
	}
	if stats.AdmittedWeight < 30 {
		t.Errorf("AdmittedWeight = %d, want >= 30", stats.AdmittedWeight)
	}

	s.CloseInput()
	s.Wait()
}

func TestSetMaxWIPCeilingTracksWorkers(t *testing.T) {
	// MaxWIP ceiling should track TargetWorkers, not initial worker count.
	ctx := context.Background()
	s := toc.Start(ctx, identityFn, toc.Options[int]{
		Capacity: 5,
		Workers:  2, // initial: ceiling = 5+2 = 7
	})

	go func() { for range s.Out() {} }()

	// Scale up to 6 workers. New ceiling should be 5+6 = 11.
	s.SetWorkers(6)

	// SetMaxWIP to 10 — should succeed (10 < 11).
	applied := s.SetMaxWIP(10)
	if applied != 10 {
		t.Errorf("SetMaxWIP(10) after SetWorkers(6) = %d, want 10 (ceiling=11)", applied)
	}

	// Before the fix, ceiling was 5+2=7, so SetMaxWIP(10) would clamp to 7.

	s.CloseInput()
	s.Wait()
}
