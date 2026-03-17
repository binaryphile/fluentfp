package toc_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/binaryphile/fluentfp/rslt"
	"github.com/binaryphile/fluentfp/toc"
)

func TestBatcherHappyPath(t *testing.T) {
	src := make(chan rslt.Result[int], 10)
	for i := 1; i <= 10; i++ {
		src <- rslt.Ok(i)
	}
	close(src)

	b := toc.NewBatcher(context.Background(), src, 3)

	var batches [][]int
	for r := range b.Out() {
		v, err := r.Unpack()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		batches = append(batches, v)
	}

	// 10 items / 3 = 3 full + 1 partial (1 item).
	if len(batches) != 4 {
		t.Fatalf("got %d batches, want 4", len(batches))
	}
	if len(batches[0]) != 3 || len(batches[1]) != 3 || len(batches[2]) != 3 {
		t.Fatalf("first 3 batches should have 3 items each")
	}
	if len(batches[3]) != 1 {
		t.Fatalf("last batch should have 1 item, got %d", len(batches[3]))
	}

	if err := b.Wait(); err != nil {
		t.Fatalf("Wait: %v", err)
	}

	stats := b.Stats()
	if stats.Received != 10 {
		t.Errorf("Received = %d, want 10", stats.Received)
	}
	if stats.Emitted != 10 {
		t.Errorf("Emitted = %d, want 10", stats.Emitted)
	}
	if stats.BatchCount != 4 {
		t.Errorf("BatchCount = %d, want 4", stats.BatchCount)
	}
}

func TestBatcherErrorAsBoundary(t *testing.T) {
	testErr := errors.New("mid-error")
	src := feedResults(
		rslt.Ok(1),
		rslt.Ok(2),
		rslt.Err[int](testErr),
		rslt.Ok(3),
		rslt.Ok(4),
	)

	b := toc.NewBatcher(context.Background(), src, 4)

	type item struct {
		batch []int
		isErr bool
	}
	var items []item
	for r := range b.Out() {
		if v, err := r.Unpack(); err != nil {
			items = append(items, item{isErr: true})
		} else {
			items = append(items, item{batch: v})
		}
	}

	// Expected: Ok([1,2]), Err, Ok([3,4])
	if len(items) != 3 {
		t.Fatalf("got %d items, want 3", len(items))
	}

	if items[0].isErr || len(items[0].batch) != 2 {
		t.Errorf("item[0]: want batch of 2, got %+v", items[0])
	}
	if !items[1].isErr {
		t.Errorf("item[1]: want error")
	}
	if items[2].isErr || len(items[2].batch) != 2 {
		t.Errorf("item[2]: want batch of 2, got %+v", items[2])
	}

	b.Wait()
}

func TestBatcherEncounterOrder(t *testing.T) {
	testErr := errors.New("err")
	src := feedResults(
		rslt.Ok(10),
		rslt.Ok(20),
		rslt.Err[int](testErr),
		rslt.Ok(30),
	)

	b := toc.NewBatcher(context.Background(), src, 5)

	type entry struct {
		vals []int
		err  bool
	}
	var entries []entry
	for r := range b.Out() {
		if v, err := r.Unpack(); err != nil {
			entries = append(entries, entry{err: true})
		} else {
			entries = append(entries, entry{vals: v})
		}
	}

	// Items before error appear before error.
	if len(entries) != 3 {
		t.Fatalf("got %d entries, want 3", len(entries))
	}
	if entries[0].err || entries[0].vals[0] != 10 || entries[0].vals[1] != 20 {
		t.Errorf("entry[0]: want [10,20], got %+v", entries[0])
	}
	if !entries[1].err {
		t.Errorf("entry[1]: want error")
	}
	if entries[2].err || entries[2].vals[0] != 30 {
		t.Errorf("entry[2]: want [30], got %+v", entries[2])
	}

	b.Wait()
}

func TestBatcherEmpty(t *testing.T) {
	src := make(chan rslt.Result[int])
	close(src)

	b := toc.NewBatcher(context.Background(), src, 5)

	count := 0
	for range b.Out() {
		count++
	}

	if count != 0 {
		t.Fatalf("got %d items, want 0", count)
	}

	b.Wait()
}

func TestBatcherAllErrors(t *testing.T) {
	src := feedResults(
		rslt.Err[int](errors.New("e1")),
		rslt.Err[int](errors.New("e2")),
		rslt.Err[int](errors.New("e3")),
	)

	b := toc.NewBatcher(context.Background(), src, 5)

	errs := 0
	for r := range b.Out() {
		if _, err := r.Unpack(); err != nil {
			errs++
		}
	}

	if errs != 3 {
		t.Fatalf("got %d errors, want 3", errs)
	}

	stats := b.Stats()
	if stats.Forwarded != 3 {
		t.Errorf("Forwarded = %d, want 3", stats.Forwarded)
	}
	if stats.Emitted != 0 {
		t.Errorf("Emitted = %d, want 0", stats.Emitted)
	}

	b.Wait()
}

func TestBatcherCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	src := make(chan rslt.Result[int])

	b := toc.NewBatcher(ctx, src, 5)

	// Send a couple items, then cancel.
	src <- rslt.Ok(1)
	src <- rslt.Ok(2)
	cancel()

	// Send more after cancel — batcher should drain.
	go func() {
		for i := 3; i <= 5; i++ {
			src <- rslt.Ok(i)
		}
		close(src)
	}()

	for range b.Out() {
	}

	err := b.Wait()
	if err == nil {
		t.Fatal("Wait should return ctx error after cancel")
	}

	stats := b.Stats()
	if stats.Received != 5 {
		t.Errorf("Received = %d, want 5", stats.Received)
	}

	// Invariant must hold.
	if stats.Received != stats.Emitted+stats.Forwarded+stats.Dropped {
		t.Errorf("invariant violated: %d != %d + %d + %d",
			stats.Received, stats.Emitted, stats.Forwarded, stats.Dropped)
	}
}

func TestBatcherSingleItem(t *testing.T) {
	src := feedResults(rslt.Ok(1), rslt.Ok(2), rslt.Ok(3))

	b := toc.NewBatcher(context.Background(), src, 1)

	var batches [][]int
	for r := range b.Out() {
		v, err := r.Unpack()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		batches = append(batches, v)
	}

	if len(batches) != 3 {
		t.Fatalf("got %d batches, want 3", len(batches))
	}
	for i, batch := range batches {
		if len(batch) != 1 {
			t.Errorf("batch[%d] has %d items, want 1", i, len(batch))
		}
	}

	b.Wait()
}

func TestBatcherStatsInvariant(t *testing.T) {
	testErr := errors.New("err")
	src := feedResults(
		rslt.Ok(1), rslt.Ok(2), rslt.Err[int](testErr),
		rslt.Ok(3), rslt.Ok(4), rslt.Ok(5),
	)

	b := toc.NewBatcher(context.Background(), src, 3)

	for range b.Out() {
	}
	b.Wait()

	stats := b.Stats()
	if stats.Received != 6 {
		t.Errorf("Received = %d, want 6", stats.Received)
	}
	if stats.Forwarded != 1 {
		t.Errorf("Forwarded = %d, want 1", stats.Forwarded)
	}
	if stats.Emitted != 5 {
		t.Errorf("Emitted = %d, want 5", stats.Emitted)
	}

	if stats.Received != stats.Emitted+stats.Forwarded+stats.Dropped {
		t.Errorf("invariant violated: %d != %d + %d + %d",
			stats.Received, stats.Emitted, stats.Forwarded, stats.Dropped)
	}
}

func TestBatcherPanics(t *testing.T) {
	t.Run("n<=0", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic for n <= 0")
			}
		}()
		ch := make(chan rslt.Result[int])
		close(ch)
		toc.NewBatcher(context.Background(), ch, 0)
	})

	t.Run("nil src", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic for nil src")
			}
		}()
		toc.NewBatcher[int](context.Background(), nil, 5)
	})

	t.Run("nil ctx", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic for nil ctx")
			}
		}()
		ch := make(chan rslt.Result[int])
		close(ch)
		toc.NewBatcher(nil, ch, 5) //nolint:staticcheck // intentional nil ctx test
	})
}

func TestBatcherSourceDrain(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	src := make(chan rslt.Result[int])

	b := toc.NewBatcher(ctx, src, 5)

	// Cancel immediately.
	cancel()

	// Feed items — batcher must drain even after cancel.
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 5; i++ {
			src <- rslt.Ok(i)
		}
		close(src)
	}()

	for range b.Out() {
	}
	b.Wait()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("sender goroutine stuck — batcher didn't drain src")
	}
}

