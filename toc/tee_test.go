package toc

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/binaryphile/fluentfp/rslt"
)

var errTest = errors.New("test error")

func TestTeeHappyPath(t *testing.T) {
	ctx := context.Background()
	src := make(chan rslt.Result[int], 5)

	for i := range 5 {
		src <- rslt.Ok(i)
	}
	close(src)

	tee := NewTee[int](ctx, src, 3)

	type branchResult struct {
		vals []int
	}

	results := make([]branchResult, 3)
	var wg sync.WaitGroup

	for b := range 3 {
		wg.Add(1)

		go func(b int) {
			defer wg.Done()

			for r := range tee.Branch(b) {
				v, err := r.Unpack()
				if err != nil {
					t.Errorf("branch %d: unexpected error: %v", b, err)
				}

				results[b].vals = append(results[b].vals, v)
			}
		}(b)
	}

	wg.Wait()

	if err := tee.Wait(); err != nil {
		t.Fatalf("Wait: %v", err)
	}

	for b := range 3 {
		if len(results[b].vals) != 5 {
			t.Errorf("branch %d: got %d items, want 5", b, len(results[b].vals))
		}

		for i, v := range results[b].vals {
			if v != i {
				t.Errorf("branch %d[%d] = %d, want %d", b, i, v, i)
			}
		}
	}

	stats := tee.Stats()
	if stats.Received != 5 {
		t.Errorf("Received = %d, want 5", stats.Received)
	}
	if stats.FullyDelivered != 5 {
		t.Errorf("FullyDelivered = %d, want 5", stats.FullyDelivered)
	}
}

func TestTeeErrorsForwarded(t *testing.T) {
	ctx := context.Background()
	src := make(chan rslt.Result[int], 3)
	src <- rslt.Err[int](errTest)
	src <- rslt.Err[int](errTest)
	src <- rslt.Err[int](errTest)
	close(src)

	tee := NewTee[int](ctx, src, 2)

	counts := [2]int{}
	var wg sync.WaitGroup

	for b := range 2 {
		wg.Add(1)

		go func(b int) {
			defer wg.Done()

			for r := range tee.Branch(b) {
				if _, err := r.Unpack(); err == nil {
					t.Errorf("branch %d: expected error, got Ok", b)
				}

				counts[b]++
			}
		}(b)
	}

	wg.Wait()

	if err := tee.Wait(); err != nil {
		t.Fatalf("Wait: %v", err)
	}

	for b := range 2 {
		if counts[b] != 3 {
			t.Errorf("branch %d: got %d errors, want 3", b, counts[b])
		}
	}

	stats := tee.Stats()
	if stats.FullyDelivered != 3 {
		t.Errorf("FullyDelivered = %d, want 3", stats.FullyDelivered)
	}
}

func TestTeeMixedOkAndErr(t *testing.T) {
	ctx := context.Background()
	src := make(chan rslt.Result[string], 4)
	src <- rslt.Ok("a")
	src <- rslt.Err[string](errTest)
	src <- rslt.Ok("b")
	src <- rslt.Err[string](errTest)
	close(src)

	tee := NewTee[string](ctx, src, 2)

	type result struct {
		oks  []string
		errs int
	}

	results := [2]result{}
	var wg sync.WaitGroup

	for b := range 2 {
		wg.Add(1)

		go func(b int) {
			defer wg.Done()

			for r := range tee.Branch(b) {
				if v, err := r.Unpack(); err != nil {
					results[b].errs++
				} else {
					results[b].oks = append(results[b].oks, v)
				}
			}
		}(b)
	}

	wg.Wait()
	tee.Wait()

	for b := range 2 {
		if len(results[b].oks) != 2 || results[b].oks[0] != "a" || results[b].oks[1] != "b" {
			t.Errorf("branch %d: oks = %v, want [a b]", b, results[b].oks)
		}

		if results[b].errs != 2 {
			t.Errorf("branch %d: errs = %d, want 2", b, results[b].errs)
		}
	}
}

func TestTeeEmptySource(t *testing.T) {
	ctx := context.Background()
	src := make(chan rslt.Result[int])
	close(src)

	tee := NewTee[int](ctx, src, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		for r := range tee.Branch(0) {
			t.Errorf("branch 0: unexpected item: %v", r)
		}
	}()

	go func() {
		defer wg.Done()

		for r := range tee.Branch(1) {
			t.Errorf("branch 1: unexpected item: %v", r)
		}
	}()

	wg.Wait()

	if err := tee.Wait(); err != nil {
		t.Fatalf("Wait: %v", err)
	}

	stats := tee.Stats()
	if stats.Received != 0 {
		t.Errorf("Received = %d, want 0", stats.Received)
	}
}

