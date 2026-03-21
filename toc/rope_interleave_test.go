package toc

import (
	"context"
	"sync"
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

// countDown fires after exactly n calls. Panics on overfire.
type countDown struct {
	remaining atomic.Int32
	done      chan struct{}
	once      sync.Once
}

func newCountDown(n int) *countDown {
	c := &countDown{done: make(chan struct{})}
	c.remaining.Store(int32(n))
	return c
}

func (c *countDown) Fire() {
	v := c.remaining.Add(-1)
	switch {
	case v == 0:
		c.once.Do(func() { close(c.done) })
	case v < 0:
		panic("countDown fired too many times")
	}
}

func (c *countDown) C() <-chan struct{} { return c.done }

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

	waiterQueued := newOneShot()
	closeReached := make(chan struct{})
	closeProceed := make(chan struct{})
	released := newOneShot()

	s.hooks.Store(&testHooks{
		afterWaiterQueued:   waiterQueued.Fire,
		afterCloseAdmission: func() { close(closeReached); <-closeProceed },
		afterRelease:        released.Fire,
	})

	go func() { for range s.Out() {} }()

	s.Submit(ctx, 1) // A: fast path, worker blocks

	errCh := make(chan error, 1)
	go func() { errCh <- s.Submit(ctx, 2) }() // B: waiter
	recv(t, waiterQueued.C(), "waiter B queued")

	go s.CloseInput()
	recv(t, closeReached, "closedForAdmission set")

	close(workerGate) // release worker → grantWaitersLocked bails
	recv(t, released.C(), "releaseAdmission completed")

	close(closeProceed) // close(s.closing) wakes B

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

	waiterQueued := newOneShot()
	granted := newOneShot()

	s.hooks.Store(&testHooks{
		afterWaiterQueued: waiterQueued.Fire,
		onGrant:           granted.Fire, // one-shot, non-blocking, safe under lock
	})

	go func() { for range s.Out() {} }()

	s.Submit(ctx, 1) // A: fast path, worker blocks

	errCh := make(chan error, 1)
	go func() { errCh <- s.Submit(ctx, 2) }() // B: waiter
	recv(t, waiterQueued.C(), "waiter B queued")

	close(workerGate) // release → grant fires
	recv(t, granted.C(), "waiter B granted")

	s.CloseInput() // close AFTER grant

	err := recvErr(t, errCh, "waiter B returns")
	// Grant linearized before close. trySend may succeed (nil) or see
	// close (ErrClosed). Both are valid — grant preceded close.
	if err != nil && err != ErrClosed {
		t.Errorf("waiter B: unexpected error %v", err)
	}

	s.Wait()

	stats := s.Stats()
	if stats.Admitted != 0 {
		t.Errorf("Admitted = %d, want 0", stats.Admitted)
	}
	// If Submit succeeded, item B was processed (Submitted incremented).
	// If Submit returned ErrClosed, item B was not processed.
	// Either way, exactly one outcome — no duplicate processing.
	if err == nil && stats.Submitted < 2 {
		t.Errorf("Submit succeeded but Submitted = %d, want >= 2", stats.Submitted)
	}
	if err == ErrClosed && stats.Submitted > 1 {
		t.Errorf("Submit returned ErrClosed but Submitted = %d, want 1", stats.Submitted)
	}
}

func TestInterleave_FastPathVsClose(t *testing.T) {
	passthrough := func(_ context.Context, n int) (int, error) { return n, nil }

	t.Run("admit_then_close", func(t *testing.T) {
		ctx := context.Background()
		s := Start(ctx, passthrough, Options[int]{Capacity: 5, Workers: 1, MaxWIP: 5})

		if err := s.Submit(ctx, 1); err != nil {
			t.Fatalf("Submit: %v", err)
		}

		s.CloseInput()
		for range s.Out() {}
		s.Wait()

		if s.Stats().Admitted != 0 {
			t.Errorf("Admitted = %d, want 0", s.Stats().Admitted)
		}
	})

	t.Run("close_then_admit", func(t *testing.T) {
		ctx := context.Background()
		s := Start(ctx, passthrough, Options[int]{Capacity: 5, Workers: 1, MaxWIP: 5})

		s.CloseInput()

		if err := s.Submit(ctx, 1); err != ErrClosed {
			t.Errorf("Submit: got %v, want ErrClosed", err)
		}

		for range s.Out() {}
		s.Wait()

		if s.Stats().Admitted != 0 {
			t.Errorf("Admitted = %d, want 0", s.Stats().Admitted)
		}
	})
}

