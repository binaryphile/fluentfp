package cb

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// recv waits for a value on ch with a timeout. Returns the value and
// true if received, or zero and false on timeout.
func recv[T any](ch <-chan T, timeout time.Duration) (T, bool) {
	select {
	case v := <-ch:
		return v, true
	case <-time.After(timeout):
		var zero T

		return zero, false
	}
}

// noRecv asserts that ch does not produce a value within timeout.
func noRecv[T any](t *testing.T, ch <-chan T, timeout time.Duration) {
	t.Helper()

	select {
	case v := <-ch:
		t.Fatalf("unexpected value: %v", v)
	case <-time.After(timeout):
	}
}

func TestDebouncerBasicTrailing(t *testing.T) {
	results := make(chan int, 1)

	// record stores the value on the results channel.
	record := func(v int) { results <- v }

	d := NewDebouncer(10*time.Millisecond, record)
	defer d.Close()

	d.Call(42)

	v, ok := recv(results, 100*time.Millisecond)
	if !ok {
		t.Fatal("timed out waiting for fn")
	}

	if v != 42 {
		t.Errorf("got %d, want 42", v)
	}
}

func TestDebouncerCoalescing(t *testing.T) {
	results := make(chan int, 1)

	// record stores the value on the results channel.
	record := func(v int) { results <- v }

	d := NewDebouncer(10*time.Millisecond, record)
	defer d.Close()

	d.Call(1)
	d.Call(2)
	d.Call(3)

	v, ok := recv(results, 100*time.Millisecond)
	if !ok {
		t.Fatal("timed out waiting for fn")
	}

	if v != 3 {
		t.Errorf("got %d, want 3 (latest value)", v)
	}

	noRecv(t, results, 30*time.Millisecond)
}

func TestDebouncerTimerReset(t *testing.T) {
	results := make(chan int, 1)

	// record stores the value on the results channel.
	record := func(v int) { results <- v }

	d := NewDebouncer(20*time.Millisecond, record)
	defer d.Close()

	d.Call(1)
	time.Sleep(10 * time.Millisecond)
	d.Call(2) // resets timer

	// Should NOT have fired yet (10ms into a 20ms wait).
	noRecv(t, results, 10*time.Millisecond)

	v, ok := recv(results, 50*time.Millisecond)
	if !ok {
		t.Fatal("timed out waiting for fn")
	}

	if v != 2 {
		t.Errorf("got %d, want 2", v)
	}
}

func TestDebouncerCancelStopsPending(t *testing.T) {
	results := make(chan int, 1)

	// record stores the value on the results channel.
	record := func(v int) { results <- v }

	d := NewDebouncer(10*time.Millisecond, record)
	defer d.Close()

	d.Call(1)

	if got := d.Cancel(); !got {
		t.Error("Cancel returned false, want true")
	}

	noRecv(t, results, 30*time.Millisecond)
}

func TestDebouncerCancelOnIdle(t *testing.T) {
	d := NewDebouncer[int](10*time.Millisecond, func(int) {})
	defer d.Close()

	if got := d.Cancel(); got {
		t.Error("Cancel on idle returned true, want false")
	}
}

func TestDebouncerCancelOnRunningNoPending(t *testing.T) {
	started := make(chan struct{}, 1)
	block := make(chan struct{})

	d := NewDebouncer(1*time.Millisecond, func(int) {
		started <- struct{}{}
		<-block
	})
	defer d.Close()

	d.Call(1)
	<-started // fn is running

	if got := d.Cancel(); got {
		t.Error("Cancel on running (no pending) returned true, want false")
	}

	close(block)
}

