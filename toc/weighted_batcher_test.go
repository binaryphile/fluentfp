package toc_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/binaryphile/fluentfp/rslt"
	"github.com/binaryphile/fluentfp/toc"
)

// weightOf returns the item itself as the weight (for int items).
func weightOf(n int) int { return n }

// constantWeight returns a weight function that always returns w.
func constantWeight(w int) func(int) int {
	return func(_ int) int { return w }
}

func TestWeightedBatcherHappyPath(t *testing.T) {
	// Items with weight = value. Threshold 10.
	// 1+2+3+4 = 10 → flush, 5+6 = 11 → flush (at 5, weight=5 < 10; at 6, weight=11 >= 10)
	// Wait — 5+6 flushed at threshold. No remainder.
	src := make(chan rslt.Result[int], 6)
	for i := 1; i <= 6; i++ {
		src <- rslt.Ok(i)
	}
	close(src)

	b := toc.NewWeightedBatcher(context.Background(), src, 10, weightOf)

	var batches [][]int
	for r := range b.Out() {
		v, err := r.Unpack()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		batches = append(batches, v)
	}

	if len(batches) != 2 {
		t.Fatalf("got %d batches, want 2", len(batches))
	}
	// First batch: 1+2+3+4 = 10
	if len(batches[0]) != 4 {
		t.Errorf("batch[0] has %d items, want 4", len(batches[0]))
	}
	// Second batch: 5+6 = 11
	if len(batches[1]) != 2 {
		t.Errorf("batch[1] has %d items, want 2", len(batches[1]))
	}

	if err := b.Wait(); err != nil {
		t.Fatalf("Wait: %v", err)
	}

	stats := b.Stats()
	if stats.Received != 6 {
		t.Errorf("Received = %d, want 6", stats.Received)
	}
	if stats.Emitted != 6 {
		t.Errorf("Emitted = %d, want 6", stats.Emitted)
	}
	if stats.BatchCount != 2 {
		t.Errorf("BatchCount = %d, want 2", stats.BatchCount)
	}
}

func TestWeightedBatcherVariableWeights(t *testing.T) {
	// Items: weight 1, 1, 1, 8 → threshold 5.
	// 1+1+1 = 3 < 5, then +8 = 11 >= 5 → flush all 4.
	src := feedResults(rslt.Ok(1), rslt.Ok(1), rslt.Ok(1), rslt.Ok(8))

	b := toc.NewWeightedBatcher(context.Background(), src, 5, weightOf)

	var batches [][]int
	for r := range b.Out() {
		v, err := r.Unpack()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		batches = append(batches, v)
	}

	// All 4 items in one batch (threshold crossed on 4th item).
	if len(batches) != 1 {
		t.Fatalf("got %d batches, want 1", len(batches))
	}
	if len(batches[0]) != 4 {
		t.Errorf("batch[0] has %d items, want 4", len(batches[0]))
	}

	b.Wait()
}

func TestWeightedBatcherPartialFlush(t *testing.T) {
	// Items never reach threshold — partial batch flushed on src close.
	src := feedResults(rslt.Ok(1), rslt.Ok(2))

	b := toc.NewWeightedBatcher(context.Background(), src, 100, weightOf)

	var batches [][]int
	for r := range b.Out() {
		v, err := r.Unpack()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		batches = append(batches, v)
	}

	if len(batches) != 1 {
		t.Fatalf("got %d batches, want 1", len(batches))
	}
	if len(batches[0]) != 2 {
		t.Errorf("batch[0] has %d items, want 2", len(batches[0]))
	}

	b.Wait()
}

func TestWeightedBatcherZeroWeightItems(t *testing.T) {
	// All items have weight 0. Item-count fallback flushes at threshold.
	// 5 items, threshold 3 → flush at 3 items, then partial flush of 2.
	src := feedResults(rslt.Ok(1), rslt.Ok(2), rslt.Ok(3), rslt.Ok(4), rslt.Ok(5))

	b := toc.NewWeightedBatcher(context.Background(), src, 3, constantWeight(0))

	var batches [][]int
	for r := range b.Out() {
		v, err := r.Unpack()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		batches = append(batches, v)
	}

	// 3 items flushed by item-count fallback, then 2 partial on close.
	if len(batches) != 2 {
		t.Fatalf("got %d batches, want 2", len(batches))
	}
	if len(batches[0]) != 3 {
		t.Errorf("batch[0] has %d items, want 3", len(batches[0]))
	}
	if len(batches[1]) != 2 {
		t.Errorf("batch[1] has %d items, want 2", len(batches[1]))
	}

	b.Wait()
}

