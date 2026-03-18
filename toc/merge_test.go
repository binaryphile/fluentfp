package toc

import (
	"context"
	"errors"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/binaryphile/fluentfp/rslt"
)

func TestMergeHappyPath(t *testing.T) {
	ctx := context.Background()

	src1 := make(chan rslt.Result[int], 3)
	src2 := make(chan rslt.Result[int], 3)

	for i := range 3 {
		src1 <- rslt.Ok(i)
		src2 <- rslt.Ok(i + 100)
	}
	close(src1)
	close(src2)

	m := NewMerge[int](ctx, src1, src2)

	var vals []int
	for r := range m.Out() {
		v, err := r.Unpack()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		vals = append(vals, v)
	}

	if err := m.Wait(); err != nil {
		t.Fatalf("Wait: %v", err)
	}

	if len(vals) != 6 {
		t.Fatalf("got %d items, want 6", len(vals))
	}

	sort.Ints(vals)
	want := []int{0, 1, 2, 100, 101, 102}

	for i, v := range vals {
		if v != want[i] {
			t.Errorf("vals[%d] = %d, want %d", i, v, want[i])
		}
	}

	stats := m.Stats()
	if stats.Received != 6 {
		t.Errorf("Received = %d, want 6", stats.Received)
	}
	if stats.Forwarded != 6 {
		t.Errorf("Forwarded = %d, want 6", stats.Forwarded)
	}
	if stats.Dropped != 0 {
		t.Errorf("Dropped = %d, want 0", stats.Dropped)
	}
}

func TestMergeErrorsForwarded(t *testing.T) {
	ctx := context.Background()

	src1 := make(chan rslt.Result[int], 2)
	src2 := make(chan rslt.Result[int], 2)

	src1 <- rslt.Err[int](errTest)
	src1 <- rslt.Err[int](errTest)
	src2 <- rslt.Err[int](errTest)
	src2 <- rslt.Err[int](errTest)
	close(src1)
	close(src2)

	m := NewMerge[int](ctx, src1, src2)

	errCount := 0
	for r := range m.Out() {
		if _, err := r.Unpack(); err != nil {
			errCount++
		}
	}

	m.Wait()

	if errCount != 4 {
		t.Errorf("got %d errors, want 4", errCount)
	}

	stats := m.Stats()
	if stats.Forwarded != 4 {
		t.Errorf("Forwarded = %d, want 4", stats.Forwarded)
	}
}

func TestMergeMixedOkAndErr(t *testing.T) {
	ctx := context.Background()

	src1 := make(chan rslt.Result[string], 3)
	src2 := make(chan rslt.Result[string], 3)

	src1 <- rslt.Ok("a")
	src1 <- rslt.Err[string](errTest)
	src1 <- rslt.Ok("b")
	src2 <- rslt.Err[string](errTest)
	src2 <- rslt.Ok("c")
	src2 <- rslt.Err[string](errTest)
	close(src1)
	close(src2)

	m := NewMerge[string](ctx, src1, src2)

	var oks []string
	errs := 0

	for r := range m.Out() {
		if v, err := r.Unpack(); err != nil {
			errs++
		} else {
			oks = append(oks, v)
		}
	}

	m.Wait()

	sort.Strings(oks)
	if len(oks) != 3 || oks[0] != "a" || oks[1] != "b" || oks[2] != "c" {
		t.Errorf("oks = %v, want [a b c]", oks)
	}
	if errs != 3 {
		t.Errorf("errs = %d, want 3", errs)
	}
}

func TestMergeEmptySources(t *testing.T) {
	ctx := context.Background()

	src1 := make(chan rslt.Result[int])
	src2 := make(chan rslt.Result[int])
	close(src1)
	close(src2)

	m := NewMerge[int](ctx, src1, src2)

	count := 0
	for range m.Out() {
		count++
	}

	if err := m.Wait(); err != nil {
		t.Fatalf("Wait: %v", err)
	}

	if count != 0 {
		t.Errorf("got %d items, want 0", count)
	}

	stats := m.Stats()
	if stats.Received != 0 {
		t.Errorf("Received = %d, want 0", stats.Received)
	}
}