func TestTeeSingleBranch(t *testing.T) {
	ctx := context.Background()
	src := make(chan rslt.Result[int], 3)
	src <- rslt.Ok(10)
	src <- rslt.Ok(20)
	src <- rslt.Ok(30)
	close(src)

	tee := NewTee[int](ctx, src, 1)

	var vals []int
	for r := range tee.Branch(0) {
		v, _ := r.Unpack()
		vals = append(vals, v)
	}

	if len(vals) != 3 || vals[0] != 10 || vals[1] != 20 || vals[2] != 30 {
		t.Errorf("got %v, want [10 20 30]", vals)
	}

	tee.Wait()

	stats := tee.Stats()
	if stats.FullyDelivered != 3 {
		t.Errorf("FullyDelivered = %d, want 3", stats.FullyDelivered)
	}
}

func TestTeeCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	src := make(chan rslt.Result[int])

	tee := NewTee[int](ctx, src, 2)

	// Send a few items, then cancel.
	delivered := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		for range tee.Branch(0) {
		}
	}()

	go func() {
		defer wg.Done()

		for range tee.Branch(1) {
		}
	}()

	go func() {
		src <- rslt.Ok(1)
		src <- rslt.Ok(2)
		delivered <- struct{}{}
		// Wait for cancel before closing src.
		<-ctx.Done()
		src <- rslt.Ok(99) // should be discarded
		close(src)
	}()

	<-delivered
	cancel()

	wg.Wait()

	err := tee.Wait()
	if err == nil {
		t.Fatal("Wait: expected context.Canceled, got nil")
	}

	stats := tee.Stats()
	total := stats.FullyDelivered + stats.PartiallyDelivered + stats.Undelivered
	if stats.Received != total {
		t.Errorf("invariant: Received(%d) != FD(%d)+PD(%d)+U(%d)=%d",
			stats.Received, stats.FullyDelivered, stats.PartiallyDelivered, stats.Undelivered, total)
	}
}

func TestTeeCancelDuringSend(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	src := make(chan rslt.Result[int], 1)
	src <- rslt.Ok(1)

	tee := NewTee[int](ctx, src, 2)

	// Read branch 0 so the Tee can deliver to it, then cancel while
	// blocked on branch 1 (which nobody reads initially).
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		<-tee.Branch(0)
		cancel()
		close(src)
	}()

	go func() {
		defer wg.Done()

		for range tee.Branch(1) {
		}
	}()

	wg.Wait()
	err := tee.Wait()

	stats := tee.Stats()
	total := stats.FullyDelivered + stats.PartiallyDelivered + stats.Undelivered

	// Invariant must hold regardless of race outcome.
	if stats.Received != total {
		t.Errorf("invariant: Received(%d) != FD(%d)+PD(%d)+U(%d)=%d",
			stats.Received, stats.FullyDelivered, stats.PartiallyDelivered, stats.Undelivered, total)
	}

	// Either: send to branch 1 won (FullyDelivered=1, err=nil)
	// or: cancel won (PartiallyDelivered=1, err=context.Canceled)
	switch {
	case stats.FullyDelivered == 1 && err == nil:
		// send won the race
	case stats.PartiallyDelivered == 1 && errors.Is(err, context.Canceled):
		// cancel won the race
	default:
		t.Errorf("unexpected state: FD=%d PD=%d U=%d err=%v",
			stats.FullyDelivered, stats.PartiallyDelivered, stats.Undelivered, err)
	}
}

func TestTeePartialDelivery(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	src := make(chan rslt.Result[int], 1)
	src <- rslt.Ok(42)

	tee := NewTee[int](ctx, src, 2)

	// Branch 0: read one item, then signal.
	got := make(chan int, 1)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		r := <-tee.Branch(0)
		v, _ := r.Unpack()
		got <- v
	}()

	// Branch 1: nobody reads. Tee blocks on branch 1 send.
	// Cancel after branch 0 receives.
	v := <-got
	if v != 42 {
		t.Errorf("branch 0 got %d, want 42", v)
	}

	cancel()
	close(src)

	// Drain branch 1 to let Tee finish.
	go func() {
		defer wg.Done()

		for range tee.Branch(1) {
		}
	}()

	wg.Wait()
	tee.Wait()

	stats := tee.Stats()
	// Cancel races with delivery. Any of these is valid:
	// - FullyDelivered=1 (both branches got it before cancel)
	// - PartiallyDelivered=1 (branch 0 got it, branch 1 didn't)
	// - Undelivered=1 (cancel won before any delivery)
	// The invariant: exactly 1 item was received.
	total := stats.FullyDelivered + stats.PartiallyDelivered + stats.Undelivered
	if total != 1 {
		t.Errorf("total items = %d (full=%d partial=%d undel=%d), want 1",
			total, stats.FullyDelivered, stats.PartiallyDelivered, stats.Undelivered)
	}
}

