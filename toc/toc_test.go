package toc_test

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/binaryphile/fluentfp/rslt"
	"github.com/binaryphile/fluentfp/toc"
)

// --- helpers ---

// doubleIt doubles its input.
func doubleIt(_ context.Context, n int) (int, error) { return n * 2, nil }

// drain reads all results from stage.Out().
func drain[T, R any](stage *toc.Stage[T, R]) []rslt.Result[R] {
	var results []rslt.Result[R]
	for r := range stage.Out() {
		results = append(results, r)
	}

	return results
}

// --- Happy path ---

func TestStartAndProcess(t *testing.T) {
	stage := toc.Start(context.Background(), doubleIt, toc.Options[int]{})
	stage.Submit(context.Background(), 5)
	stage.CloseInput()

	results := drain(stage)

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	val, err := results[0].Unpack()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 10 {
		t.Fatalf("got %d, want 10", val)
	}
	if err := stage.Wait(); err != nil {
		t.Fatalf("Wait: %v", err)
	}
}

func TestSerialOrdering(t *testing.T) {
	stage := toc.Start(context.Background(), doubleIt, toc.Options[int]{Capacity: 10})

	for i := 0; i < 10; i++ {
		if err := stage.Submit(context.Background(), i); err != nil {
			t.Fatalf("Submit(%d): %v", i, err)
		}
	}

	stage.CloseInput()

	results := drain(stage)

	if len(results) != 10 {
		t.Fatalf("got %d results, want 10", len(results))
	}
	for i, r := range results {
		val, err := r.Unpack()
		if err != nil {
			t.Fatalf("result[%d]: unexpected error: %v", i, err)
		}
		if val != i*2 {
			t.Fatalf("result[%d]: got %d, want %d", i, val, i*2)
		}
	}

	stage.Wait()
}

func TestMultipleWorkers(t *testing.T) {
	var active atomic.Int32

	// slowFn tracks concurrent execution.
	slowFn := func(_ context.Context, n int) (int, error) {
		active.Add(1)
		defer active.Add(-1)
		time.Sleep(10 * time.Millisecond)

		return n, nil
	}

	stage := toc.Start(context.Background(), slowFn, toc.Options[int]{
		Capacity: 10,
		Workers:  3,
	})

	for i := 0; i < 6; i++ {
		stage.Submit(context.Background(), i)
	}

	stage.CloseInput()
	drain(stage)
	stage.Wait()

	// With 3 workers and 10ms sleep, all 6 items processed.
	stats := stage.Stats()
	if stats.Completed != 6 {
		t.Fatalf("Completed = %d, want 6", stats.Completed)
	}
}

func TestCapacityBackpressure(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})

	// blockFn signals when it starts, then blocks until released.
	blockFn := func(_ context.Context, n int) (int, error) {
		if n == 1 {
			close(started)
		}

		<-release

		return n, nil
	}

	stage := toc.Start(context.Background(), blockFn, toc.Options[int]{Capacity: 1})

	// First submit goes to worker.
	stage.Submit(context.Background(), 1)
	<-started // worker is now blocked in fn

	// Second fills the buffer (capacity 1).
	stage.Submit(context.Background(), 2)

	// Third submit should block because buffer is full and worker is busy.
	submitted := make(chan struct{})
	go func() {
		stage.Submit(context.Background(), 3)
		close(submitted)
	}()

	select {
	case <-submitted:
		t.Fatal("third Submit should block when buffer is full")
	case <-time.After(50 * time.Millisecond):
		// expected: blocked
	}

	// Start draining before releasing — workers need Out() read to unblock.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		drain(stage)
	}()

	close(release)
	<-submitted
	stage.CloseInput()
	wg.Wait()
	stage.Wait()
}

func TestResultCardinality(t *testing.T) {
	n := 100
	stage := toc.Start(context.Background(), doubleIt, toc.Options[int]{Capacity: 10})

	// Drain concurrently — out is unbuffered, so workers block if nobody reads.
	var results []rslt.Result[int]
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		results = drain(stage)
	}()

	for i := 0; i < n; i++ {
		if err := stage.Submit(context.Background(), i); err != nil {
			t.Fatalf("Submit(%d): %v", i, err)
		}
	}

	stage.CloseInput()
	wg.Wait()
	stage.Wait()

	if len(results) != n {
		t.Fatalf("got %d results, want %d", len(results), n)
	}
}

// --- Shutdown ---

func TestSubmitAfterCloseInput(t *testing.T) {
	stage := toc.Start(context.Background(), doubleIt, toc.Options[int]{})
	stage.CloseInput()

	err := stage.Submit(context.Background(), 1)
	if !errors.Is(err, toc.ErrClosed) {
		t.Fatalf("got %v, want ErrClosed", err)
	}

	drain(stage)
	stage.Wait()
}

func TestCloseInputDrainsWorkers(t *testing.T) {
	stage := toc.Start(context.Background(), doubleIt, toc.Options[int]{Capacity: 5})

	for i := 0; i < 5; i++ {
		stage.Submit(context.Background(), i)
	}

	stage.CloseInput()

	results := drain(stage)
	if len(results) != 5 {
		t.Fatalf("got %d results, want 5", len(results))
	}

	stage.Wait()
}

func TestCloseInputIdempotent(t *testing.T) {
	stage := toc.Start(context.Background(), doubleIt, toc.Options[int]{})

	// Should not panic.
	stage.CloseInput()
	stage.CloseInput()
	stage.CloseInput()

	drain(stage)
	stage.Wait()
}