func TestMergeSingleSource(t *testing.T) {
	ctx := context.Background()

	src := make(chan rslt.Result[int], 3)
	src <- rslt.Ok(10)
	src <- rslt.Ok(20)
	src <- rslt.Ok(30)
	close(src)

	m := NewMerge[int](ctx, src)

	var vals []int
	for r := range m.Out() {
		v, _ := r.Unpack()
		vals = append(vals, v)
	}

	m.Wait()

	if len(vals) != 3 || vals[0] != 10 || vals[1] != 20 || vals[2] != 30 {
		t.Errorf("got %v, want [10 20 30]", vals)
	}

	stats := m.Stats()
	if stats.Forwarded != 3 {
		t.Errorf("Forwarded = %d, want 3", stats.Forwarded)
	}
}

func TestMergeCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	src1 := make(chan rslt.Result[int])
	src2 := make(chan rslt.Result[int])

	m := NewMerge[int](ctx, src1, src2)

	delivered := make(chan struct{})

	go func() {
		src1 <- rslt.Ok(1)
		src2 <- rslt.Ok(2)
		delivered <- struct{}{}
		<-ctx.Done()
		src1 <- rslt.Ok(99)
		src2 <- rslt.Ok(98)
		close(src1)
		close(src2)
	}()

	// Drain enough to know items were forwarded.
	<-m.Out()
	<-m.Out()
	<-delivered

	cancel()

	for range m.Out() {
	}

	err := m.Wait()
	if err == nil {
		t.Fatal("Wait: expected context.Canceled, got nil")
	}

	stats := m.Stats()
	if stats.Received != stats.Forwarded+stats.Dropped {
		t.Errorf("invariant: Received(%d) != Forwarded(%d)+Dropped(%d)",
			stats.Received, stats.Forwarded, stats.Dropped)
	}
}

func TestMergeCancelDuringOutputSend(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	src := make(chan rslt.Result[int], 1)
	src <- rslt.Ok(1)

	m := NewMerge[int](ctx, src)

	// Don't read Out — goroutine blocks on output send.
	// Cancel should unblock it.
	time.Sleep(50 * time.Millisecond)
	cancel()
	close(src)

	// Drain remaining.
	for range m.Out() {
	}

	m.Wait()

	// No deadlock is the test. Stats invariant should hold.
	stats := m.Stats()
	if stats.Received != stats.Forwarded+stats.Dropped {
		t.Errorf("invariant: Received(%d) != Forwarded(%d)+Dropped(%d)",
			stats.Received, stats.Forwarded, stats.Dropped)
	}
}

func TestMergePreCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	src1 := make(chan rslt.Result[int], 3)
	src2 := make(chan rslt.Result[int], 3)

	for i := range 3 {
		src1 <- rslt.Ok(i)
		src2 <- rslt.Ok(i + 10)
	}
	close(src1)
	close(src2)

	m := NewMerge[int](ctx, src1, src2)

	for range m.Out() {
	}

	err := m.Wait()
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Wait: got %v, want context.Canceled", err)
	}

	stats := m.Stats()
	// Pre-send checkpoint catches all items — all should be dropped.
	if stats.Forwarded != 0 {
		t.Errorf("Forwarded = %d, want 0", stats.Forwarded)
	}
	if stats.Dropped != stats.Received {
		t.Errorf("Dropped(%d) != Received(%d)", stats.Dropped, stats.Received)
	}
}

func TestMergeSourcesCloseAtDifferentTimes(t *testing.T) {
	ctx := context.Background()

	src1 := make(chan rslt.Result[int], 2)
	src2 := make(chan rslt.Result[int])

	src1 <- rslt.Ok(1)
	src1 <- rslt.Ok(2)
	close(src1)

	// src2 stays open briefly.
	go func() {
		time.Sleep(50 * time.Millisecond)
		src2 <- rslt.Ok(3)
		close(src2)
	}()

	m := NewMerge[int](ctx, src1, src2)

	var vals []int
	for r := range m.Out() {
		v, _ := r.Unpack()
		vals = append(vals, v)
	}

	m.Wait()

	sort.Ints(vals)
	if len(vals) != 3 || vals[0] != 1 || vals[1] != 2 || vals[2] != 3 {
		t.Errorf("got %v, want [1 2 3]", vals)
	}
}