func TestTeeSlowBranchThrottlesFast(t *testing.T) {
	ctx := context.Background()
	src := make(chan rslt.Result[int], 5)

	for i := range 5 {
		src <- rslt.Ok(i)
	}
	close(src)

	tee := NewTee[int](ctx, src, 2)

	// Branch 0: slow — sleep before each read.
	// Branch 1: fast — read immediately.
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		for range tee.Branch(0) {
			time.Sleep(10 * time.Millisecond)
		}
	}()

	go func() {
		defer wg.Done()

		for range tee.Branch(1) {
		}
	}()

	wg.Wait()
	tee.Wait()

	stats := tee.Stats()
	if stats.BranchDelivered[0] != 5 {
		t.Errorf("BranchDelivered[0] = %d, want 5", stats.BranchDelivered[0])
	}
	if stats.BranchDelivered[1] != 5 {
		t.Errorf("BranchDelivered[1] = %d, want 5", stats.BranchDelivered[1])
	}
	// Branch 0 blocked time should exceed branch 1 (slow reader creates
	// more send-side wait). Relative comparison avoids CI timing flakiness.
	if stats.BranchBlockedTime[0] <= stats.BranchBlockedTime[1] {
		t.Errorf("expected BranchBlockedTime[0](%v) > BranchBlockedTime[1](%v)",
			stats.BranchBlockedTime[0], stats.BranchBlockedTime[1])
	}
}

func TestTeeAbandonedBranchBlocks(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	src := make(chan rslt.Result[int], 1)
	src <- rslt.Ok(1)

	tee := NewTee[int](ctx, src, 2)

	// Branch 0 reads one item and signals.
	gotOne := make(chan struct{})

	go func() {
		<-tee.Branch(0)
		close(gotOne)
		// Stop reading — now Tee is blocked on branch 1.
	}()

	// Synchronize: branch 0 has received its item.
	<-gotOne

	// Tee should be blocked on branch 1 send. Use a generous deadline
	// to avoid CI flakiness — we just need to prove it didn't finish.
	select {
	case <-tee.done:
		t.Fatal("Tee finished despite abandoned branch 1")
	case <-time.After(200 * time.Millisecond):
		// Expected: Tee is blocked.
	}

	// Cancel releases the Tee.
	cancel()
	close(src)

	// Drain branch 1 to let close happen.
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		for range tee.Branch(1) {
		}
	}()

	wg.Wait()
	tee.Wait()
}

func TestTeeStatsInvariant(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	src := make(chan rslt.Result[int])

	tee := NewTee[int](ctx, src, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		for range tee.Branch(0) {
		}
	}()

	go func() {
		defer wg.Done()

		for range tee.Branch(1) {
		}
	}()

	// Send items then cancel.
	for i := range 10 {
		select {
		case src <- rslt.Ok(i):
		case <-time.After(100 * time.Millisecond):
			// Tee may be blocked; cancel and move on.
			cancel()
		}
	}

	cancel()
	close(src)
	wg.Wait()
	tee.Wait()

	stats := tee.Stats()
	total := stats.FullyDelivered + stats.PartiallyDelivered + stats.Undelivered

	if stats.Received != total {
		t.Errorf("invariant broken: Received(%d) != FD(%d)+PD(%d)+U(%d)=%d",
			stats.Received, stats.FullyDelivered, stats.PartiallyDelivered, stats.Undelivered, total)
	}
}