func TestBlockedSubmitUnblocksOnCloseInput(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})

	// blockFn signals start, then blocks until released.
	blockFn := func(_ context.Context, n int) (int, error) {
		if n == 1 {
			close(started)
		}

		<-release

		return n, nil
	}

	stage := toc.Start(context.Background(), blockFn, toc.Options[int]{Capacity: 1})

	// One goes to worker.
	stage.Submit(context.Background(), 1)
	<-started // worker is blocked in fn

	// One fills buffer.
	stage.Submit(context.Background(), 2)

	// This Submit blocks on full buffer.
	errCh := make(chan error, 1)
	go func() {
		errCh <- stage.Submit(context.Background(), 3)
	}()

	time.Sleep(20 * time.Millisecond)
	stage.CloseInput()

	select {
	case err := <-errCh:
		if !errors.Is(err, toc.ErrClosed) {
			t.Fatalf("got %v, want ErrClosed", err)
		}
	case <-time.After(time.Second):
		t.Fatal("blocked Submit did not unblock after CloseInput")
	}

	// Drain concurrently — workers need Out() read to unblock after release.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		drain(stage)
	}()

	close(release)
	wg.Wait()
	stage.Wait()
}

// --- Cancellation ---

func TestBlockedSubmitUnblocksOnParentCancel(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})

	// blockFn signals start, then blocks until released.
	blockFn := func(_ context.Context, n int) (int, error) {
		if n == 1 {
			close(started)
		}

		<-release

		return n, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stage := toc.Start(ctx, blockFn, toc.Options[int]{Capacity: 1})

	// One goes to worker.
	stage.Submit(ctx, 1)
	<-started // worker is blocked in fn

	// One fills buffer.
	stage.Submit(ctx, 2)

	// This Submit blocks on full buffer.
	errCh := make(chan error, 1)
	go func() {
		errCh <- stage.Submit(ctx, 3)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	case <-time.After(time.Second):
		t.Fatal("blocked Submit did not unblock after parent cancel")
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		drain(stage)
	}()

	close(release)
	wg.Wait()
	stage.Wait()
}

func TestSubmitAlreadyCanceledCtx(t *testing.T) {
	stage := toc.Start(context.Background(), doubleIt, toc.Options[int]{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := stage.Submit(ctx, 1)
	if err == nil {
		t.Fatal("expected error for canceled ctx")
	}

	stage.CloseInput()
	drain(stage)
	stage.Wait()
}

func TestParentCtxCancelStopsStage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	stage := toc.Start(ctx, doubleIt, toc.Options[int]{Capacity: 5})

	stage.Submit(ctx, 1)
	cancel()

	// Stage should shut down.
	drain(stage)

	err := stage.Wait()
	if err != nil {
		t.Fatalf("Wait: got %v, want nil", err)
	}
}

func TestParentCancelWithoutCloseInput(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	stage := toc.Start(ctx, doubleIt, toc.Options[int]{})

	stage.Submit(ctx, 1)
	cancel()

	// No explicit CloseInput — cancel watcher handles it.
	drain(stage)
	err := stage.Wait()
	if err != nil {
		t.Fatalf("Wait: got %v, want nil", err)
	}
}

func TestCancelWhileWorkersInFn(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})

	// blockFn signals when it starts, then blocks.
	blockFn := func(_ context.Context, n int) (int, error) {
		close(started)
		<-release

		return n, nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	stage := toc.Start(ctx, blockFn, toc.Options[int]{Capacity: 5})

	stage.Submit(ctx, 1)
	// Submit more to buffer — these will become canceled.
	stage.Submit(ctx, 2)
	stage.Submit(ctx, 3)

	<-started // worker is in fn
	cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		drain(stage)
	}()

	close(release) // let worker finish
	wg.Wait()
	stage.Wait()
}

func TestStartAlreadyCanceledParent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	stage := toc.Start(ctx, doubleIt, toc.Options[int]{Capacity: 5})

	// Many concurrent Submits — all should return error.
	var wg sync.WaitGroup
	var successes atomic.Int32

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			if stage.Submit(ctx, n) == nil {
				successes.Add(1)
			}
		}(i)
	}

	wg.Wait()

	if s := successes.Load(); s > 0 {
		t.Fatalf("got %d successful submits after parent cancel, want 0", s)
	}

	stage.CloseInput()
	drain(stage)
	stage.Wait()
}

func TestPostCancelBurstAdmissionBuffered(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	stage := toc.Start(ctx, doubleIt, toc.Options[int]{Capacity: 10})

	cancel()
	// Wait for cancel to propagate.
	time.Sleep(5 * time.Millisecond)

	var successes atomic.Int32
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			if stage.Submit(ctx, n) == nil {
				successes.Add(1)
			}
		}(i)
	}

	wg.Wait()

	if s := successes.Load(); s > 0 {
		t.Fatalf("got %d successful submits after cancel, want 0", s)
	}

	stage.CloseInput()
	drain(stage)
	stage.Wait()
}

func TestPostCancelBurstAdmissionUnbuffered(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	stage := toc.Start(ctx, doubleIt, toc.Options[int]{Capacity: 0})

	cancel()
	// Wait for cancel to propagate.
	time.Sleep(5 * time.Millisecond)

	var successes atomic.Int32
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			if stage.Submit(ctx, n) == nil {
				successes.Add(1)
			}
		}(i)
	}

	wg.Wait()

	if s := successes.Load(); s > 0 {
		t.Fatalf("got %d successful submits after cancel (unbuffered), want 0", s)
	}

	stage.CloseInput()
	drain(stage)
	stage.Wait()
}