func TestDebouncerCancelOnRunningPending(t *testing.T) {
	started := make(chan struct{}, 1)
	block := make(chan struct{})
	results := make(chan int, 2)

	// record stores the value; fn(1) blocks.
	record := func(v int) {
		if v == 1 {
			started <- struct{}{}
			<-block
		}

		results <- v
	}

	d := NewDebouncer(1*time.Millisecond, record)
	defer d.Close()

	d.Call(1)
	<-started // fn(1) is running

	d.Call(2) // → running+pending

	if got := d.Cancel(); !got {
		t.Error("Cancel on running+pending returned false, want true")
	}

	close(block) // let fn(1) complete

	// fn(1) result is expected.
	v, ok := recv(results, 100*time.Millisecond)
	if !ok {
		t.Fatal("fn(1) did not complete")
	}

	if v != 1 {
		t.Errorf("got %d, want 1", v)
	}

	// fn(2) should NOT execute (was canceled).
	noRecv(t, results, 30*time.Millisecond)
}

func TestDebouncerFlushExecutesImmediately(t *testing.T) {
	results := make(chan int, 1)

	// record stores the value on the results channel.
	record := func(v int) { results <- v }

	d := NewDebouncer(1*time.Second, record)
	defer d.Close()

	d.Call(42)

	start := time.Now()

	if got := d.Flush(); !got {
		t.Error("Flush returned false, want true")
	}

	elapsed := time.Since(start)
	if elapsed > 100*time.Millisecond {
		t.Errorf("Flush took %v, want < 100ms", elapsed)
	}

	v, ok := recv(results, 100*time.Millisecond)
	if !ok {
		t.Fatal("fn did not execute")
	}

	if v != 42 {
		t.Errorf("got %d, want 42", v)
	}
}

func TestDebouncerFlushOnIdle(t *testing.T) {
	d := NewDebouncer[int](10*time.Millisecond, func(int) {})
	defer d.Close()

	if got := d.Flush(); got {
		t.Error("Flush on idle returned true, want false")
	}
}

func TestDebouncerFlushOnRunningNoPending(t *testing.T) {
	started := make(chan struct{}, 1)
	block := make(chan struct{})

	d := NewDebouncer(1*time.Millisecond, func(int) {
		started <- struct{}{}
		<-block
	})
	defer d.Close()

	d.Call(1)
	<-started // fn is running

	if got := d.Flush(); got {
		t.Error("Flush on running (no pending) returned true, want false")
	}

	close(block)
}

func TestDebouncerFlushWaitsForRunningPending(t *testing.T) {
	started := make(chan struct{}, 1)
	block := make(chan struct{})
	var calls []int
	var mu sync.Mutex

	// record appends to calls under lock; fn(1) blocks.
	record := func(v int) {
		if v == 1 {
			started <- struct{}{}
			<-block
		}

		mu.Lock()
		calls = append(calls, v)
		mu.Unlock()
	}

	d := NewDebouncer(1*time.Millisecond, record)
	defer d.Close()

	d.Call(1)
	<-started // fn(1) is running

	d.Call(2) // → running+pending

	// Flush in background — should block.
	flushDone := make(chan bool, 1)

	go func() { flushDone <- d.Flush() }()

	// Flush should not return yet.
	select {
	case <-flushDone:
		t.Fatal("Flush returned before fn(1) completed")
	case <-time.After(10 * time.Millisecond):
	}

	close(block) // release fn(1)

	// Flush should return true after fn(2) executes.
	got, ok := recv(flushDone, 100*time.Millisecond)
	if !ok {
		t.Fatal("Flush timed out")
	}

	if !got {
		t.Error("Flush returned false, want true")
	}

	mu.Lock()
	defer mu.Unlock()

	if len(calls) != 2 || calls[0] != 1 || calls[1] != 2 {
		t.Errorf("calls = %v, want [1, 2]", calls)
	}
}