func TestTeePerBranchStats(t *testing.T) {
	ctx := context.Background()
	src := make(chan rslt.Result[int], 3)
	src <- rslt.Ok(1)
	src <- rslt.Ok(2)
	src <- rslt.Ok(3)
	close(src)

	tee := NewTee[int](ctx, src, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		for range tee.Branch(0) {
		}
	}()

	go func() {
		defer wg.Done()

		for range tee.Branch(1) {
		}
	}()

	wg.Wait()
	tee.Wait()

	stats := tee.Stats()

	if len(stats.BranchDelivered) != 2 {
		t.Fatalf("BranchDelivered len = %d, want 2", len(stats.BranchDelivered))
	}

	if len(stats.BranchBlockedTime) != 2 {
		t.Fatalf("BranchBlockedTime len = %d, want 2", len(stats.BranchBlockedTime))
	}

	for b := range 2 {
		if stats.BranchDelivered[b] != 3 {
			t.Errorf("BranchDelivered[%d] = %d, want 3", b, stats.BranchDelivered[b])
		}
	}

	// Cross-field: sum of per-branch delivered should equal
	// FullyDelivered * N for a no-cancel run.
	n := int64(2)
	sumDelivered := stats.BranchDelivered[0] + stats.BranchDelivered[1]

	if sumDelivered != stats.FullyDelivered*n {
		t.Errorf("sum(BranchDelivered)=%d != FullyDelivered(%d)*N(%d)=%d",
			sumDelivered, stats.FullyDelivered, n, stats.FullyDelivered*n)
	}
}

func TestTeePanics(t *testing.T) {
	tests := []struct {
		name string
		fn   func()
	}{
		{
			name: "n<=0",
			fn:   func() { NewTee[int](context.Background(), make(chan rslt.Result[int]), 0) },
		},
		{
			name: "n negative",
			fn:   func() { NewTee[int](context.Background(), make(chan rslt.Result[int]), -1) },
		},
		{
			name: "nil src",
			fn:   func() { NewTee[int](context.Background(), nil, 2) },
		},
		{
			name: "nil ctx",
			fn:   func() { NewTee[int](nil, make(chan rslt.Result[int]), 2) },
		},
		{
			name: "Branch out of range negative",
			fn: func() {
				// Use closed src so the goroutine exits promptly.
				src := make(chan rslt.Result[int])
				close(src)

				tee := NewTee[int](context.Background(), src, 2)
				tee.Branch(-1)
			},
		},
		{
			name: "Branch out of range high",
			fn: func() {
				src := make(chan rslt.Result[int])
				close(src)

				tee := NewTee[int](context.Background(), src, 2)
				tee.Branch(2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Error("expected panic, got none")
				}
			}()

			tt.fn()
		})
	}
}

func TestTeePayloadAliasing(t *testing.T) {
	ctx := context.Background()

	type payload struct {
		data []int
	}

	src := make(chan rslt.Result[*payload], 1)
	p := &payload{data: []int{1, 2, 3}}
	src <- rslt.Ok(p)
	close(src)

	tee := NewTee[*payload](ctx, src, 2)

	var p0, p1 *payload
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		r := <-tee.Branch(0)
		p0, _ = r.Unpack()
	}()

	go func() {
		defer wg.Done()

		r := <-tee.Branch(1)
		p1, _ = r.Unpack()
	}()

	wg.Wait()
	tee.Wait()

	// Same pointer — no deep copy contract.
	if p0 != p1 {
		t.Error("branches received different pointers; expected same reference")
	}
	if p0 != p {
		t.Error("branch pointer differs from original; expected same reference")
	}
}

func TestTeeBranchCloseDelayedUntilSrcClose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	src := make(chan rslt.Result[int])

	tee := NewTee[int](ctx, src, 2)

	// Drain branches in background.
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		for range tee.Branch(0) {
		}
	}()

	go func() {
		defer wg.Done()

		for range tee.Branch(1) {
		}
	}()

	// Cancel context — Tee enters discard mode but src is still open.
	cancel()

	// Tee should NOT be done yet because src hasn't closed.
	select {
	case <-tee.done:
		t.Fatal("Tee finished before src closed")
	case <-time.After(200 * time.Millisecond):
		// Expected.
	}

	// Now close src — Tee should finish.
	close(src)

	wg.Wait()
	tee.Wait()
}

func TestTeeStatsSnapshotMidFlight(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	src := make(chan rslt.Result[int])
	tee := NewTee[int](ctx, src, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		for range tee.Branch(0) {
		}
	}()

	go func() {
		defer wg.Done()

		for range tee.Branch(1) {
		}
	}()

	// Send one item.
	src <- rslt.Ok(1)

	// Mid-flight snapshot should not panic.
	stats := tee.Stats()
	if stats.Received < 0 {
		t.Error("Received is negative")
	}

	// Values should be monotonically non-decreasing.
	stats2 := tee.Stats()
	if stats2.Received < stats.Received {
		t.Errorf("Received decreased: %d -> %d", stats.Received, stats2.Received)
	}

	cancel()
	close(src)
	wg.Wait()
	tee.Wait()
}