func TestSubmitDuringParentCancelWindow(t *testing.T) {
	for i := 0; i < 100; i++ {
		ctx, cancel := context.WithCancel(context.Background())

		stage := toc.Start(ctx, doubleIt, toc.Options[int]{Capacity: 5})

		go cancel()

		// Should not panic.
		stage.Submit(ctx, 1)
		stage.CloseInput()
		drain(stage)
		stage.Wait()
	}
}

// --- Fail-fast ---

func TestFailFastCancelsRemaining(t *testing.T) {
	errBoom := errors.New("boom")
	var calls atomic.Int32

	// failOnSecond fails on the second call.
	failOnSecond := func(_ context.Context, n int) (int, error) {
		c := calls.Add(1)
		if c == 2 {
			return 0, errBoom
		}
		time.Sleep(5 * time.Millisecond) // let second call start

		return n, nil
	}

	stage := toc.Start(context.Background(), failOnSecond, toc.Options[int]{Capacity: 10})

	for i := 0; i < 5; i++ {
		stage.Submit(context.Background(), i)
	}

	stage.CloseInput()
	drain(stage)

	err := stage.Wait()
	if !errors.Is(err, errBoom) {
		t.Fatalf("Wait: got %v, want %v", err, errBoom)
	}
}

func TestFailFastRejectsSubmit(t *testing.T) {
	errBoom := errors.New("boom")
	started := make(chan struct{})

	// failFn signals start then returns error.
	failFn := func(_ context.Context, n int) (int, error) {
		close(started)

		return 0, errBoom
	}

	stage := toc.Start(context.Background(), failFn, toc.Options[int]{Capacity: 5})

	stage.Submit(context.Background(), 1)
	<-started

	// Give fail-fast time to propagate.
	time.Sleep(20 * time.Millisecond)

	err := stage.Submit(context.Background(), 2)
	if !errors.Is(err, toc.ErrClosed) {
		t.Fatalf("got %v, want ErrClosed", err)
	}

	stage.CloseInput()
	drain(stage)
	stage.Wait()
}

func TestFailFastUnblocksSubmit(t *testing.T) {
	errBoom := errors.New("boom")
	started := make(chan struct{}, 1)

	// failOnSignal returns error after signaling.
	failOnSignal := func(_ context.Context, n int) (int, error) {
		select {
		case started <- struct{}{}:
		default:
		}

		return 0, errBoom
	}

	stage := toc.Start(context.Background(), failOnSignal, toc.Options[int]{Capacity: 1, Workers: 2})

	// Fill buffer and have workers busy.
	stage.Submit(context.Background(), 1)
	stage.Submit(context.Background(), 2)
	stage.Submit(context.Background(), 3)

	<-started // at least one worker started

	// Blocked Submit should unblock from fail-fast.
	errCh := make(chan error, 1)
	go func() {
		errCh <- stage.Submit(context.Background(), 99)
	}()

	select {
	case err := <-errCh:
		if err == nil {
			// Might have been admitted before fail-fast.
		}
	case <-time.After(time.Second):
		t.Fatal("blocked Submit did not unblock after fail-fast")
	}

	stage.CloseInput()
	drain(stage)
	stage.Wait()
}