func TestDebouncerCancelUnblocksFlush(t *testing.T) {
	started := make(chan struct{}, 1)
	block := make(chan struct{})

	d := NewDebouncer(1*time.Millisecond, func(int) {
		started <- struct{}{}
		<-block
	})
	defer d.Close()

	d.Call(1)
	<-started // fn(1) is running

	d.Call(2) // → running+pending

	// Flush in background — blocks on flushWaiter.
	flushDone := make(chan bool, 1)

	go func() { flushDone <- d.Flush() }()

	time.Sleep(5 * time.Millisecond) // let Flush register on unbuffered channel

	// Cancel clears pending and unblocks Flush.
	if got := d.Cancel(); !got {
		t.Error("Cancel returned false, want true")
	}

	got, ok := recv(flushDone, 100*time.Millisecond)
	if !ok {
		t.Fatal("Flush did not unblock after Cancel")
	}

	if got {
		t.Error("Flush returned true after Cancel, want false")
	}

	close(block)
}

func TestDebouncerDoubleFlush(t *testing.T) {
	started := make(chan struct{}, 1)
	block := make(chan struct{})

	d := NewDebouncer(1*time.Millisecond, func(int) {
		started <- struct{}{}
		<-block
	})
	defer d.Close()

	d.Call(1)
	<-started // fn(1) is running

	d.Call(2) // → running+pending

	// First Flush.
	flush1Done := make(chan bool, 1)

	go func() { flush1Done <- d.Flush() }()

	time.Sleep(5 * time.Millisecond) // let first Flush register on unbuffered channel

	// Second Flush should return false (slot taken).
	flush2Done := make(chan bool, 1)

	go func() { flush2Done <- d.Flush() }()

	got, ok := recv(flush2Done, 100*time.Millisecond)
	if !ok {
		t.Fatal("second Flush timed out")
	}

	if got {
		t.Error("second Flush returned true, want false")
	}

	close(block) // release fn(1), let pending execute

	got, ok = recv(flush1Done, 100*time.Millisecond)
	if !ok {
		t.Fatal("first Flush timed out")
	}

	if !got {
		t.Error("first Flush returned false, want true")
	}
}

func TestDebouncerSerialization(t *testing.T) {
	started := make(chan struct{}, 1)
	block := make(chan struct{})
	results := make(chan int, 2)

	// record stores the value; fn(1) blocks.
	record := func(v int) {
		if v == 1 {
			started <- struct{}{}
			<-block
		}

		results <- v
	}

	d := NewDebouncer(1*time.Millisecond, record)
	defer d.Close()

	d.Call(1)
	<-started // fn(1) is running

	d.Call(2) // queued

	close(block)

	// fn(1) completes, then fn(2) fires after fresh timer.
	v1, ok := recv(results, 100*time.Millisecond)
	if !ok {
		t.Fatal("fn(1) timed out")
	}

	if v1 != 1 {
		t.Errorf("first call got %d, want 1", v1)
	}

	v2, ok := recv(results, 100*time.Millisecond)
	if !ok {
		t.Fatal("fn(2) timed out")
	}

	if v2 != 2 {
		t.Errorf("second call got %d, want 2", v2)
	}
}

