package toc

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

const testTimeout = 5 * time.Second

func recv(t *testing.T, ch <-chan struct{}, msg string) {
	t.Helper()

	select {
	case <-ch:
	case <-time.After(testTimeout):
		t.Fatalf("timeout: %s", msg)
	}
}

func recvErr(t *testing.T, ch <-chan error, msg string) error {
	t.Helper()

	select {
	case err := <-ch:
		return err
	case <-time.After(testTimeout):
		t.Fatalf("timeout: %s", msg)

		return nil
	}
}

// blockingFn returns a fn that blocks until gate is closed.
func blockingFn(gate <-chan struct{}) func(context.Context, int) (int, error) {
	return func(ctx context.Context, n int) (int, error) {
		select {
		case <-gate:
			return n * 10, nil
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	}
}

func TestInterleave_CloseWinsBeforeGrant(t *testing.T) {
	// Prove: closedForAdmission prevents grantWaitersLocked from granting.
	// Waiter wakes via closing channel and returns ErrClosed.
	workerGate := make(chan struct{})
	ctx := context.Background()

	s := Start(ctx, blockingFn(workerGate), Options[int]{Capacity: 5, Workers: 1, MaxWIP: 1})

	// Hook: know when B is queued as waiter.
	waiterQueued := make(chan struct{}, 1)
	// Hook: pause CloseInput between closedForAdmission and close(s.closing).
	closeReached := make(chan struct{})
	closeProceed := make(chan struct{})
	// Hook: know when releaseAdmission completes.
	released := make(chan struct{}, 1)

	s.hooks = &testHooks{
		afterWaiterQueued:   func() { notifyNonBlocking(waiterQueued) },
		afterCloseAdmission: func() { close(closeReached); <-closeProceed },
		afterRelease:        func() { notifyNonBlocking(released) },
	}

	// Drain output in background so workers don't block on unbuffered send.
	go func() { for range s.Out() {} }()

	// A: fast path, worker blocks.
	s.Submit(ctx, 1)

	// B: will queue as waiter.
	errCh := make(chan error, 1)
	go func() { errCh <- s.Submit(ctx, 2) }()
	recv(t, waiterQueued, "waiter B queued")

	// CloseInput: sets closedForAdmission, pauses before close(s.closing).
	go s.CloseInput()
	recv(t, closeReached, "closedForAdmission set")

	// Release worker — releaseAdmission → grantWaitersLocked bails (closedForAdmission=true).
	close(workerGate)
	recv(t, released, "releaseAdmission completed")

	// Let CloseInput proceed → close(s.closing) wakes waiter B.
	close(closeProceed)

	err := recvErr(t, errCh, "waiter B returns")
	if err != ErrClosed {
		t.Errorf("waiter B: got %v, want ErrClosed", err)
	}

	s.Wait()

	stats := s.Stats()
	if stats.Admitted != 0 {
		t.Errorf("Admitted = %d, want 0", stats.Admitted)
	}
	if stats.WaiterCount != 0 {
		t.Errorf("WaiterCount = %d, want 0", stats.WaiterCount)
	}
}

func TestInterleave_GrantWinsBeforeClose(t *testing.T) {
	// Prove: granted waiter + subsequent close → no permit leak.
	workerGate := make(chan struct{})
	ctx := context.Background()

	s := Start(ctx, blockingFn(workerGate), Options[int]{Capacity: 5, Workers: 1, MaxWIP: 1})

	waiterQueued := make(chan struct{}, 1)
	granted := make(chan struct{}, 1)

	s.hooks = &testHooks{
		afterWaiterQueued: func() { notifyNonBlocking(waiterQueued) },
		onGrant:           func() { notifyNonBlocking(granted) },
	}

	go func() { for range s.Out() {} }()

	// A: fast path, worker blocks.
	s.Submit(ctx, 1)

	// B: queues as waiter.
	errCh := make(chan error, 1)
	go func() { errCh <- s.Submit(ctx, 2) }()
	recv(t, waiterQueued, "waiter B queued")

	// Release worker → grant fires BEFORE close.
	close(workerGate)
	recv(t, granted, "waiter B granted")

	// Close AFTER grant.
	s.CloseInput()

	err := recvErr(t, errCh, "waiter B returns")
	if err != nil && err != ErrClosed {
		t.Errorf("waiter B: unexpected error %v", err)
	}

	s.Wait()

	stats := s.Stats()
	if stats.Admitted != 0 {
		t.Errorf("Admitted = %d, want 0", stats.Admitted)
	}
}

func TestInterleave_FastPathVsClose(t *testing.T) {
	passthrough := func(_ context.Context, n int) (int, error) { return n, nil }

	t.Run("admit_then_close", func(t *testing.T) {
		ctx := context.Background()
		s := Start(ctx, passthrough, Options[int]{Capacity: 5, Workers: 1, MaxWIP: 5})

		err := s.Submit(ctx, 1)
		if err != nil {
			t.Fatalf("Submit: %v", err)
		}

		s.CloseInput()

		for range s.Out() {
		}

		s.Wait()

		if s.Stats().Admitted != 0 {
			t.Errorf("Admitted = %d, want 0", s.Stats().Admitted)
		}
	})

	t.Run("close_then_admit", func(t *testing.T) {
		ctx := context.Background()
		s := Start(ctx, passthrough, Options[int]{Capacity: 5, Workers: 1, MaxWIP: 5})

		s.CloseInput()

		err := s.Submit(ctx, 1)
		if err != ErrClosed {
			t.Errorf("Submit: got %v, want ErrClosed", err)
		}

		for range s.Out() {
		}

		s.Wait()

		if s.Stats().Admitted != 0 {
			t.Errorf("Admitted = %d, want 0", s.Stats().Admitted)
		}
	})
}

func TestInterleave_HonorBranch(t *testing.T) {
	// Deterministically prove revokeOrHonor honor path:
	// ctx cancels → waiter enters revokeOrHonor → pauses before lock →
	// grant happens under lock → revokeOrHonor resumes → finds waiter
	// absent → returns nil (honors grant).
	workerGate := make(chan struct{})
	ctx := context.Background()
	cancelCtx, cancel := context.WithCancel(ctx)

	s := Start(ctx, blockingFn(workerGate), Options[int]{Capacity: 5, Workers: 1, MaxWIP: 1})

	waiterQueued := make(chan struct{}, 1)
	revokeReached := make(chan struct{})
	revokeProceed := make(chan struct{})
	granted := make(chan struct{}, 1)

	s.hooks = &testHooks{
		afterWaiterQueued: func() { notifyNonBlocking(waiterQueued) },
		beforeRevokeOrHonor: func() {
			// Signal that revokeOrHonor is entered, wait for test to proceed.
			close(revokeReached)
			<-revokeProceed
		},
		onGrant: func() { notifyNonBlocking(granted) },
	}

	go func() { for range s.Out() {} }()

	// A: fast path, worker blocks.
	s.Submit(ctx, 1)

	// B: queues as waiter with cancellable ctx.
	errCh := make(chan error, 1)
	go func() { errCh <- s.Submit(cancelCtx, 2) }()
	recv(t, waiterQueued, "waiter B queued")

	// Cancel B's ctx → waiter's select picks ctx.Done() → enters revokeOrHonor.
	cancel()
	recv(t, revokeReached, "revokeOrHonor entered")

	// While revokeOrHonor is paused, release worker → grant fires.
	close(workerGate)
	recv(t, granted, "waiter B granted under lock")

	// Resume revokeOrHonor → finds waiter absent → honors grant.
	close(revokeProceed)

	err := recvErr(t, errCh, "waiter B returns")
	// Honor path: revokeOrHonor returns nil, Submit proceeds to trySend.
	// trySend sees cancelled ctx and returns ctx.Err(), Submit rolls back.
	// The key proof: no permit leak despite grant + cancel race.
	// err may be context.Canceled (trySend ctx check) or nil (if trySend
	// succeeds before ctx propagates) or ErrClosed (if close races).
	_ = err

	s.CloseInput()

	for range s.Out() {
	}

	s.Wait()

	if s.Stats().Admitted != 0 {
		t.Errorf("Admitted = %d, want 0", s.Stats().Admitted)
	}
}

func TestInterleave_WaiterQueueDrainsOnClose(t *testing.T) {
	// Prove: N waiters all return ErrClosed, queue empties, admitted=0.
	const N = 5
	workerGate := make(chan struct{})
	ctx := context.Background()

	s := Start(ctx, blockingFn(workerGate), Options[int]{Capacity: 5, Workers: 1, MaxWIP: 1})

	waiterCount := make(chan struct{}, N)
	closeReached := make(chan struct{})
	closeProceed := make(chan struct{})
	released := make(chan struct{}, 1)

	s.hooks = &testHooks{
		afterWaiterQueued:   func() { notifyNonBlocking(waiterCount) },
		afterCloseAdmission: func() { close(closeReached); <-closeProceed },
		afterRelease:        func() { notifyNonBlocking(released) },
	}

	go func() { for range s.Out() {} }()

	// A: fast path, worker blocks.
	s.Submit(ctx, 1)

	// N waiters.
	errs := make(chan error, N)
	for i := range N {
		go func() { errs <- s.Submit(ctx, i+10) }()
	}
	for range N {
		recv(t, waiterCount, "waiter queued")
	}

	// CloseInput: pause after closedForAdmission.
	go s.CloseInput()
	recv(t, closeReached, "closedForAdmission set")

	// Release worker — grantWaitersLocked bails.
	close(workerGate)
	recv(t, released, "releaseAdmission completed")

	// Let CloseInput proceed — wakes all waiters.
	close(closeProceed)

	for range N {
		err := recvErr(t, errs, "waiter returns")
		if err != ErrClosed {
			t.Errorf("waiter: got %v, want ErrClosed", err)
		}
	}

	s.Wait()

	stats := s.Stats()
	if stats.Admitted != 0 {
		t.Errorf("Admitted = %d, want 0", stats.Admitted)
	}
	if stats.WaiterCount != 0 {
		t.Errorf("WaiterCount = %d, want 0", stats.WaiterCount)
	}
}

func TestInterleave_StageDoneUnblocksWaiter(t *testing.T) {
	// Prove: stageDone (parent cancel) wakes a blocked waiter deterministically.
	workerGate := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())

	s := Start(ctx, blockingFn(workerGate), Options[int]{Capacity: 5, Workers: 1, MaxWIP: 1})

	waiterQueued := make(chan struct{}, 1)

	s.hooks = &testHooks{
		afterWaiterQueued: func() { notifyNonBlocking(waiterQueued) },
	}

	go func() { for range s.Out() {} }()

	// A: fast path, worker blocks.
	s.Submit(ctx, 1)

	// B: queues as waiter.
	errCh := make(chan error, 1)
	go func() { errCh <- s.Submit(ctx, 2) }()
	recv(t, waiterQueued, "waiter B queued")

	// Cancel parent context → stageDone fires → waiter wakes.
	cancel()

	err := recvErr(t, errCh, "waiter B returns")
	if err == nil {
		t.Fatal("expected error after parent cancel")
	}

	// Release worker so stage can shut down.
	close(workerGate)

	s.Wait()

	if s.Stats().Admitted != 0 {
		t.Errorf("Admitted = %d, want 0", s.Stats().Admitted)
	}
}

// BenchmarkAcquireAdmissionFastPath measures the fast-path overhead of
// admission control, including the nil hook check.
func BenchmarkAcquireAdmissionFastPath(b *testing.B) {
	var sink atomic.Int64

	for _, withHooks := range []bool{false, true} {
		name := "no_hooks"
		if withHooks {
			name = "with_hooks"
		}

		b.Run(name, func(b *testing.B) {
			ctx := context.Background()
			fn := func(_ context.Context, n int) (int, error) { return n, nil }
			s := Start(ctx, fn, Options[int]{Capacity: b.N, Workers: 1, MaxWIP: b.N + 1})

			if withHooks {
				s.hooks = &testHooks{} // all nil funcs — measures pointer + nil check overhead
			}

			go func() { for range s.Out() {} }()

			b.ResetTimer()
			for i := range b.N {
				s.Submit(ctx, i)
			}
			b.StopTimer()

			s.CloseInput()
			s.Wait()
			sink.Store(int64(b.N))
		})
	}

	_ = sink.Load()
}