func TestInterleave_HonorBranch(t *testing.T) {
	// Deterministically prove revokeOrHonor honor path:
	// ctx cancels → revokeOrHonor pauses before lock → grant happens →
	// revokeOrHonor resumes → finds waiter absent → returns nil.
	workerGate := make(chan struct{})
	ctx := context.Background()
	cancelCtx, cancel := context.WithCancel(ctx)

	s := Start(ctx, blockingFn(workerGate), Options[int]{Capacity: 5, Workers: 1, MaxWIP: 1})

	waiterQueued := newOneShot()
	revokeReached := make(chan struct{})
	revokeProceed := make(chan struct{})
	granted := newOneShot()

	s.hooks.Store(&testHooks{
		afterWaiterQueued: waiterQueued.Fire,
		beforeRevokeOrHonor: func() {
			close(revokeReached)
			<-revokeProceed
		},
		onGrant: granted.Fire,
	})

	go func() { for range s.Out() {} }()

	s.Submit(ctx, 1) // A: fast path, worker blocks

	errCh := make(chan error, 1)
	go func() { errCh <- s.Submit(cancelCtx, 2) }() // B: cancellable waiter
	recv(t, waiterQueued.C(), "waiter B queued")

	cancel() // B's select picks ctx.Done() → enters revokeOrHonor
	recv(t, revokeReached, "revokeOrHonor entered, paused before lock")

	close(workerGate) // grant fires under lock while revokeOrHonor paused
	recv(t, granted.C(), "waiter B granted under lock")

	close(revokeProceed) // revokeOrHonor resumes → honor path

	_ = recvErr(t, errCh, "waiter B returns")
	// Honor path exercised. trySend sees cancelled ctx → rolls back.
	// No permit leak is the proof.

	s.CloseInput()
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

	waitersQueued := newCountDown(N) // panics on overfire
	closeReached := make(chan struct{})
	closeProceed := make(chan struct{})
	released := newOneShot()

	s.hooks.Store(&testHooks{
		afterWaiterQueued:   waitersQueued.Fire,
		afterCloseAdmission: func() { close(closeReached); <-closeProceed },
		afterRelease:        released.Fire,
	})

	go func() { for range s.Out() {} }()

	s.Submit(ctx, 1) // A: fast path, worker blocks

	errs := make(chan error, N)
	for i := range N {
		go func() { errs <- s.Submit(ctx, i+10) }()
	}
	recv(t, waitersQueued.C(), "all N waiters queued")

	go s.CloseInput()
	recv(t, closeReached, "closedForAdmission set")

	close(workerGate)
	recv(t, released.C(), "releaseAdmission completed")

	close(closeProceed) // wakes all waiters

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
	// Prove: stageDone (parent cancel) wakes a blocked waiter.
	workerGate := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())

	s := Start(ctx, blockingFn(workerGate), Options[int]{Capacity: 5, Workers: 1, MaxWIP: 1})

	waiterQueued := newOneShot()

	s.hooks.Store(&testHooks{
		afterWaiterQueued: waiterQueued.Fire,
	})

	go func() { for range s.Out() {} }()

	s.Submit(ctx, 1) // A: fast path, worker blocks

	errCh := make(chan error, 1)
	go func() { errCh <- s.Submit(ctx, 2) }() // B: waiter
	recv(t, waiterQueued.C(), "waiter B queued")

	cancel() // stageDone fires → waiter wakes

	err := recvErr(t, errCh, "waiter B returns")
	// Parent cancel triggers multiple events: ctx.Done(), stageDone,
	// and (via cancel watcher) CloseInput → closing. The worker may also
	// complete and release, granting B before closedForAdmission is set.
	// Valid outcomes:
	//   nil              — B granted + trySend succeeded before close
	//   ErrClosed        — B rejected via closing/stageDone/closedForAdmission
	//   context.Canceled — B's ctx.Done() won the select + trySend saw ctx
	if err != nil && err != ErrClosed && err != context.Canceled {
		t.Errorf("waiter B: got %v, want nil, ErrClosed, or context.Canceled", err)
	}

	close(workerGate)
	s.Wait()

	if s.Stats().Admitted != 0 {
		t.Errorf("Admitted = %d, want 0", s.Stats().Admitted)
	}
}

// BenchmarkAcquireAdmissionFastPath measures fast-path overhead including hook check.
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
				s.hooks.Store(&testHooks{}) // all nil funcs — measures atomic.Load + nil check
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