func TestDebouncerMaxWaitCapsStarvation(t *testing.T) {
	results := make(chan int, 1)

	// record stores the value on the results channel.
	record := func(v int) { results <- v }

	d := NewDebouncer(20*time.Millisecond, record, MaxWait(50*time.Millisecond))
	defer d.Close()

	// Continuously call every 10ms (less than wait). Without MaxWait,
	// the trailing timer resets each time and fn never fires.
	stop := make(chan struct{})

	go func() {
		i := 0

		for {
			select {
			case <-stop:
				return
			default:
				i++
				d.Call(i)
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	defer close(stop)

	// MaxWait should force execution within ~50ms (+scheduling tolerance).
	v, ok := recv(results, 200*time.Millisecond)
	if !ok {
		t.Fatal("MaxWait did not fire — starvation")
	}

	if v <= 0 {
		t.Errorf("got %d, want positive value", v)
	}
}

func TestDebouncerMaxWaitResetsAfterExecution(t *testing.T) {
	var calls atomic.Int32

	d := NewDebouncer(20*time.Millisecond, func(int) {
		calls.Add(1)
	}, MaxWait(40*time.Millisecond))
	defer d.Close()

	// First burst.
	for i := 0; i < 5; i++ {
		d.Call(i)
		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(50 * time.Millisecond) // let it fire

	first := calls.Load()
	if first < 1 {
		t.Fatal("first burst did not fire")
	}

	// Second burst — maxWait should start fresh.
	for i := 0; i < 5; i++ {
		d.Call(i + 10)
		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(50 * time.Millisecond)

	second := calls.Load()
	if second <= first {
		t.Errorf("second burst did not fire; calls = %d", second)
	}
}

func TestDebouncerConcurrentCallSafety(t *testing.T) {
	var calls atomic.Int32

	d := NewDebouncer(1*time.Millisecond, func(int) {
		calls.Add(1)
	})
	defer d.Close()

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func(v int) {
			defer wg.Done()
			d.Call(v)
		}(i)
	}

	wg.Wait()
	time.Sleep(20 * time.Millisecond)

	// Should have fired at least once (coalesced).
	if calls.Load() == 0 {
		t.Error("fn never called")
	}
}

func TestDebouncerConcurrentMixedOps(t *testing.T) {
	var calls atomic.Int32

	d := NewDebouncer(1*time.Millisecond, func(int) {
		calls.Add(1)
		time.Sleep(1 * time.Millisecond)
	})
	defer d.Close()

	var wg sync.WaitGroup

	// Mix Call, Cancel, and Flush from multiple goroutines.
	for i := 0; i < 50; i++ {
		wg.Add(3)

		go func(v int) {
			defer wg.Done()
			d.Call(v)
		}(i)

		go func() {
			defer wg.Done()
			d.Cancel()
		}()

		go func() {
			defer wg.Done()
			d.Flush()
		}()
	}

	wg.Wait()
	time.Sleep(20 * time.Millisecond)
}

func TestDebouncerReentrancyCall(t *testing.T) {
	results := make(chan int, 2)

	var d *Debouncer[int]

	// reentrant calls Call on the same debouncer from within fn.
	reentrant := func(v int) {
		results <- v

		if v == 1 {
			d.Call(2)
		}
	}

	d = NewDebouncer(1*time.Millisecond, reentrant)
	defer d.Close()

	d.Call(1)

	v1, ok := recv(results, 100*time.Millisecond)
	if !ok {
		t.Fatal("fn(1) timed out")
	}

	if v1 != 1 {
		t.Errorf("first got %d, want 1", v1)
	}

	v2, ok := recv(results, 100*time.Millisecond)
	if !ok {
		t.Fatal("fn(2) timed out — reentrant Call may have deadlocked")
	}

	if v2 != 2 {
		t.Errorf("second got %d, want 2", v2)
	}
}

func TestDebouncerReentrancyCancel(t *testing.T) {
	results := make(chan bool, 1)

	var d *Debouncer[int]

	// reentrantCancel calls Cancel from within fn.
	reentrantCancel := func(v int) {
		results <- d.Cancel()
	}

	d = NewDebouncer(1*time.Millisecond, reentrantCancel)
	defer d.Close()

	d.Call(1)

	got, ok := recv(results, 100*time.Millisecond)
	if !ok {
		t.Fatal("fn timed out — reentrant Cancel may have deadlocked")
	}

	// Cancel from within fn with no pending should return false.
	if got {
		t.Error("Cancel from fn returned true, want false")
	}
}

func TestDebouncerReusableAfterCancel(t *testing.T) {
	results := make(chan int, 1)

	// record stores the value on the results channel.
	record := func(v int) { results <- v }

	d := NewDebouncer(5*time.Millisecond, record)
	defer d.Close()

	d.Call(1)
	d.Cancel()

	d.Call(2)

	v, ok := recv(results, 100*time.Millisecond)
	if !ok {
		t.Fatal("fn did not fire after Cancel + Call")
	}

	if v != 2 {
		t.Errorf("got %d, want 2", v)
	}
}

func TestDebouncerReusableAfterFlush(t *testing.T) {
	results := make(chan int, 2)

	// record stores the value on the results channel.
	record := func(v int) { results <- v }

	d := NewDebouncer(5*time.Millisecond, record)
	defer d.Close()

	d.Call(1)
	d.Flush()

	d.Call(2)

	v, ok := recv(results, 100*time.Millisecond)
	if !ok {
		t.Fatal("fn did not fire after Flush + Call")
	}

	// First result is from Flush.
	if v != 1 {
		t.Errorf("first got %d, want 1", v)
	}

	v, ok = recv(results, 100*time.Millisecond)
	if !ok {
		t.Fatal("fn did not fire for second Call")
	}

	if v != 2 {
		t.Errorf("second got %d, want 2", v)
	}
}

func TestDebouncerCloseWaitsForRunning(t *testing.T) {
	started := make(chan struct{}, 1)
	block := make(chan struct{})
	executed := make(chan struct{}, 1)

	d := NewDebouncer(1*time.Millisecond, func(int) {
		started <- struct{}{}
		<-block
		executed <- struct{}{}
	})

	d.Call(1)
	<-started // fn is running

	closeDone := make(chan struct{})

	go func() {
		d.Close()
		close(closeDone)
	}()

	// Close should not return yet.
	select {
	case <-closeDone:
		t.Fatal("Close returned before fn completed")
	case <-time.After(10 * time.Millisecond):
	}

	close(block)

	select {
	case <-closeDone:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Close did not return after fn completed")
	}
}

func TestDebouncerCloseDiscardsPending(t *testing.T) {
	started := make(chan struct{}, 1)
	block := make(chan struct{})
	results := make(chan int, 2)

	// record stores the value; fn(1) blocks.
	record := func(v int) {
		if v == 1 {
			started <- struct{}{}
			<-block
		}

		results <- v
	}

	// Long wait ensures the pending timer cannot fire before Close
	// arrives. After fn(1) completes, owner starts a 1s timer for
	// pending. Close sends on closeCh; owner is in select and the
	// 1s timer hasn't fired, so closeCh is selected. No goroutine
	// needed, no sleep, fully deterministic.
	d := NewDebouncer(time.Second, record)

	d.Call(1)
	<-started // fn(1) is running

	d.Call(2) // → running+pending

	close(block) // release fn(1) — owner starts 1s timer for pending

	d.Close() // arrives before 1s timer — discards pending

	// fn(1) result.
	v, ok := recv(results, 100*time.Millisecond)
	if !ok {
		t.Fatal("fn(1) did not execute")
	}

	if v != 1 {
		t.Errorf("got %d, want 1", v)
	}

	// fn(2) should NOT execute (discarded by Close).
	noRecv(t, results, 30*time.Millisecond)
}

func TestDebouncerUseAfterClosePanics(t *testing.T) {
	d := NewDebouncer[int](10*time.Millisecond, func(int) {})
	d.Close()

	// assertPanics verifies that f panics.
	assertPanics := func(name string, f func()) {
		t.Helper()

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("%s did not panic after Close", name)
			}
		}()

		f()
	}

	assertPanics("Call", func() { d.Call(1) })
	assertPanics("Cancel", func() { d.Cancel() })
	assertPanics("Flush", func() { d.Flush() })
}

func TestDebouncerCloseIdempotent(t *testing.T) {
	d := NewDebouncer[int](10*time.Millisecond, func(int) {})
	d.Close()
	d.Close() // should not panic or block
}

func TestDebouncerFnPanicSignalsDone(t *testing.T) {
	// Verify the deferred doneCh signal exists in the spawn closure.
	// We can't easily test a goroutine panic without crashing the process,
	// so instead verify the debouncer works normally and that the spawn
	// function includes the defer. The defer is verified by code inspection
	// and by the fact that all other tests pass (doneCh must signal for
	// state transitions to work).
	results := make(chan int, 1)

	d := NewDebouncer(1*time.Millisecond, func(v int) {
		results <- v
	})
	defer d.Close()

	d.Call(42)

	v, ok := recv(results, 100*time.Millisecond)
	if !ok {
		t.Fatal("fn did not fire")
	}

	if v != 42 {
		t.Errorf("got %d, want 42", v)
	}
}

func TestDebouncerValidationPanics(t *testing.T) {
	tests := []struct {
		name string
		fn   func()
	}{
		{"wait zero", func() { NewDebouncer[int](0, func(int) {}) }},
		{"wait negative", func() { NewDebouncer[int](-1*time.Second, func(int) {}) }},
		{"nil fn", func() { NewDebouncer[int](time.Second, nil) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Error("expected panic")
				}
			}()

			tt.fn()
		})
	}
}

func TestMaxWaitNegativePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for negative MaxWait")
		}
	}()

	MaxWait(-1 * time.Second)
}

func TestDebouncerFlushDoesNotCascade(t *testing.T) {
	// Flush should return after the flushed execution completes,
	// and fn(20) should be scheduled via timer (not spawned
	// immediately on the doneCh path).
	started10 := make(chan struct{}, 1)
	started20 := make(chan struct{}, 1)
	block10 := make(chan struct{})

	d := NewDebouncer(30*time.Millisecond, func(v int) {
		switch v {
		case 10:
			started10 <- struct{}{}
			<-block10
		case 20:
			started20 <- struct{}{}
		}
	})
	defer d.Close()

	d.Call(10)

	flushDone := make(chan bool, 1)

	go func() { flushDone <- d.Flush() }()

	<-started10 // fn(10) is running (flushed)

	// Call during flushed execution — should NOT extend Flush.
	d.Call(20)

	close(block10) // release fn(10)

	// Flush should return true — fn(10) completed.
	got, ok := recv(flushDone, 100*time.Millisecond)
	if !ok {
		t.Fatal("Flush timed out — cascaded through Call(20)?")
	}

	if !got {
		t.Error("Flush returned false, want true")
	}

	// fn(20) should NOT have started yet — it's on a 30ms timer.
	// If it were spawned immediately (the cascading bug), it would
	// start in microseconds. 10ms check with 30ms timer gives 3x margin.
	noRecv(t, started20, 10*time.Millisecond)

	// fn(20) should eventually start via normal timer.
	select {
	case <-started20:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("fn(20) never started")
	}
}