func TestFailFastWithoutCloseInput(t *testing.T) {
	errBoom := errors.New("boom")

	// alwaysFail always returns an error.
	alwaysFail := func(_ context.Context, n int) (int, error) { return 0, errBoom }

	stage := toc.Start(context.Background(), alwaysFail, toc.Options[int]{})

	stage.Submit(context.Background(), 1)

	// No explicit CloseInput.
	drain(stage)
	err := stage.Wait()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWaitReturnsFirstError(t *testing.T) {
	errBoom := errors.New("boom")

	// alwaysFail always returns errBoom.
	alwaysFail := func(_ context.Context, n int) (int, error) { return 0, errBoom }

	stage := toc.Start(context.Background(), alwaysFail, toc.Options[int]{})

	stage.Submit(context.Background(), 1)
	stage.CloseInput()
	drain(stage)

	err := stage.Wait()
	if !errors.Is(err, errBoom) {
		t.Fatalf("Wait: got %v, want %v", err, errBoom)
	}
}

func TestCanceledResultCarriesFailFastCause(t *testing.T) {
	errBoom := errors.New("boom")
	started := make(chan struct{})

	// failAndSignal returns error after signaling.
	failAndSignal := func(_ context.Context, n int) (int, error) {
		if n == 0 {
			close(started)

			return 0, errBoom
		}
		// Block until canceled.
		time.Sleep(time.Second)

		return n, nil
	}

	stage := toc.Start(context.Background(), failAndSignal, toc.Options[int]{Capacity: 5})

	stage.Submit(context.Background(), 0) // triggers fail-fast
	stage.Submit(context.Background(), 1) // should become canceled

	<-started

	stage.CloseInput()

	var hasBoomCause bool
	for r := range stage.Out() {
		if _, err := r.Unpack(); err != nil && errors.Is(err, errBoom) {
			hasBoomCause = true
		}
	}

	// At least one result (the triggering error or a canceled result)
	// should carry errBoom as its cause.
	if !hasBoomCause {
		t.Fatal("no result carried errBoom cause")
	}

	stage.Wait()
}

func TestInFlightSuccessAfterFailFast(t *testing.T) {
	errBoom := errors.New("boom")
	successStarted := make(chan struct{})
	failStarted := make(chan struct{})
	successRelease := make(chan struct{})

	// fn: item 1 blocks until released (succeeds), item 2 fails immediately.
	fn := func(_ context.Context, n int) (int, error) {
		if n == 1 {
			close(successStarted)
			<-successRelease

			return n * 10, nil // success even after fail-fast
		}

		close(failStarted)

		return 0, errBoom
	}

	stage := toc.Start(context.Background(), fn, toc.Options[int]{Capacity: 5, Workers: 2})

	stage.Submit(context.Background(), 1) // success worker
	stage.Submit(context.Background(), 2) // fail worker

	<-successStarted
	<-failStarted

	// Fail worker has failed, triggering fail-fast.
	// Release success worker to return.
	time.Sleep(10 * time.Millisecond) // let fail-fast propagate
	close(successRelease)

	stage.CloseInput()

	var successes int
	for r := range stage.Out() {
		if _, err := r.Unpack(); err == nil {
			successes++
		}
	}

	if successes != 1 {
		t.Fatalf("got %d in-flight successes, want 1", successes)
	}

	stage.Wait()
}

// --- ContinueOnError ---

func TestContinueOnError(t *testing.T) {
	errBoom := errors.New("boom")
	var calls atomic.Int32

	// failEveryOther fails on odd calls.
	failEveryOther := func(_ context.Context, n int) (int, error) {
		c := calls.Add(1)
		if c%2 == 0 {
			return 0, errBoom
		}

		return n, nil
	}

	stage := toc.Start(context.Background(), failEveryOther, toc.Options[int]{
		Capacity:        10,
		ContinueOnError: true,
	})

	for i := 0; i < 4; i++ {
		stage.Submit(context.Background(), i)
	}

	stage.CloseInput()
	results := drain(stage)
	stage.Wait()

	if len(results) != 4 {
		t.Fatalf("got %d results, want 4 (all items processed)", len(results))
	}
}

func TestWaitReturnsNilOnSuccess(t *testing.T) {
	stage := toc.Start(context.Background(), doubleIt, toc.Options[int]{})

	stage.Submit(context.Background(), 1)
	stage.CloseInput()
	drain(stage)

	if err := stage.Wait(); err != nil {
		t.Fatalf("Wait: got %v, want nil", err)
	}
}

func TestWaitReturnsNilContinueOnError(t *testing.T) {
	errBoom := errors.New("boom")

	// alwaysFail always returns errBoom.
	alwaysFail := func(_ context.Context, n int) (int, error) { return 0, errBoom }

	stage := toc.Start(context.Background(), alwaysFail, toc.Options[int]{
		ContinueOnError: true,
	})

	stage.Submit(context.Background(), 1)
	stage.CloseInput()
	drain(stage)

	if err := stage.Wait(); err != nil {
		t.Fatalf("Wait: got %v, want nil (ContinueOnError)", err)
	}
}

func TestWaitReturnsNilParentCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	stage := toc.Start(ctx, doubleIt, toc.Options[int]{})

	stage.Submit(ctx, 1)
	cancel()

	drain(stage)

	if err := stage.Wait(); err != nil {
		t.Fatalf("Wait: got %v, want nil (parent cancel)", err)
	}
}

func TestContinueOnErrorParentCancel(t *testing.T) {
	errBoom := errors.New("boom")
	started := make(chan struct{})

	// failAndSignal errors then signals.
	failAndSignal := func(ctx context.Context, n int) (int, error) {
		if n == 0 {
			close(started)

			return 0, errBoom
		}

		// Block until canceled — don't use time.Sleep so cancel unblocks promptly.
		<-ctx.Done()

		return n, ctx.Err()
	}

	ctx, cancel := context.WithCancel(context.Background())

	stage := toc.Start(ctx, failAndSignal, toc.Options[int]{
		Capacity:        5,
		ContinueOnError: true,
	})

	stage.Submit(ctx, 0)
	stage.Submit(ctx, 1)
	stage.Submit(ctx, 2)

	<-started
	cancel()

	drain(stage)

	// Wait returns nil in ContinueOnError.
	if err := stage.Wait(); err != nil {
		t.Fatalf("Wait: got %v, want nil", err)
	}

	// Cause returns parent cancel cause.
	cause := stage.Cause()
	if cause == nil {
		t.Fatal("Cause: got nil, want parent cancel cause")
	}
}

// --- Panic ---

func TestPanicRecovery(t *testing.T) {
	// panicFn panics with a string.
	panicFn := func(_ context.Context, n int) (int, error) {
		panic("kaboom")
	}

	stage := toc.Start(context.Background(), panicFn, toc.Options[int]{})

	stage.Submit(context.Background(), 1)
	stage.CloseInput()

	results := drain(stage)

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	_, err := results[0].Unpack()
	if err == nil {
		t.Fatal("expected error from panic")
	}

	var pe *rslt.PanicError
	if !errors.As(err, &pe) {
		t.Fatalf("error is %T, want *rslt.PanicError", err)
	}
	if pe.Value != "kaboom" {
		t.Fatalf("panic value = %v, want %q", pe.Value, "kaboom")
	}
	if len(pe.Stack) == 0 {
		t.Fatal("expected non-empty stack trace")
	}

	stage.Wait()
}

// --- Stats ---

func TestStatsSubmittedCompleted(t *testing.T) {
	stage := toc.Start(context.Background(), doubleIt, toc.Options[int]{Capacity: 10})

	for i := 0; i < 5; i++ {
		stage.Submit(context.Background(), i)
	}

	stage.CloseInput()
	drain(stage)
	stage.Wait()

	stats := stage.Stats()
	if stats.Submitted != 5 {
		t.Fatalf("Submitted = %d, want 5", stats.Submitted)
	}
	if stats.Completed != 5 {
		t.Fatalf("Completed = %d, want 5", stats.Completed)
	}
	if stats.Canceled != 0 {
		t.Fatalf("Canceled = %d, want 0", stats.Canceled)
	}
}