func TestMergeSlowDownstreamThrottlesSources(t *testing.T) {
	ctx := context.Background()

	src1 := make(chan rslt.Result[int], 5)
	src2 := make(chan rslt.Result[int], 5)

	for i := range 5 {
		src1 <- rslt.Ok(i)
		src2 <- rslt.Ok(i + 100)
	}
	close(src1)
	close(src2)

	m := NewMerge[int](ctx, src1, src2)

	count := 0
	for range m.Out() {
		time.Sleep(5 * time.Millisecond)
		count++
	}

	m.Wait()

	if count != 10 {
		t.Errorf("got %d items, want 10", count)
	}
}

func TestMergeStatsInvariant(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	src1 := make(chan rslt.Result[int])
	src2 := make(chan rslt.Result[int])

	m := NewMerge[int](ctx, src1, src2)

	go func() {
		for i := range 10 {
			select {
			case src1 <- rslt.Ok(i):
			case <-ctx.Done():
			}
		}
		close(src1)
	}()

	go func() {
		for i := range 10 {
			select {
			case src2 <- rslt.Ok(i + 100):
			case <-ctx.Done():
			}
		}
		close(src2)
	}()

	count := 0
	for range m.Out() {
		count++
		if count >= 5 {
			cancel()
		}
	}

	m.Wait()

	stats := m.Stats()
	if stats.Received != stats.Forwarded+stats.Dropped {
		t.Errorf("invariant: Received(%d) != Forwarded(%d)+Dropped(%d)",
			stats.Received, stats.Forwarded, stats.Dropped)
	}
}

func TestMergePerSourceStats(t *testing.T) {
	ctx := context.Background()

	src1 := make(chan rslt.Result[int], 3)
	src2 := make(chan rslt.Result[int], 5)

	for range 3 {
		src1 <- rslt.Ok(1)
	}
	for range 5 {
		src2 <- rslt.Ok(2)
	}
	close(src1)
	close(src2)

	m := NewMerge[int](ctx, src1, src2)

	for range m.Out() {
	}
	m.Wait()

	stats := m.Stats()
	if stats.SourceReceived[0] != 3 {
		t.Errorf("SourceReceived[0] = %d, want 3", stats.SourceReceived[0])
	}
	if stats.SourceReceived[1] != 5 {
		t.Errorf("SourceReceived[1] = %d, want 5", stats.SourceReceived[1])
	}
	if stats.SourceForwarded[0] != 3 {
		t.Errorf("SourceForwarded[0] = %d, want 3", stats.SourceForwarded[0])
	}
	if stats.SourceForwarded[1] != 5 {
		t.Errorf("SourceForwarded[1] = %d, want 5", stats.SourceForwarded[1])
	}
}

func TestMergePerSourceInvariant(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	src1 := make(chan rslt.Result[int])
	src2 := make(chan rslt.Result[int])

	m := NewMerge[int](ctx, src1, src2)

	go func() {
		for i := range 20 {
			select {
			case src1 <- rslt.Ok(i):
			case <-ctx.Done():
			}
		}
		close(src1)
	}()

	go func() {
		for i := range 20 {
			select {
			case src2 <- rslt.Ok(i + 100):
			case <-ctx.Done():
			}
		}
		close(src2)
	}()

	count := 0
	for range m.Out() {
		count++
		if count >= 10 {
			cancel()
		}
	}

	m.Wait()

	stats := m.Stats()
	for i := range 2 {
		if stats.SourceReceived[i] != stats.SourceForwarded[i]+stats.SourceDropped[i] {
			t.Errorf("source %d: SourceReceived(%d) != SourceForwarded(%d)+SourceDropped(%d)",
				i, stats.SourceReceived[i], stats.SourceForwarded[i], stats.SourceDropped[i])
		}
	}
}