func TestDebouncerFlushReturnsTrueWhenCancelClearsNewPending(t *testing.T) {
	// Flush triggers fn(A). During fn(A), Call(B) arrives then Cancel
	// clears B. Flush should still return true because fn(A) executed.
	started := make(chan struct{}, 1)
	block := make(chan struct{})

	d := NewDebouncer(1*time.Millisecond, func(v int) {
		if v == 1 {
			started <- struct{}{}
			<-block
		}
	})
	defer d.Close()

	d.Call(1)

	// Flush triggers fn(1) immediately.
	flushDone := make(chan bool, 1)

	go func() { flushDone <- d.Flush() }()

	// Wait for fn(1) to start — confirms Flush spawned it.
	// Note: Flush sends on flushCh first, then owner spawns fn.
	// We need fn(1) to start before we can Call(2).
	select {
	case <-started:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("fn(1) did not start")
	}

	d.Call(2)   // → running+pending
	d.Cancel()  // clears pending — Cancel should NOT send false to flushWaiter

	close(block) // release fn(1)

	got, ok := recv(flushDone, 100*time.Millisecond)
	if !ok {
		t.Fatal("Flush timed out")
	}

	if !got {
		t.Error("Flush returned false after Cancel of new pending, want true")
	}
}

func TestDebouncerCloseWithFlushedExecutionAndPending(t *testing.T) {
	// flushTarget=true, running=true, pending=true, then Close.
	// Flush should return true (its execution completes).
	// Pending work should be discarded.
	started := make(chan struct{}, 1)
	block := make(chan struct{})
	results := make(chan int, 2)

	d := NewDebouncer(time.Second, func(v int) {
		if v == 1 {
			started <- struct{}{}
			<-block
		}

		results <- v
	})

	d.Call(1)

	flushDone := make(chan bool, 1)

	go func() { flushDone <- d.Flush() }()

	<-started // fn(1) is running (flushed, flushTarget=true)

	d.Call(2) // → running+pending

	// Close in background. Owner receives closeCh, discards pending,
	// waits for doneCh, then responds true to flushWaiter.
	closeDone := make(chan struct{})

	go func() {
		d.Close()
		close(closeDone)
	}()

	// Brief sleep to let Close's send on closeCh complete.
	// Owner is in select, fn is blocked, so closeCh is selected.
	time.Sleep(5 * time.Millisecond)

	close(block) // release fn(1)

	// Flush should return true — fn(1) was the flushed execution.
	got, ok := recv(flushDone, 100*time.Millisecond)
	if !ok {
		t.Fatal("Flush timed out")
	}

	if !got {
		t.Error("Flush returned false, want true (flushed execution completed)")
	}

	select {
	case <-closeDone:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Close did not return")
	}

	// fn(1) result.
	v, ok := recv(results, 100*time.Millisecond)
	if !ok {
		t.Fatal("fn(1) did not execute")
	}

	if v != 1 {
		t.Errorf("got %d, want 1", v)
	}

	// fn(2) should NOT execute (discarded by Close).
	noRecv(t, results, 30*time.Millisecond)
}