func TestStatsServiceTime(t *testing.T) {
	// slowFn sleeps for 10ms.
	slowFn := func(_ context.Context, n int) (int, error) {
		time.Sleep(10 * time.Millisecond)

		return n, nil
	}

	stage := toc.Start(context.Background(), slowFn, toc.Options[int]{})

	stage.Submit(context.Background(), 1)
	stage.CloseInput()
	drain(stage)
	stage.Wait()

	stats := stage.Stats()
	if stats.ServiceTime < 10*time.Millisecond {
		t.Fatalf("ServiceTime = %v, want >= 10ms", stats.ServiceTime)
	}
}

func TestStatsOutputBlockedTime(t *testing.T) {
	stage := toc.Start(context.Background(), doubleIt, toc.Options[int]{})

	stage.Submit(context.Background(), 1)
	stage.CloseInput()

	// Delay reading to create output-blocked time.
	time.Sleep(30 * time.Millisecond)

	drain(stage)
	stage.Wait()

	stats := stage.Stats()
	if stats.OutputBlockedTime < 20*time.Millisecond {
		t.Fatalf("OutputBlockedTime = %v, want >= 20ms", stats.OutputBlockedTime)
	}
}

func TestStatsOutputBlockedCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	started := make(chan struct{})

	// blockFn blocks until context canceled.
	blockFn := func(ctx context.Context, n int) (int, error) {
		if n == 0 {
			close(started)
		}

		<-ctx.Done()

		return n, ctx.Err()
	}

	stage := toc.Start(ctx, blockFn, toc.Options[int]{
		Capacity:        5,
		ContinueOnError: true,
	})

	stage.Submit(ctx, 0)
	stage.Submit(ctx, 1) // buffered, will be canceled

	<-started
	cancel()

	// Delay reading to create output-blocked time for canceled results.
	time.Sleep(30 * time.Millisecond)

	drain(stage)
	stage.Wait()

	stats := stage.Stats()
	if stats.OutputBlockedTime == 0 {
		t.Fatal("OutputBlockedTime should be > 0 (includes canceled result sends)")
	}
}

func TestStatsBufferedDepth(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})

	// blockFn signals start, then blocks until released.
	blockFn := func(_ context.Context, n int) (int, error) {
		if n == 1 {
			close(started)
		}

		<-release

		return n, nil
	}

	stage := toc.Start(context.Background(), blockFn, toc.Options[int]{Capacity: 5})

	// First goes to worker.
	stage.Submit(context.Background(), 1)
	<-started // worker is blocked in fn

	// These go to buffer.
	stage.Submit(context.Background(), 2)
	stage.Submit(context.Background(), 3)

	stats := stage.Stats()
	if stats.BufferedDepth < 1 {
		t.Fatalf("BufferedDepth = %d, want >= 1", stats.BufferedDepth)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		drain(stage)
	}()

	close(release)
	stage.CloseInput()
	wg.Wait()
	stage.Wait()

	stats = stage.Stats()
	if stats.BufferedDepth != 0 {
		t.Fatalf("final BufferedDepth = %d, want 0", stats.BufferedDepth)
	}
}

func TestStatsInFlightWeight(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})

	// blockFn signals start then blocks.
	blockFn := func(_ context.Context, n int) (int, error) {
		close(started)
		<-release

		return n, nil
	}

	// customWeight returns 10 for every item.
	customWeight := func(_ int) int64 { return 10 }

	stage := toc.Start(context.Background(), blockFn, toc.Options[int]{
		Weight: customWeight,
	})

	stage.Submit(context.Background(), 1)
	<-started

	stats := stage.Stats()
	if stats.InFlightWeight != 10 {
		t.Fatalf("InFlightWeight = %d, want 10", stats.InFlightWeight)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		drain(stage)
	}()

	close(release)
	stage.CloseInput()
	wg.Wait()
	stage.Wait()

	stats = stage.Stats()
	if stats.InFlightWeight != 0 {
		t.Fatalf("final InFlightWeight = %d, want 0", stats.InFlightWeight)
	}
}

// --- Validation ---

func TestNilFnPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for nil fn")
		}
	}()

	toc.Start[int, int](context.Background(), nil, toc.Options[int]{})
}

func TestZeroCapacityUnbuffered(t *testing.T) {
	stage := toc.Start(context.Background(), doubleIt, toc.Options[int]{Capacity: 0})

	// Submit and receive work with unbuffered channel.
	go func() {
		stage.Submit(context.Background(), 5)
		stage.CloseInput()
	}()

	results := drain(stage)
	stage.Wait()

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	val, _ := results[0].Unpack()
	if val != 10 {
		t.Fatalf("got %d, want 10", val)
	}

	stats := stage.Stats()
	if stats.QueueCapacity != 0 {
		t.Fatalf("QueueCapacity = %d, want 0", stats.QueueCapacity)
	}
	if stats.BufferedDepth != 0 {
		t.Fatalf("BufferedDepth = %d, want 0 (clamped for unbuffered)", stats.BufferedDepth)
	}
}

func TestUnbufferedBlockedSubmitCloseInput(t *testing.T) {
	// blockFn blocks forever.
	blockFn := func(_ context.Context, n int) (int, error) {
		select {}
	}

	stage := toc.Start(context.Background(), blockFn, toc.Options[int]{Capacity: 0})

	// First submit goes directly to worker (unbuffered rendezvous).
	go stage.Submit(context.Background(), 1)

	time.Sleep(10 * time.Millisecond) // let first submit rendezvous

	// Second submit blocks because worker is busy and channel is unbuffered.
	errCh := make(chan error, 1)
	go func() {
		errCh <- stage.Submit(context.Background(), 2)
	}()

	time.Sleep(20 * time.Millisecond)
	stage.CloseInput()

	select {
	case err := <-errCh:
		if !errors.Is(err, toc.ErrClosed) {
			t.Fatalf("got %v, want ErrClosed", err)
		}
	case <-time.After(time.Second):
		t.Fatal("blocked Submit did not unblock after CloseInput (unbuffered)")
	}

	// blockFn never returns, but stage is still live. Just verify Submit was unblocked.
	// We can't cleanly shut down since blockFn ignores context.
}