func TestMergeStatsDefensiveCopies(t *testing.T) {
	ctx := context.Background()

	src := make(chan rslt.Result[int], 1)
	src <- rslt.Ok(1)
	close(src)

	m := NewMerge[int](ctx, src)
	for range m.Out() {
	}
	m.Wait()

	stats1 := m.Stats()
	stats1.SourceReceived[0] = 999
	stats1.SourceForwarded[0] = 999
	stats1.SourceDropped[0] = 999

	stats2 := m.Stats()
	if stats2.SourceReceived[0] == 999 {
		t.Error("SourceReceived is not a defensive copy")
	}
	if stats2.SourceForwarded[0] == 999 {
		t.Error("SourceForwarded is not a defensive copy")
	}
	if stats2.SourceDropped[0] == 999 {
		t.Error("SourceDropped is not a defensive copy")
	}
}

func TestMergeMultipleWaitCalls(t *testing.T) {
	ctx := context.Background()

	src := make(chan rslt.Result[int], 1)
	src <- rslt.Ok(1)
	close(src)

	m := NewMerge[int](ctx, src)

	for range m.Out() {
	}

	err1 := m.Wait()
	err2 := m.Wait()
	err3 := m.Wait()

	if err1 != err2 || err2 != err3 {
		t.Errorf("Wait returned different values: %v, %v, %v", err1, err2, err3)
	}
}

func TestMergeOutReturnsSameChannel(t *testing.T) {
	src := make(chan rslt.Result[int])
	close(src)

	m := NewMerge[int](context.Background(), src)
	m.Wait()

	ch1 := m.Out()
	ch2 := m.Out()

	if ch1 != ch2 {
		t.Error("Out() returned different channels on successive calls")
	}
}

func TestMergeComposeWithPipe(t *testing.T) {
	ctx := context.Background()

	src1 := make(chan rslt.Result[int], 2)
	src2 := make(chan rslt.Result[int], 2)

	src1 <- rslt.Ok(1)
	src1 <- rslt.Ok(2)
	src2 <- rslt.Ok(3)
	src2 <- rslt.Ok(4)
	close(src1)
	close(src2)

	m := NewMerge[int](ctx, src1, src2)

	// doubler doubles an int.
	doubler := func(_ context.Context, v int) (int, error) {
		return v * 2, nil
	}

	pipe := Pipe(ctx, m.Out(), doubler, Options[int]{})

	var vals []int
	for r := range pipe.Out() {
		v, err := r.Unpack()
		if err != nil {
			t.Errorf("pipe error: %v", err)
		}

		vals = append(vals, v)
	}

	pipe.Wait()
	m.Wait()

	sort.Ints(vals)
	want := []int{2, 4, 6, 8}

	if len(vals) != 4 {
		t.Fatalf("got %d items, want 4", len(vals))
	}

	for i, v := range vals {
		if v != want[i] {
			t.Errorf("vals[%d] = %d, want %d", i, v, want[i])
		}
	}
}

func TestMergeComposeWithTee(t *testing.T) {
	ctx := context.Background()

	// Source → Tee → (Pipe, Pipe) → Merge
	src := make(chan rslt.Result[int], 3)
	src <- rslt.Ok(1)
	src <- rslt.Ok(2)
	src <- rslt.Ok(3)
	close(src)

	tee := NewTee[int](ctx, src, 2)

	// addTen adds 10 to an int.
	addTen := func(_ context.Context, v int) (int, error) {
		return v + 10, nil
	}

	// addHundred adds 100 to an int.
	addHundred := func(_ context.Context, v int) (int, error) {
		return v + 100, nil
	}

	pipe1 := Pipe(ctx, tee.Branch(0), addTen, Options[int]{})
	pipe2 := Pipe(ctx, tee.Branch(1), addHundred, Options[int]{})

	m := NewMerge[int](ctx, pipe1.Out(), pipe2.Out())

	var vals []int
	for r := range m.Out() {
		v, err := r.Unpack()
		if err != nil {
			t.Errorf("merge error: %v", err)
		}

		vals = append(vals, v)
	}

	m.Wait()
	pipe1.Wait()
	pipe2.Wait()
	tee.Wait()

	sort.Ints(vals)
	want := []int{11, 12, 13, 101, 102, 103}

	if len(vals) != 6 {
		t.Fatalf("got %d items, want 6", len(vals))
	}

	for i, v := range vals {
		if v != want[i] {
			t.Errorf("vals[%d] = %d, want %d", i, v, want[i])
		}
	}
}