func TestTeeComposeWithPipe(t *testing.T) {
	ctx := context.Background()
	src := make(chan rslt.Result[int], 3)
	src <- rslt.Ok(1)
	src <- rslt.Ok(2)
	src <- rslt.Ok(3)
	close(src)

	tee := NewTee[int](ctx, src, 2)

	// doubler doubles an int.
	doubler := func(_ context.Context, v int) (int, error) {
		return v * 2, nil
	}

	// Branch 0 feeds a Pipe stage.
	pipe := Pipe(ctx, tee.Branch(0), doubler, Options[int]{})

	// Branch 1 just drains.
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		for range tee.Branch(1) {
		}
	}()

	var doubled []int
	for r := range pipe.Out() {
		v, err := r.Unpack()
		if err != nil {
			t.Errorf("pipe error: %v", err)
		}

		doubled = append(doubled, v)
	}

	wg.Wait()
	pipe.Wait()
	tee.Wait()

	if len(doubled) != 3 {
		t.Fatalf("got %d items, want 3", len(doubled))
	}

	// With Workers=1 (default), order is preserved.
	want := []int{2, 4, 6}
	for i, v := range doubled {
		if v != want[i] {
			t.Errorf("doubled[%d] = %d, want %d", i, v, want[i])
		}
	}
}

func TestTeePreCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-canceled

	src := make(chan rslt.Result[int], 3)
	src <- rslt.Ok(1)
	src <- rslt.Ok(2)
	src <- rslt.Ok(3)
	close(src)

	tee := NewTee[int](ctx, src, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		for range tee.Branch(0) {
		}
	}()

	go func() {
		defer wg.Done()

		for range tee.Branch(1) {
		}
	}()

	wg.Wait()
	err := tee.Wait()

	if err == nil {
		t.Fatal("Wait: expected error for pre-canceled context")
	}

	stats := tee.Stats()
	// All items should be undelivered.
	if stats.Undelivered != stats.Received {
		t.Errorf("expected all items undelivered: Received=%d Undelivered=%d",
			stats.Received, stats.Undelivered)
	}
	if stats.FullyDelivered != 0 {
		t.Errorf("FullyDelivered = %d, want 0", stats.FullyDelivered)
	}

	total := stats.FullyDelivered + stats.PartiallyDelivered + stats.Undelivered
	if stats.Received != total {
		t.Errorf("invariant: Received(%d) != total(%d)", stats.Received, total)
	}
}

func TestTeeMultipleWaitCalls(t *testing.T) {
	ctx := context.Background()
	src := make(chan rslt.Result[int], 1)
	src <- rslt.Ok(1)
	close(src)

	tee := NewTee[int](ctx, src, 1)

	for range tee.Branch(0) {
	}

	// Multiple sequential Wait calls should be safe.
	err1 := tee.Wait()
	err2 := tee.Wait()
	err3 := tee.Wait()

	if err1 != err2 || err2 != err3 {
		t.Errorf("Wait returned different values: %v, %v, %v", err1, err2, err3)
	}
}

func TestTeeBranchReturnsSameChannel(t *testing.T) {
	src := make(chan rslt.Result[int])
	close(src)

	tee := NewTee[int](context.Background(), src, 2)
	tee.Wait()

	ch0a := tee.Branch(0)
	ch0b := tee.Branch(0)
	ch1 := tee.Branch(1)

	if ch0a != ch0b {
		t.Error("Branch(0) returned different channels on successive calls")
	}
	if ch0a == ch1 {
		t.Error("Branch(0) and Branch(1) returned the same channel")
	}
}

func TestTeeStatsSlicesAreDefensiveCopies(t *testing.T) {
	ctx := context.Background()
	src := make(chan rslt.Result[int], 2)
	src <- rslt.Ok(1)
	src <- rslt.Ok(2)
	close(src)

	tee := NewTee[int](ctx, src, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		for range tee.Branch(0) {
		}
	}()

	go func() {
		defer wg.Done()

		for range tee.Branch(1) {
		}
	}()

	wg.Wait()
	tee.Wait()

	stats1 := tee.Stats()

	// Mutate the returned slices.
	stats1.BranchDelivered[0] = 999
	stats1.BranchBlockedTime[0] = 999 * time.Hour

	// Fresh Stats() should be unaffected.
	stats2 := tee.Stats()
	if stats2.BranchDelivered[0] == 999 {
		t.Error("BranchDelivered is not a defensive copy")
	}
	if stats2.BranchBlockedTime[0] == 999*time.Hour {
		t.Error("BranchBlockedTime is not a defensive copy")
	}
}