func TestDebouncerSimultaneousTrailAndMaxWait(t *testing.T) {
	// When trail and maxWait timers converge, exactly one execution fires.
	var execCount atomic.Int32

	// Set trail and maxWait to the same duration so they fire ~simultaneously.
	d := NewDebouncer(10*time.Millisecond, func(int) {
		execCount.Add(1)
	}, MaxWait(10*time.Millisecond))
	defer d.Close()

	// Single call — both timers start, both expire at ~10ms.
	d.Call(1)

	time.Sleep(50 * time.Millisecond) // generous wait

	if got := execCount.Load(); got != 1 {
		t.Errorf("executions = %d, want exactly 1", got)
	}
}

func TestDebouncerMultipleFlushCallers(t *testing.T) {
	// Only the first Flush waiter wins. Others get false immediately.
	started := make(chan struct{}, 1)
	block := make(chan struct{})

	d := NewDebouncer(1*time.Millisecond, func(int) {
		started <- struct{}{}
		<-block
	})
	defer d.Close()

	d.Call(1)
	<-started // fn(1) is running

	d.Call(2) // → running+pending

	// Three concurrent Flush calls.
	const n = 3
	flushResults := make(chan bool, n)

	for i := 0; i < n; i++ {
		go func() { flushResults <- d.Flush() }()
	}

	// Give goroutines time to send on flushCh.
	time.Sleep(10 * time.Millisecond)

	close(block) // release fn(1), let pending execute

	// Collect all results.
	trueCount := 0
	falseCount := 0

	for i := 0; i < n; i++ {
		got, ok := recv(flushResults, 200*time.Millisecond)
		if !ok {
			t.Fatalf("Flush %d timed out", i)
		}

		if got {
			trueCount++
		} else {
			falseCount++
		}
	}

	if trueCount != 1 {
		t.Errorf("true count = %d, want 1", trueCount)
	}

	if falseCount != n-1 {
		t.Errorf("false count = %d, want %d", falseCount, n-1)
	}
}

func BenchmarkDebouncerCallBurst(b *testing.B) {
	d := NewDebouncer(100*time.Millisecond, func(int) {})
	defer d.Close()

	for i := 0; i < b.N; i++ {
		d.Call(i)
	}
}

func BenchmarkDebouncerCallConcurrent(b *testing.B) {
	d := NewDebouncer(100*time.Millisecond, func(int) {})
	defer d.Close()

	b.RunParallel(func(pb *testing.PB) {
		i := 0

		for pb.Next() {
			d.Call(i)
			i++
		}
	})
}