func TestMergePreservesPerSourceOrder(t *testing.T) {
	ctx := context.Background()

	src1 := make(chan rslt.Result[int], 5)
	src2 := make(chan rslt.Result[int], 5)

	for i := range 5 {
		src1 <- rslt.Ok(i)         // 0,1,2,3,4
		src2 <- rslt.Ok(i + 100)   // 100,101,102,103,104
	}
	close(src1)
	close(src2)

	m := NewMerge[int](ctx, src1, src2)

	var from1, from2 []int
	for r := range m.Out() {
		v, _ := r.Unpack()
		if v < 100 {
			from1 = append(from1, v)
		} else {
			from2 = append(from2, v)
		}
	}

	m.Wait()

	// Per-source order must be preserved.
	for i := 1; i < len(from1); i++ {
		if from1[i] <= from1[i-1] {
			t.Errorf("source 1 out of order: %v", from1)

			break
		}
	}

	for i := 1; i < len(from2); i++ {
		if from2[i] <= from2[i-1] {
			t.Errorf("source 2 out of order: %v", from2)

			break
		}
	}
}

func TestMergeLateCancelAfterCompletion(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	src := make(chan rslt.Result[int], 2)
	src <- rslt.Ok(1)
	src <- rslt.Ok(2)
	close(src)

	m := NewMerge[int](ctx, src)

	for range m.Out() {
	}

	err := m.Wait()
	if err != nil {
		t.Fatalf("Wait: got %v, want nil", err)
	}

	// Cancel after natural completion — should not affect Wait result.
	cancel()

	err = m.Wait()
	if err != nil {
		t.Fatalf("Wait after late cancel: got %v, want nil", err)
	}
}

func TestMergeEmptySourcesPreCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	src1 := make(chan rslt.Result[int])
	src2 := make(chan rslt.Result[int])
	close(src1)
	close(src2)

	m := NewMerge[int](ctx, src1, src2)

	for range m.Out() {
	}

	// Empty sources + pre-canceled ctx → nothing to drop → Wait returns nil.
	err := m.Wait()
	if err != nil {
		t.Fatalf("Wait: got %v, want nil (nothing to drop)", err)
	}
}

func TestMergeCancelWhileBlockedOnReceive(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	src := make(chan rslt.Result[int])

	m := NewMerge[int](ctx, src)

	// Goroutine is blocked in `range src`. Cancel, then close src
	// without sending. Goroutine never enters a send path.
	time.Sleep(50 * time.Millisecond)
	cancel()
	close(src)

	for range m.Out() {
	}

	err := m.Wait()
	// Cancel not observed on a checked path → nil.
	if err != nil {
		t.Fatalf("Wait: got %v, want nil", err)
	}
}

func TestMergePostCancelForwardingBounded(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	const n = 3
	sources := make([]chan rslt.Result[int], n)
	srcChans := make([]<-chan rslt.Result[int], n)

	for i := range n {
		ch := make(chan rslt.Result[int], 100)
		sources[i] = ch
		srcChans[i] = ch
	}

	m := NewMerge[int](ctx, srcChans...)

	// Send one item per source, read them.
	for i := range n {
		sources[i] <- rslt.Ok(i)
	}

	received := 0
	for received < n {
		<-m.Out()
		received++
	}

	// Now cancel and flood items.
	cancel()
	for i := range n {
		for j := range 50 {
			sources[i] <- rslt.Ok(1000 + i*100 + j)
		}
		close(sources[i])
	}

	postCancel := 0
	for range m.Out() {
		postCancel++
	}

	m.Wait()

	// At most n items may forward after cancel (one per source goroutine
	// that already passed the pre-send checkpoint).
	if postCancel > n {
		t.Errorf("post-cancel forwarded %d items, want at most %d", postCancel, n)
	}
}

func TestMergeWaitNilWhenCancelRacesCompletion(t *testing.T) {
	// Run multiple times to exercise the race.
	for range 50 {
		ctx, cancel := context.WithCancel(context.Background())

		src := make(chan rslt.Result[int], 1)
		src <- rslt.Ok(1)
		close(src)

		m := NewMerge[int](ctx, src)

		for range m.Out() {
		}

		// Cancel races with natural completion.
		cancel()

		err := m.Wait()
		// Either nil (natural completion won) or context.Canceled
		// (cancel path taken). Both are valid.
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Fatalf("Wait: got %v, want nil or context.Canceled", err)
		}
	}
}