func TestWeightedBatcherItemCountFallback(t *testing.T) {
	// Items with weight 1, threshold 5. Weight reaches 5 at item 5.
	// But item count also reaches 5 at item 5. Test that 10 items with
	// weight 0 still flush at 5 items (item-count fallback).
	src := make(chan rslt.Result[int], 10)
	for i := 0; i < 10; i++ {
		src <- rslt.Ok(i)
	}
	close(src)

	b := toc.NewWeightedBatcher(context.Background(), src, 5, constantWeight(0))

	var batches [][]int
	for r := range b.Out() {
		v, err := r.Unpack()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		batches = append(batches, v)
	}

	// 10 items / 5 = 2 batches (item-count fallback).
	if len(batches) != 2 {
		t.Fatalf("got %d batches, want 2", len(batches))
	}
	if len(batches[0]) != 5 {
		t.Errorf("batch[0] has %d items, want 5", len(batches[0]))
	}
	if len(batches[1]) != 5 {
		t.Errorf("batch[1] has %d items, want 5", len(batches[1]))
	}

	b.Wait()
}


func TestWeightedBatcherErrorAsBoundary(t *testing.T) {
	testErr := errors.New("mid-error")
	// Items: Ok(3), Ok(3), Err, Ok(3), Ok(3). Threshold 5.
	// 3+3=6 >= 5 → flush [3,3]. Err → forward. 3+3=6 >= 5 → flush [3,3].
	src := feedResults(
		rslt.Ok(3),
		rslt.Ok(3),
		rslt.Err[int](testErr),
		rslt.Ok(3),
		rslt.Ok(3),
	)

	b := toc.NewWeightedBatcher(context.Background(), src, 5, weightOf)

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

	// [3,3] was flushed at threshold before the error arrived.
	// Then error forwarded. Then [3,3] flushed at threshold.
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

func TestWeightedBatcherErrorFlushesPartial(t *testing.T) {
	testErr := errors.New("mid-error")
	// Items: Ok(1), Ok(1), Err, Ok(1). Threshold 10.
	// 1+1=2 < 10, then Err → flush partial [1,1], forward error. 1 < 10, flush on close.
	src := feedResults(
		rslt.Ok(1),
		rslt.Ok(1),
		rslt.Err[int](testErr),
		rslt.Ok(1),
	)

	b := toc.NewWeightedBatcher(context.Background(), src, 10, weightOf)

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

	if len(items) != 3 {
		t.Fatalf("got %d items, want 3", len(items))
	}
	if items[0].isErr || len(items[0].batch) != 2 {
		t.Errorf("item[0]: want batch of 2, got %+v", items[0])
	}
	if !items[1].isErr {
		t.Errorf("item[1]: want error")
	}
	if items[2].isErr || len(items[2].batch) != 1 {
		t.Errorf("item[2]: want batch of 1, got %+v", items[2])
	}

	b.Wait()
}

func TestWeightedBatcherEncounterOrder(t *testing.T) {
	testErr := errors.New("err")
	src := feedResults(
		rslt.Ok(10),
		rslt.Ok(20),
		rslt.Err[int](testErr),
		rslt.Ok(30),
	)

	b := toc.NewWeightedBatcher(context.Background(), src, 100, weightOf)

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

func TestWeightedBatcherEmpty(t *testing.T) {
	src := make(chan rslt.Result[int])
	close(src)

	b := toc.NewWeightedBatcher(context.Background(), src, 5, weightOf)

	count := 0
	for range b.Out() {
		count++
	}

	if count != 0 {
		t.Fatalf("got %d items, want 0", count)
	}

	b.Wait()
}

func TestWeightedBatcherAllErrors(t *testing.T) {
	src := feedResults(
		rslt.Err[int](errors.New("e1")),
		rslt.Err[int](errors.New("e2")),
		rslt.Err[int](errors.New("e3")),
	)

	b := toc.NewWeightedBatcher(context.Background(), src, 5, weightOf)

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

func TestWeightedBatcherCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	src := make(chan rslt.Result[int])

	b := toc.NewWeightedBatcher(ctx, src, 100, constantWeight(1))

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

func TestWeightedBatcherSingleItemBatch(t *testing.T) {
	// Each item has weight >= threshold → each item is its own batch.
	src := feedResults(rslt.Ok(10), rslt.Ok(20), rslt.Ok(30))

	b := toc.NewWeightedBatcher(context.Background(), src, 5, weightOf)

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

func TestWeightedBatcherStatsInvariant(t *testing.T) {
	testErr := errors.New("err")
	src := feedResults(
		rslt.Ok(3), rslt.Ok(3), rslt.Err[int](testErr),
		rslt.Ok(2), rslt.Ok(2), rslt.Ok(2),
	)

	b := toc.NewWeightedBatcher(context.Background(), src, 5, weightOf)

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

	if stats.Received != stats.Emitted+stats.Forwarded+stats.Dropped {
		t.Errorf("invariant violated: %d != %d + %d + %d",
			stats.Received, stats.Emitted, stats.Forwarded, stats.Dropped)
	}
}

func TestWeightedBatcherBufferedWeightTracking(t *testing.T) {
	src := make(chan rslt.Result[int], 3)
	src <- rslt.Ok(3)
	src <- rslt.Ok(4)
	src <- rslt.Ok(5)
	close(src)

	// Threshold 100 — nothing flushes until close.
	b := toc.NewWeightedBatcher(context.Background(), src, 100, weightOf)

	for range b.Out() {
	}
	b.Wait()

	stats := b.Stats()
	// After Wait, BufferedWeight and BufferedDepth should be 0 (batch was flushed).
	if stats.BufferedWeight != 0 {
		t.Errorf("BufferedWeight = %d, want 0 after flush", stats.BufferedWeight)
	}
	if stats.BufferedDepth != 0 {
		t.Errorf("BufferedDepth = %d, want 0 after flush", stats.BufferedDepth)
	}
	if stats.Emitted != 3 {
		t.Errorf("Emitted = %d, want 3", stats.Emitted)
	}
}

func TestWeightedBatcherPanics(t *testing.T) {
	t.Run("threshold<=0", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic for threshold <= 0")
			}
		}()
		ch := make(chan rslt.Result[int])
		close(ch)
		toc.NewWeightedBatcher(context.Background(), ch, 0, weightOf)
	})

	t.Run("nil weightFn", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic for nil weightFn")
			}
		}()
		ch := make(chan rslt.Result[int])
		close(ch)
		toc.NewWeightedBatcher[int](context.Background(), ch, 5, nil)
	})

	t.Run("nil src", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic for nil src")
			}
		}()
		toc.NewWeightedBatcher[int](context.Background(), nil, 5, weightOf)
	})

	t.Run("nil ctx", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic for nil ctx")
			}
		}()
		ch := make(chan rslt.Result[int])
		close(ch)
		toc.NewWeightedBatcher(nil, ch, 5, weightOf) //nolint:staticcheck // intentional nil ctx test
	})
}

func TestWeightedBatcherSourceDrain(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	src := make(chan rslt.Result[int])

	b := toc.NewWeightedBatcher(ctx, src, 100, constantWeight(1))

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

func TestWeightedBatcherCancelDuringEmit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	src := make(chan rslt.Result[int])

	b := toc.NewWeightedBatcher(ctx, src, 5, constantWeight(3))

	// Send 2 items (weight 6 >= 5) — emit will block because nobody reads Out().
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

func TestWeightedBatcherCancelDuringFinalFlush(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	src := make(chan rslt.Result[int])

	b := toc.NewWeightedBatcher(ctx, src, 100, constantWeight(1))

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

func TestWeightedBatcherRace(t *testing.T) {
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

	b := toc.NewWeightedBatcher(context.Background(), src, 50, weightOf)

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