func TestUnbufferedBlockedSubmitFailFast(t *testing.T) {
	errBoom := errors.New("boom")
	started := make(chan struct{})

	// failAndSignal returns error after signaling.
	failAndSignal := func(_ context.Context, n int) (int, error) {
		close(started)

		return 0, errBoom
	}

	stage := toc.Start(context.Background(), failAndSignal, toc.Options[int]{Capacity: 0, Workers: 2})

	// First submit rendezvous with worker 1.
	go stage.Submit(context.Background(), 1)
	<-started // worker 1 has started

	// Second submit should block or get rejected.
	errCh := make(chan error, 1)
	go func() {
		errCh <- stage.Submit(context.Background(), 2)
	}()

	select {
	case err := <-errCh:
		if err == nil {
			// Was admitted to worker 2 before fail-fast; acceptable.
		}
	case <-time.After(time.Second):
		t.Fatal("blocked Submit did not unblock after fail-fast (unbuffered)")
	}

	stage.CloseInput()
	drain(stage)
	stage.Wait()
}

func TestNegativeCapacityPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for negative Capacity")
		}
	}()

	toc.Start[int, int](context.Background(), doubleIt, toc.Options[int]{Capacity: -1})
}

func TestDefaultWorkersOne(t *testing.T) {
	stage := toc.Start(context.Background(), doubleIt, toc.Options[int]{Workers: 0})

	stage.Submit(context.Background(), 1)
	stage.CloseInput()
	drain(stage)
	stage.Wait()

	stats := stage.Stats()
	if stats.Completed != 1 {
		t.Fatalf("Completed = %d, want 1", stats.Completed)
	}
}

func TestNegativeWorkersPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for negative Workers")
		}
	}()

	toc.Start[int, int](context.Background(), doubleIt, toc.Options[int]{Workers: -1})
}

func TestWeightPanicPropagates(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic from Weight")
		}
	}()

	// nilWeight panics with nil pointer dereference.
	var nilPtr *int

	// panicWeight dereferences nil.
	panicWeight := func(_ int) int64 { return int64(*nilPtr) }

	stage := toc.Start(context.Background(), doubleIt, toc.Options[int]{Weight: panicWeight})
	defer stage.CloseInput()

	stage.Submit(context.Background(), 1)
}

func TestNegativeWeightPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for negative weight")
		}
	}()

	// negativeWeight always returns -1.
	negativeWeight := func(_ int) int64 { return -1 }

	stage := toc.Start(context.Background(), doubleIt, toc.Options[int]{Weight: negativeWeight})
	defer stage.CloseInput()

	stage.Submit(context.Background(), 1)
}

func TestNilCtxStartPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for nil ctx")
		}
	}()

	toc.Start[int, int](nil, doubleIt, toc.Options[int]{})
}

func TestNilCtxSubmitPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for nil ctx")
		}
	}()

	stage := toc.Start(context.Background(), doubleIt, toc.Options[int]{})
	defer stage.CloseInput()

	stage.Submit(nil, 1)
}

// --- Cause/Discard ---

func TestCauseReturnsFailFastError(t *testing.T) {
	errBoom := errors.New("boom")

	// alwaysFail always returns errBoom.
	alwaysFail := func(_ context.Context, n int) (int, error) { return 0, errBoom }

	stage := toc.Start(context.Background(), alwaysFail, toc.Options[int]{})

	stage.Submit(context.Background(), 1)
	stage.CloseInput()
	drain(stage)

	cause := stage.Cause()
	if !errors.Is(cause, errBoom) {
		t.Fatalf("Cause: got %v, want %v", cause, errBoom)
	}
}

func TestCauseReturnsParentCancelCause(t *testing.T) {
	errCustom := errors.New("custom cancel")
	ctx, cancel := context.WithCancelCause(context.Background())

	stage := toc.Start(ctx, doubleIt, toc.Options[int]{})

	stage.Submit(ctx, 1)
	cancel(errCustom)

	drain(stage)

	cause := stage.Cause()
	if !errors.Is(cause, errCustom) {
		t.Fatalf("Cause: got %v, want %v", cause, errCustom)
	}
}

func TestCauseReturnsNilOnSuccess(t *testing.T) {
	stage := toc.Start(context.Background(), doubleIt, toc.Options[int]{})

	stage.Submit(context.Background(), 1)
	stage.CloseInput()
	drain(stage)

	if cause := stage.Cause(); cause != nil {
		t.Fatalf("Cause: got %v, want nil", cause)
	}
}

func TestCauseStableAfterLateParentCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	stage := toc.Start(ctx, doubleIt, toc.Options[int]{})

	stage.Submit(ctx, 1)
	stage.CloseInput()
	drain(stage)
	stage.Wait()

	// Cancel AFTER stage is complete.
	cancel()

	cause := stage.Cause()
	if cause != nil {
		t.Fatalf("Cause: got %v, want nil (cancel after completion)", cause)
	}
}

func TestCauseIdempotent(t *testing.T) {
	errBoom := errors.New("boom")

	// alwaysFail always returns errBoom.
	alwaysFail := func(_ context.Context, n int) (int, error) { return 0, errBoom }

	stage := toc.Start(context.Background(), alwaysFail, toc.Options[int]{})

	stage.Submit(context.Background(), 1)
	stage.CloseInput()
	drain(stage)

	cause1 := stage.Cause()
	cause2 := stage.Cause()
	cause3 := stage.Cause()

	if cause1 != cause2 || cause2 != cause3 {
		t.Fatalf("Cause not idempotent: %v, %v, %v", cause1, cause2, cause3)
	}
}