func TestMergeDeadlineExceeded(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(50*time.Millisecond))
	defer cancel()

	src := make(chan rslt.Result[int])

	m := NewMerge[int](ctx, src)

	// Send an item before deadline so goroutine is active.
	go func() {
		src <- rslt.Ok(1)
		// Wait for deadline to pass, then send another item so the
		// goroutine loops back and hits the pre-send checkpoint.
		<-ctx.Done()
		src <- rslt.Ok(2)
		close(src)
	}()

	for range m.Out() {
	}

	err := m.Wait()
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Wait: got %v, want context.DeadlineExceeded", err)
	}
}

func TestMergeStatsSingleCallCoherence(t *testing.T) {
	ctx := context.Background()

	src1 := make(chan rslt.Result[int], 10)
	src2 := make(chan rslt.Result[int], 10)

	for i := range 10 {
		src1 <- rslt.Ok(i)
		src2 <- rslt.Ok(i + 100)
	}
	close(src1)
	close(src2)

	m := NewMerge[int](ctx, src1, src2)

	for range m.Out() {
	}
	m.Wait()

	stats := m.Stats()

	// After Wait, all invariants hold.
	sumReceived := int64(0)
	sumForwarded := int64(0)
	sumDropped := int64(0)

	for i := range stats.SourceReceived {
		sumReceived += stats.SourceReceived[i]
		sumForwarded += stats.SourceForwarded[i]
		sumDropped += stats.SourceDropped[i]
	}

	if stats.Received != sumReceived {
		t.Errorf("Received(%d) != sum(SourceReceived)(%d)", stats.Received, sumReceived)
	}
	if stats.Forwarded != sumForwarded {
		t.Errorf("Forwarded(%d) != sum(SourceForwarded)(%d)", stats.Forwarded, sumForwarded)
	}
	if stats.Dropped != sumDropped {
		t.Errorf("Dropped(%d) != sum(SourceDropped)(%d)", stats.Dropped, sumDropped)
	}
}

func TestMergeConcurrentStatsUnderRace(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	src1 := make(chan rslt.Result[int])
	src2 := make(chan rslt.Result[int])

	m := NewMerge[int](ctx, src1, src2)

	// Hammer Stats() concurrently with active merge.
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()

		for i := range 100 {
			select {
			case src1 <- rslt.Ok(i):
			case <-ctx.Done():
				close(src1)

				return
			}
		}

		close(src1)
	}()

	go func() {
		defer wg.Done()

		for i := range 100 {
			select {
			case src2 <- rslt.Ok(i + 100):
			case <-ctx.Done():
				close(src2)

				return
			}
		}

		close(src2)
	}()

	go func() {
		defer wg.Done()

		for range 1000 {
			stats := m.Stats()
			// Should not panic. Per-metric coherence.
			if stats.Received < 0 || stats.Forwarded < 0 || stats.Dropped < 0 {
				t.Errorf("negative stat value")
			}
		}
	}()

	// Drain output.
	for range m.Out() {
	}

	cancel()
	wg.Wait()
	m.Wait()
}

func TestMergeOutClosedBeforeWaitReturns(t *testing.T) {
	ctx := context.Background()

	src := make(chan rslt.Result[int], 3)
	src <- rslt.Ok(1)
	src <- rslt.Ok(2)
	src <- rslt.Ok(3)
	close(src)

	m := NewMerge[int](ctx, src)

	outClosed := make(chan struct{})

	go func() {
		for range m.Out() {
		}
		close(outClosed)
	}()

	// Out() closes before Wait() returns.
	<-outClosed
	err := m.Wait()

	if err != nil {
		t.Fatalf("Wait: %v", err)
	}
}