func TestBatcherCancelDuringEmit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	src := make(chan rslt.Result[int])

	b := toc.NewBatcher(ctx, src, 2)

	// Send 2 items to fill a batch — emit will block because nobody reads Out().
	src <- rslt.Ok(1)
	src <- rslt.Ok(2)

	// Give batcher time to attempt emit (blocks on out send).
	time.Sleep(20 * time.Millisecond)

	// Cancel while emit is blocked.
	cancel()

	// Send more items — batcher must drain even if emit was interrupted.
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 3; i <= 5; i++ {
			src <- rslt.Ok(i)
		}
		close(src)
	}()

	// Drain output (may get nothing if cancel won the race).
	for range b.Out() {
	}

	err := b.Wait()
	if err == nil {
		t.Fatal("Wait should return ctx error after cancel")
	}

	// Sender must not be stuck.
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("sender stuck — batcher didn't drain src after cancel during emit")
	}

	stats := b.Stats()
	if stats.Received != 5 {
		t.Errorf("Received = %d, want 5", stats.Received)
	}
	if stats.Received != stats.Emitted+stats.Forwarded+stats.Dropped {
		t.Errorf("invariant violated: %d != %d + %d + %d",
			stats.Received, stats.Emitted, stats.Forwarded, stats.Dropped)
	}
}

func TestBatcherCancelDuringErrorForward(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	src := make(chan rslt.Result[int])

	b := toc.NewBatcher(ctx, src, 5)

	// Send an error — forwarding will block because nobody reads Out().
	src <- rslt.Err[int](errors.New("test error"))

	// Give batcher time to attempt forward (blocks on out send).
	time.Sleep(20 * time.Millisecond)

	// Cancel while forward is blocked.
	cancel()

	// Send more — batcher must drain.
	done := make(chan struct{})
	go func() {
		defer close(done)
		src <- rslt.Ok(1)
		close(src)
	}()

	for range b.Out() {
	}

	err := b.Wait()
	if err == nil {
		t.Fatal("Wait should return ctx error after cancel")
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("sender stuck — batcher didn't drain src after cancel during error forward")
	}

	stats := b.Stats()
	if stats.Received != 2 {
		t.Errorf("Received = %d, want 2", stats.Received)
	}
	if stats.Received != stats.Emitted+stats.Forwarded+stats.Dropped {
		t.Errorf("invariant violated: %d != %d + %d + %d",
			stats.Received, stats.Emitted, stats.Forwarded, stats.Dropped)
	}
}

func TestBatcherUpstreamDeadlockOnCancel(t *testing.T) {
	// Start → Batcher pipeline. Never drain batcher output. Cancel context.
	// Assert upstream stage completes.
	ctx, cancel := context.WithCancel(context.Background())

	// identityFn passes through unchanged.
	identityFn := func(_ context.Context, n int) (int, error) { return n, nil }

	stage := toc.Start(ctx, identityFn, toc.Options[int]{Capacity: 5})
	b := toc.NewBatcher(ctx, stage.Out(), 2)

	// Submit items to fill the pipeline.
	for i := 0; i < 5; i++ {
		stage.Submit(ctx, i)
	}
	stage.CloseInput()

	// Don't drain b.Out() — cancel instead.
	time.Sleep(20 * time.Millisecond)
	cancel()

	// Drain batcher output now (after cancel) — should complete.
	for range b.Out() {
	}

	// Both must complete without deadlock.
	done := make(chan struct{})
	go func() {
		defer close(done)
		b.Wait()
		stage.Wait()
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("pipeline deadlocked — blocked upstream or batcher didn't drain")
	}
}

func TestBatcherCancelDuringFinalFlush(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	src := make(chan rslt.Result[int])

	b := toc.NewBatcher(ctx, src, 3)

	// Send 2 items (partial batch), then close src — triggers final flush.
	src <- rslt.Ok(1)
	src <- rslt.Ok(2)
	close(src)

	// Give batcher time to reach final flush (blocks on out send).
	time.Sleep(20 * time.Millisecond)

	// Cancel while final flush is blocked.
	cancel()

	// Drain output.
	for range b.Out() {
	}

	// Wait must report cancellation — items were lost.
	if err := b.Wait(); err == nil {
		t.Fatal("Wait should return ctx error after cancel interrupts final flush")
	}

	stats := b.Stats()
	if stats.Received != 2 {
		t.Errorf("Received = %d, want 2", stats.Received)
	}
	if stats.Received != stats.Emitted+stats.Forwarded+stats.Dropped {
		t.Errorf("invariant violated: %d != %d + %d + %d",
			stats.Received, stats.Emitted, stats.Forwarded, stats.Dropped)
	}
}

func TestBatcherRace(t *testing.T) {
	// Exercise under -race with mixed items.
	src := make(chan rslt.Result[int], 100)
	for i := 0; i < 100; i++ {
		if i%7 == 0 {
			src <- rslt.Err[int](errors.New("err"))
		} else {
			src <- rslt.Ok(i)
		}
	}
	close(src)

	b := toc.NewBatcher(context.Background(), src, 10)

	count := 0
	for range b.Out() {
		count++
	}

	b.Wait()

	stats := b.Stats()
	if stats.Received != 100 {
		t.Errorf("Received = %d, want 100", stats.Received)
	}
}