func TestCausePrecedenceWorkerErrorBeatsParentCancel(t *testing.T) {
	errBoom := errors.New("boom")
	started := make(chan struct{})

	// failAndSignal errors after signaling.
	failAndSignal := func(_ context.Context, n int) (int, error) {
		close(started)

		return 0, errBoom
	}

	ctx, cancel := context.WithCancel(context.Background())

	stage := toc.Start(ctx, failAndSignal, toc.Options[int]{})

	stage.Submit(ctx, 1)
	<-started

	// Cancel parent AFTER worker has already failed.
	time.Sleep(10 * time.Millisecond)
	cancel()

	drain(stage)

	cause := stage.Cause()
	if !errors.Is(cause, errBoom) {
		t.Fatalf("Cause: got %v, want %v (worker error should take precedence)", cause, errBoom)
	}
}

func TestDiscardAndWait(t *testing.T) {
	errBoom := errors.New("boom")

	// alwaysFail always returns errBoom.
	alwaysFail := func(_ context.Context, n int) (int, error) { return 0, errBoom }

	stage := toc.Start(context.Background(), alwaysFail, toc.Options[int]{})

	stage.Submit(context.Background(), 1)
	stage.CloseInput()

	err := stage.DiscardAndWait()
	if !errors.Is(err, errBoom) {
		t.Fatalf("DiscardAndWait: got %v, want %v", err, errBoom)
	}
}

func TestDiscardAndCause(t *testing.T) {
	errBoom := errors.New("boom")

	// alwaysFail always returns errBoom.
	alwaysFail := func(_ context.Context, n int) (int, error) { return 0, errBoom }

	stage := toc.Start(context.Background(), alwaysFail, toc.Options[int]{})

	stage.Submit(context.Background(), 1)
	stage.CloseInput()

	cause := stage.DiscardAndCause()
	if !errors.Is(cause, errBoom) {
		t.Fatalf("DiscardAndCause: got %v, want %v", cause, errBoom)
	}
}

func TestCauseParentCancelDuringHandoff(t *testing.T) {
	started := make(chan struct{})

	// succeedAndSignal succeeds after signaling.
	succeedAndSignal := func(_ context.Context, n int) (int, error) {
		close(started)

		return n * 2, nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	stage := toc.Start(ctx, succeedAndSignal, toc.Options[int]{})

	stage.Submit(ctx, 1)
	<-started

	// fn has returned but result not yet consumed.
	// Cancel parent before consumer drains.
	time.Sleep(5 * time.Millisecond)
	cancel()

	// Now drain and check.
	drain(stage)

	// Cause may report parent cancellation if cancel was observed
	// before workers fully exited (result handoff completed).
	// Wait should return nil (no fail-fast error).
	err := stage.Wait()
	if err != nil {
		t.Fatalf("Wait: got %v, want nil", err)
	}

	// Cause may be nil or parent cancel depending on timing.
	// This test documents the behavior, not asserts a specific outcome.
	_ = stage.Cause()
}

// --- Concurrency safety ---

func TestConcurrentSubmitCloseInput(t *testing.T) {
	for i := 0; i < 100; i++ {
		release := make(chan struct{})

		// blockFn blocks until released, creating backpressure so
		// submitters are genuinely blocked when CloseInput fires.
		blockFn := func(_ context.Context, n int) (int, error) {
			<-release

			return n, nil
		}

		stage := toc.Start(context.Background(), blockFn, toc.Options[int]{Capacity: 1})

		// Launch submitters that will block on full buffer.
		var submitWg sync.WaitGroup

		for j := 0; j < 10; j++ {
			submitWg.Add(1)
			go func() {
				defer submitWg.Done()
				stage.Submit(context.Background(), 1)
			}()
		}

		// Let submitters hit the blocking select, then close concurrently.
		runtime.Gosched()

		go stage.CloseInput()

		// Drain concurrently — workers need Out() read to unblock.
		var drainWg sync.WaitGroup
		drainWg.Add(1)
		go func() {
			defer drainWg.Done()
			drain(stage)
		}()

		close(release)
		submitWg.Wait()
		drainWg.Wait()
		stage.Wait()
	}
}

func TestFailFastConcurrentSubmit(t *testing.T) {
	errBoom := errors.New("boom")

	for i := 0; i < 100; i++ {
		release := make(chan struct{})

		// blockThenFail: first call blocks then fails, rest block until released.
		var calls atomic.Int32

		blockThenFail := func(_ context.Context, n int) (int, error) {
			c := calls.Add(1)
			if c == 1 {
				<-release

				return 0, errBoom
			}

			<-release

			return n, nil
		}

		stage := toc.Start(context.Background(), blockThenFail, toc.Options[int]{Capacity: 1, Workers: 2})

		// Launch submitters that will block on full buffer.
		var submitWg sync.WaitGroup

		for j := 0; j < 10; j++ {
			submitWg.Add(1)
			go func() {
				defer submitWg.Done()
				stage.Submit(context.Background(), 1)
			}()
		}

		// Let submitters block, then trigger fail-fast.
		runtime.Gosched()

		var drainWg sync.WaitGroup
		drainWg.Add(1)
		go func() {
			defer drainWg.Done()
			drain(stage)
		}()

		close(release)
		submitWg.Wait()
		stage.CloseInput()
		drainWg.Wait()
		stage.Wait()
	}
}

func TestNormalShutdownNoGoroutineLeak(t *testing.T) {
	before := runtime.NumGoroutine()

	stage := toc.Start(context.Background(), doubleIt, toc.Options[int]{Capacity: 5, Workers: 3})

	// Drain concurrently — 10 items exceeds cap+workers buffer.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		drain(stage)
	}()

	for i := 0; i < 10; i++ {
		stage.Submit(context.Background(), i)
	}

	stage.CloseInput()
	wg.Wait()
	stage.Wait()

	// Give goroutines time to exit.
	time.Sleep(50 * time.Millisecond)

	after := runtime.NumGoroutine()
	// Allow some slack for runtime goroutines.
	if after > before+2 {
		t.Fatalf("goroutine leak: before=%d, after=%d", before, after)
	}
}