func TestMergeStatsSourceIndexMatchesInputOrder(t *testing.T) {
	ctx := context.Background()

	// Source 0: 3 items. Source 1: 7 items. Source 2: 1 item.
	src0 := make(chan rslt.Result[int], 3)
	src1 := make(chan rslt.Result[int], 7)
	src2 := make(chan rslt.Result[int], 1)

	for range 3 {
		src0 <- rslt.Ok(0)
	}
	for range 7 {
		src1 <- rslt.Ok(1)
	}
	src2 <- rslt.Ok(2)
	close(src0)
	close(src1)
	close(src2)

	m := NewMerge[int](ctx, src0, src1, src2)

	for range m.Out() {
	}
	m.Wait()

	stats := m.Stats()
	if stats.SourceReceived[0] != 3 {
		t.Errorf("SourceReceived[0] = %d, want 3", stats.SourceReceived[0])
	}
	if stats.SourceReceived[1] != 7 {
		t.Errorf("SourceReceived[1] = %d, want 7", stats.SourceReceived[1])
	}
	if stats.SourceReceived[2] != 1 {
		t.Errorf("SourceReceived[2] = %d, want 1", stats.SourceReceived[2])
	}
}

func TestMergeDrainAfterCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	src := make(chan rslt.Result[int])

	m := NewMerge[int](ctx, src)

	// Send one item, read it.
	src <- rslt.Ok(1)
	<-m.Out()

	cancel()

	// Send known tail after cancel.
	src <- rslt.Ok(2)
	src <- rslt.Ok(3)
	close(src)

	for range m.Out() {
	}

	err := m.Wait()
	if err == nil {
		t.Fatal("Wait: expected error, got nil")
	}

	stats := m.Stats()
	// tail items (2, 3) should be counted as received and dropped.
	if stats.Received < 3 {
		t.Errorf("Received = %d, want >= 3", stats.Received)
	}
	if stats.Received != stats.Forwarded+stats.Dropped {
		t.Errorf("invariant: Received(%d) != Forwarded(%d)+Dropped(%d)",
			stats.Received, stats.Forwarded, stats.Dropped)
	}
}

func TestMergeCancelThenSendThenClose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	src := make(chan rslt.Result[int])

	m := NewMerge[int](ctx, src)

	// Goroutine is blocked in range src. Cancel fires, then source
	// sends one item, then closes. The goroutine receives the item
	// and either drops it (pre-send checkpoint catches cancel) or
	// forwards it (if it passed the checkpoint before observing cancel).
	time.Sleep(50 * time.Millisecond)
	cancel()
	src <- rslt.Ok(42)
	close(src)

	for range m.Out() {
	}

	err := m.Wait()
	// Cancel was observed on a checked path — Wait must return error.
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Wait: got %v, want context.Canceled", err)
	}

	stats := m.Stats()
	if stats.Received != 1 {
		t.Errorf("Received = %d, want 1", stats.Received)
	}
	if stats.Received != stats.Forwarded+stats.Dropped {
		t.Errorf("invariant: Received(%d) != Forwarded(%d)+Dropped(%d)",
			stats.Received, stats.Forwarded, stats.Dropped)
	}
}

func TestMergeConcurrentWaitCallers(t *testing.T) {
	ctx := context.Background()

	src := make(chan rslt.Result[int], 3)
	src <- rslt.Ok(1)
	src <- rslt.Ok(2)
	src <- rslt.Ok(3)
	close(src)

	m := NewMerge[int](ctx, src)

	for range m.Out() {
	}

	// Multiple goroutines calling Wait concurrently.
	var wg sync.WaitGroup
	errs := make([]error, 10)

	for i := range 10 {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			errs[i] = m.Wait()
		}(i)
	}

	wg.Wait()

	for i := range 10 {
		if errs[i] != errs[0] {
			t.Errorf("Wait() caller %d returned %v, caller 0 returned %v", i, errs[i], errs[0])
		}
	}
}

func TestMergePanicZeroSources(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic, got none")
		}
	}()

	NewMerge[int](context.Background())
}

func TestMergePanicNilCtx(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic, got none")
		}
	}()

	src := make(chan rslt.Result[int])
	NewMerge[int](nil, src)
}

func TestMergePanicNilSource(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic, got none")
		}
	}()

	src := make(chan rslt.Result[int])
	NewMerge[int](context.Background(), src, nil)
}
