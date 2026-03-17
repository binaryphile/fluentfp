package toc

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/binaryphile/fluentfp/rslt"
)

// WeightedBatcherStats holds metrics for a [WeightedBatcher].
//
// Invariant (after Wait): Received = Emitted + Forwarded + Dropped.
type WeightedBatcherStats struct {
	Received          int64         // individual items consumed from src
	Emitted           int64         // individual Ok items included in emitted batches
	Forwarded         int64         // Err items forwarded downstream
	Dropped           int64         // items lost to shutdown/cancel (includes partial batch items)
	BufferedDepth     int64         // current items in partial batch accumulator
	BufferedWeight    int64         // current accumulated weight in partial batch
	BatchCount        int64         // number of batch results emitted
	OutputBlockedTime time.Duration // cumulative time blocked sending to out
}

// WeightedBatcher accumulates Ok items from an upstream Result channel
// into batches, flushing when accumulated weight OR item count reaches
// the threshold. Each item's weight is determined by weightFn. Errors
// act as batch boundaries: the partial batch is flushed, then the error
// is forwarded, and a fresh accumulator starts.
//
// Created by [NewWeightedBatcher]. The zero value is not usable.
type WeightedBatcher[T any] struct {
	out  chan rslt.Result[[]T]
	done chan struct{}

	received        atomic.Int64
	emitted         atomic.Int64
	forwarded       atomic.Int64
	dropped         atomic.Int64
	bufferedDepth   atomic.Int64
	bufferedWeight  atomic.Int64
	batchCount      atomic.Int64
	outputBlockedNs atomic.Int64
	ctxErr          error // latched after goroutine exits, before close(done)
}

// NewWeightedBatcher creates a WeightedBatcher that reads from src,
// accumulates Ok values, and emits batches on Out(). A batch is flushed
// when either accumulated weight (per weightFn) reaches threshold or
// item count reaches threshold — whichever comes first. The item-count
// fallback prevents unbounded accumulation of zero/low-weight items.
// Errors from src act as batch boundaries: flush partial batch (if
// non-empty), forward the error, start fresh.
//
// weightFn must return a non-negative weight for each item. A panic
// occurs if weightFn returns a negative value.
//
// The WeightedBatcher drains src to completion (source ownership rule),
// provided the consumer drains [WeightedBatcher.Out] or ctx is canceled,
// and src eventually closes. Cancellation unblocks output sends but does
// not force-close src. After ctx cancellation, it switches to discard
// mode: continues reading src but discards all items without flushing
// partial batches. Batch emission and error forwarding are best-effort
// during shutdown — cancel may race with a successful send, so output
// may still appear after cancellation. All drops are reflected in stats.
// If the consumer stops reading Out and ctx is never canceled, the
// WeightedBatcher blocks on output delivery and cannot drain src.
//
// Panics if threshold <= 0, weightFn is nil, src is nil, or ctx is nil.
func NewWeightedBatcher[T any](
	ctx context.Context,
	src <-chan rslt.Result[T],
	threshold int,
	weightFn func(T) int,
) *WeightedBatcher[T] {
	if threshold <= 0 {
		panic("toc.NewWeightedBatcher: threshold must be positive")
	}
	if weightFn == nil {
		panic("toc.NewWeightedBatcher: weightFn must not be nil")
	}
	if src == nil {
		panic("toc.NewWeightedBatcher: src must not be nil")
	}
	if ctx == nil {
		panic("toc.NewWeightedBatcher: ctx must not be nil")
	}

	b := &WeightedBatcher[T]{
		out:  make(chan rslt.Result[[]T]),
		done: make(chan struct{}),
	}

	go b.run(ctx, src, threshold, weightFn)

	return b
}

// Out returns the receive-only output channel. It closes after the
// source channel closes and all batches have been emitted.
//
// Callers MUST drain Out to completion. If the consumer stops reading,
// the WeightedBatcher blocks and cannot drain its source, potentially
// causing upstream deadlocks.
func (b *WeightedBatcher[T]) Out() <-chan rslt.Result[[]T] { return b.out }