func TestConcurrentWait(t *testing.T) {
	errBoom := errors.New("boom")

	// alwaysFail always returns errBoom.
	alwaysFail := func(_ context.Context, n int) (int, error) { return 0, errBoom }

	stage := toc.Start(context.Background(), alwaysFail, toc.Options[int]{})

	stage.Submit(context.Background(), 1)
	stage.CloseInput()

	go drain(stage)

	var wg sync.WaitGroup
	errs := make([]error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			errs[idx] = stage.Wait()
		}(i)
	}

	wg.Wait()

	for i, err := range errs {
		if !errors.Is(err, errBoom) {
			t.Fatalf("Wait[%d]: got %v, want %v", i, err, errBoom)
		}
	}
}

func TestConcurrentWaitAndCause(t *testing.T) {
	errBoom := errors.New("boom")

	// alwaysFail always returns errBoom.
	alwaysFail := func(_ context.Context, n int) (int, error) { return 0, errBoom }

	stage := toc.Start(context.Background(), alwaysFail, toc.Options[int]{})

	stage.Submit(context.Background(), 1)
	stage.CloseInput()

	go drain(stage)

	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			stage.Wait()
		}()
		go func() {
			defer wg.Done()
			stage.Cause()
		}()
	}

	wg.Wait()
}

func TestParentCancelRacesWorkerError(t *testing.T) {
	errBoom := errors.New("boom")

	for i := 0; i < 100; i++ {
		ctx, cancel := context.WithCancel(context.Background())

		// alwaysFail always returns errBoom.
		alwaysFail := func(_ context.Context, n int) (int, error) { return 0, errBoom }

		stage := toc.Start(ctx, alwaysFail, toc.Options[int]{})

		stage.Submit(ctx, 1)
		go cancel()

		stage.CloseInput()
		drain(stage)

		// Either nil or errBoom is acceptable.
		err := stage.Wait()
		_ = err
		cause := stage.Cause()
		_ = cause
	}
}

func TestFinalStatsInvariants(t *testing.T) {
	errBoom := errors.New("boom")

	// failEveryThird fails on every 3rd call.
	var calls atomic.Int32

	failEveryThird := func(_ context.Context, n int) (int, error) {
		if calls.Add(1)%3 == 0 {
			return 0, errBoom
		}

		return n, nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	stage := toc.Start(ctx, failEveryThird, toc.Options[int]{
		Capacity:        10,
		ContinueOnError: true,
	})

	for i := 0; i < 10; i++ {
		stage.Submit(ctx, i)
	}

	cancel()
	stage.CloseInput()
	drain(stage)
	stage.Wait()

	stats := stage.Stats()

	// Submitted == Completed + Canceled
	if stats.Submitted != stats.Completed+stats.Canceled {
		t.Fatalf("Submitted(%d) != Completed(%d) + Canceled(%d)",
			stats.Submitted, stats.Completed, stats.Canceled)
	}

	if stats.BufferedDepth != 0 {
		t.Fatalf("final BufferedDepth = %d, want 0", stats.BufferedDepth)
	}

	if stats.InFlightWeight != 0 {
		t.Fatalf("final InFlightWeight = %d, want 0", stats.InFlightWeight)
	}
}

func TestCardinalityProperty(t *testing.T) {
	// Property: successful submits == len(results), across random interleavings.
	for i := 0; i < 50; i++ {
		stage := toc.Start(context.Background(), doubleIt, toc.Options[int]{
			Capacity: 3,
			Workers:  2,
		})

		var results []rslt.Result[int]
		var drainWg sync.WaitGroup
		drainWg.Add(1)
		go func() {
			defer drainWg.Done()
			results = drain(stage)
		}()

		var successes atomic.Int32
		var submitWg sync.WaitGroup
		n := 20

		for j := 0; j < n; j++ {
			submitWg.Add(1)
			go func(v int) {
				defer submitWg.Done()
				if stage.Submit(context.Background(), v) == nil {
					successes.Add(1)
				}
			}(j)
		}

		submitWg.Wait()
		stage.CloseInput()
		drainWg.Wait()
		stage.Wait()

		if int(successes.Load()) != len(results) {
			t.Fatalf("iteration %d: %d successful submits but %d results",
				i, successes.Load(), len(results))
		}
	}
}

func TestUndrainedOutBlocksWait(t *testing.T) {
	stage := toc.Start(context.Background(), doubleIt, toc.Options[int]{})

	stage.Submit(context.Background(), 1)
	stage.CloseInput()

	// Do NOT drain Out. Wait should not return.
	done := make(chan struct{})
	go func() {
		stage.Wait()
		close(done)
	}()

	select {
	case <-done:
		t.Fatal("Wait returned without draining Out")
	case <-time.After(100 * time.Millisecond):
		// Expected: Wait is blocked.
	}

	// Clean up: drain so the test doesn't leak.
	drain(stage)
	<-done
}
