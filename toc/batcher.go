package toc

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/binaryphile/fluentfp/rslt"
)

// BatcherStats holds metrics for a [Batcher].
//
// Invariant (after Wait): Received = Emitted + Forwarded + Dropped.
type BatcherStats struct {
	Received          int64         // individual items consumed from src
	Emitted           int64         // individual Ok items included in emitted batches
	Forwarded         int64         // Err items forwarded downstream
	Dropped           int64         // items lost to shutdown/cancel (includes partial batch items)
	BufferedDepth     int64         // current items in partial batch accumulator
	BatchCount        int64         // number of batch results emitted
	OutputBlockedTime time.Duration // cumulative time blocked sending to out
}

// ToStats converts BatcherStats to [Stats] for use with the analyze package.
func (s BatcherStats) ToStats() Stats {
	return Stats{
		Received:          s.Received,
		Submitted:         s.Emitted,
		Completed:         s.BatchCount,
		Forwarded:         s.Forwarded,
		Dropped:           s.Dropped,
		BufferedDepth:     s.BufferedDepth,
		OutputBlockedTime: s.OutputBlockedTime,
	}
}

// Batcher accumulates up to n Ok items from an upstream Result channel
// into batches, emitting each batch as rslt.Result[[]T]. Errors act as
// batch boundaries: the partial batch is flushed, then the error is
// forwarded, and a fresh accumulator starts.
//
// Created by [NewBatcher]. The zero value is not usable.
type Batcher[T any] struct {
	out  chan rslt.Result[[]T]
	done chan struct{}

	received          atomic.Int64
	emitted           atomic.Int64
	forwarded         atomic.Int64
	dropped           atomic.Int64
	bufferedDepth     atomic.Int64
	batchCount        atomic.Int64
	outputBlockedNs   atomic.Int64
	ctxErr            error // latched after goroutine exits, before close(done)
}

// NewBatcher creates a Batcher that reads from src, accumulates up to n
// Ok values per batch, and emits batches on Out(). Errors from src act
// as batch boundaries: flush partial batch (if non-empty), forward the
// error, start fresh.
//
// The Batcher drains src to completion (source ownership rule), provided
// the consumer drains [Batcher.Out] or ctx is canceled, and src eventually
// closes. Cancellation unblocks output sends but does not force-close src.
// After ctx cancellation, it switches to discard mode: continues reading
// src but discards all items without flushing partial batches. Batch
// emission and error forwarding are best-effort during shutdown — cancel
// may race with a successful send, so output may still appear after
// cancellation. All drops are reflected in stats. If the consumer stops
// reading Out and ctx is never canceled, the Batcher blocks on output
// delivery and cannot drain src.
//
// Panics if n <= 0, src is nil, or ctx is nil.
func NewBatcher[T any](
	ctx context.Context,
	src <-chan rslt.Result[T],
	n int,
) *Batcher[T] {
	if n <= 0 {
		panic("toc.NewBatcher: n must be positive")
	}
	if src == nil {
		panic("toc.NewBatcher: src must not be nil")
	}
	if ctx == nil {
		panic("toc.NewBatcher: ctx must not be nil")
	}

	b := &Batcher[T]{
		out:  make(chan rslt.Result[[]T]),
		done: make(chan struct{}),
	}

	go b.run(ctx, src, n)

	return b
}

// Out returns the receive-only output channel. It closes after the
// source channel closes and all batches have been emitted.
//
// Callers MUST drain Out to completion. If the consumer stops reading,
// the Batcher blocks and cannot drain its source, potentially causing
// upstream deadlocks.
func (b *Batcher[T]) Out() <-chan rslt.Result[[]T] { return b.out }

// Wait blocks until the Batcher goroutine exits. Returns ctx.Err() if
// context cancellation caused items to be dropped (discard mode or
// interrupted output send), nil otherwise. The Batcher has no fn, so
// there is no fail-fast error.
func (b *Batcher[T]) Wait() error {
	<-b.done

	return b.ctxErr
}

// Stats returns approximate metrics. See [BatcherStats] for caveats.
// Stats are only reliable as final values after [Batcher.Wait] returns.
func (b *Batcher[T]) Stats() BatcherStats {
	return BatcherStats{
		Received:          b.received.Load(),
		Emitted:           b.emitted.Load(),
		Forwarded:         b.forwarded.Load(),
		Dropped:           b.dropped.Load(),
		BufferedDepth:     b.bufferedDepth.Load(),
		BatchCount:        b.batchCount.Load(),
		OutputBlockedTime: time.Duration(b.outputBlockedNs.Load()),
	}
}

// run is the Batcher's single goroutine.
func (b *Batcher[T]) run(ctx context.Context, src <-chan rslt.Result[T], n int) {
	defer close(b.done)
	defer close(b.out)

	buf := make([]T, 0, n)

	// drainAndDiscard enters discard mode: drops current buf contents,
	// then drains src to completion (source ownership rule).
	drainAndDiscard := func() {
		if len(buf) > 0 {
			b.dropped.Add(int64(len(buf)))
			b.bufferedDepth.Store(0)
			buf = buf[:0]
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
		buf = make([]T, 0, n)

		if !trySend(rslt.Ok(batch)) {
			// Batch was not delivered — count items as dropped.
			b.dropped.Add(int64(len(batch)))
			b.bufferedDepth.Store(0)

			return false
		}

		b.emitted.Add(int64(len(batch)))
		b.batchCount.Add(1)
		b.bufferedDepth.Store(0)

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
			buf = append(buf, v)
			b.bufferedDepth.Add(1)

			if len(buf) == n {
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