// Wait blocks until the WeightedBatcher goroutine exits. Returns
// ctx.Err() if context cancellation caused items to be dropped (discard
// mode or interrupted output send), nil otherwise. The WeightedBatcher
// has no fn, so there is no fail-fast error.
func (b *WeightedBatcher[T]) Wait() error {
	<-b.done

	return b.ctxErr
}

// Stats returns approximate metrics. See [WeightedBatcherStats] for
// caveats. Stats are only reliable as final values after
// [WeightedBatcher.Wait] returns.
func (b *WeightedBatcher[T]) Stats() WeightedBatcherStats {
	return WeightedBatcherStats{
		Received:          b.received.Load(),
		Emitted:           b.emitted.Load(),
		Forwarded:         b.forwarded.Load(),
		Dropped:           b.dropped.Load(),
		BufferedDepth:     b.bufferedDepth.Load(),
		BufferedWeight:    b.bufferedWeight.Load(),
		BatchCount:        b.batchCount.Load(),
		OutputBlockedTime: time.Duration(b.outputBlockedNs.Load()),
	}
}

// run is the WeightedBatcher's single goroutine.
func (b *WeightedBatcher[T]) run(
	ctx context.Context,
	src <-chan rslt.Result[T],
	threshold int,
	weightFn func(T) int,
) {
	defer close(b.done)
	defer close(b.out)

	var buf []T
	var weight int

	// drainAndDiscard enters discard mode: drops current buf contents,
	// then drains src to completion (source ownership rule).
	drainAndDiscard := func() {
		if len(buf) > 0 {
			b.dropped.Add(int64(len(buf)))
			b.bufferedDepth.Store(0)
			b.bufferedWeight.Store(0)
			buf = nil
			weight = 0
		}

		for range src {
			b.received.Add(1)
			b.dropped.Add(1)
		}

		b.ctxErr = ctx.Err()
	}

	// trySend attempts to send r to out, respecting cancellation.
	// Returns true if sent, false if canceled (caller should enter discard mode).
	trySend := func(r rslt.Result[[]T]) bool {
		outStart := time.Now()

		select {
		case b.out <- r:
			b.outputBlockedNs.Add(int64(time.Since(outStart)))

			return true
		case <-ctx.Done():
			b.outputBlockedNs.Add(int64(time.Since(outStart)))

			return false
		}
	}

	// emit sends a batch to out and resets the accumulator.
	// Returns false if canceled during send (caller enters discard mode).
	emit := func() bool {
		batch := buf
		buf = nil
		weight = 0

		if !trySend(rslt.Ok(batch)) {
			// Batch was not delivered — count items as dropped.
			b.dropped.Add(int64(len(batch)))
			b.bufferedDepth.Store(0)
			b.bufferedWeight.Store(0)

			return false
		}

		b.emitted.Add(int64(len(batch)))
		b.batchCount.Add(1)
		b.bufferedDepth.Store(0)
		b.bufferedWeight.Store(0)

		return true
	}

	for r := range src {
		b.received.Add(1)

		// Check for cancellation — switch to discard mode.
		select {
		case <-ctx.Done():
			b.dropped.Add(1) // the current item
			drainAndDiscard()

			return
		default:
		}

		if v, err := r.Unpack(); err != nil {
			// Error = batch boundary: flush partial, forward error.
			if len(buf) > 0 {
				if !emit() {
					b.dropped.Add(1) // the current error item
					drainAndDiscard()

					return
				}
			}

			if !trySend(rslt.Err[[]T](err)) {
				b.dropped.Add(1) // the error item itself
				drainAndDiscard()

				return
			}

			b.forwarded.Add(1)
		} else {
			w := weightFn(v)
			if w < 0 {
				panic("toc.NewWeightedBatcher: weightFn returned negative weight")
			}

			buf = append(buf, v)
			weight += w
			b.bufferedDepth.Add(1)
			b.bufferedWeight.Add(int64(w))

			if weight >= threshold || len(buf) >= threshold {
				if !emit() {
					drainAndDiscard()

					return
				}
			}
		}
	}

	// Flush remaining partial batch on normal close.
	// If cancel interrupts the flush, latch the error so Wait() reports it.
	if len(buf) > 0 {
		if !emit() {
			b.ctxErr = ctx.Err()
		}
	}
}
