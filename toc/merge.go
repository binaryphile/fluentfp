package toc

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/binaryphile/fluentfp/rslt"
)

// MergeStats holds metrics for a [Merge].
//
// Fields are derived from per-source atomic counters, so a mid-flight Stats
// value is NOT a consistent snapshot — cross-metric invariants
// (e.g., Received == Forwarded + Dropped) may not hold because an in-flight
// item has been counted as received but not yet as forwarded or dropped.
// Invariants are guaranteed only after [Merge.Wait] returns.
//
// Per-metric aggregates are coherent within a single Stats() call even
// mid-flight: Received == sum(SourceReceived), Forwarded == sum(SourceForwarded),
// Dropped == sum(SourceDropped). This holds because aggregates are computed
// from the same copied values.
//
// Invariant (after Wait): Received = Forwarded + Dropped.
// Per-source invariant (after Wait): SourceReceived[i] = SourceForwarded[i] + SourceDropped[i].
//
// SourceReceived[i] corresponds to sources[i] as passed to [NewMerge].
// The index mapping is stable and matches construction order.
type MergeStats struct {
	Received  int64 // total items consumed from all sources
	Forwarded int64 // items sent to output
	Dropped   int64 // items discarded during cancel

	// Per-source stats. Len == N (number of sources).
	// Index i corresponds to sources[i] as passed to NewMerge.
	SourceReceived  []int64
	SourceForwarded []int64
	SourceDropped   []int64
}

// Merge is a nondeterministic interleaving fan-in from N sources.
//
// One goroutine per source forwards items to a shared unbuffered output
// channel. Go runtime scheduler determines send order — no cross-source
// ordering guarantee, no fairness guarantee, no provenance tracking.
// Per-source order IS preserved: items from each individual source appear
// in the merged output in the same order they were received from that
// source (follows from one goroutine per source with sequential
// receive/send).
//
// Merge is NOT the inverse of [Tee]. Tee broadcasts identical items to
// all branches. Merge interleaves distinct items from independent sources.
// Tee → ... → Merge does not restore original ordering, does not correlate
// outputs from sibling branches, and does not pair items across sources.
//
// Created by [NewMerge]. The zero value is not usable.
// A Merge must not be copied after first use.
type Merge[T any] struct {
	out  chan rslt.Result[T]
	done chan struct{}

	// Per-source atomics. Aggregates derived in Stats().
	sourceReceived  []atomic.Int64
	sourceForwarded []atomic.Int64
	sourceDropped   []atomic.Int64

	cancelOnce sync.Once
	ctxErr     error
}

// NewMerge creates a Merge that reads from each source and forwards all
// items to a single unbuffered output channel. Items are interleaved
// nondeterministically — Go runtime scheduler determines send order.
//
// Each source is drained to completion by its own goroutine (source
// ownership rule). Sources may close at different times — early closure
// of one source does not affect others. All sources must be finite and
// must eventually close (including on cancellation paths). If a source
// never closes, the corresponding goroutine blocks indefinitely and
// [Merge.Wait] hangs.
//
// Cancellation is advisory, not a hard stop. On ctx cancellation, each
// source goroutine enters discard mode at its next cancellation
// checkpoint: stops forwarding but continues draining its source to
// completion. Two cancellation checkpoints per iteration: a non-blocking
// pre-send check and a blocking send-select. This bounds post-cancel
// forwarding to at most 1 item per source goroutine that has already
// passed the pre-send checkpoint. Cancel-aware sends ensure goroutines
// are not blocked on output when downstream stops reading. If a goroutine
// is blocked waiting on its source (for r := range src) when ctx cancels,
// it does not observe cancellation until the source produces an item or
// closes.
//
// [Merge.Wait] returns only after all source goroutines exit — which
// requires all sources to close. Cancellation alone does not guarantee
// prompt return. After observing cancellation, each goroutine drains and
// discards remaining items from its source until that source closes.
//
// [Merge.Out] is closed before done is closed, so Out() is guaranteed
// closed before Wait() returns. Callers can safely range Out() and then
// call Wait().
//
// Goroutine lifecycle: constructor launches N source goroutines + 1 closer
// goroutine. Each source goroutine drains its source and sends to output.
// The closer goroutine waits on a WaitGroup for all source goroutines,
// closes output, and closes done.
//
// Each source channel must be distinct and exclusively owned by the Merge.
// Passing the same channel twice creates two goroutines racing on one
// source — per-source ordering and stats become meaningless. The
// constructor does not check for duplicates.
//
// Panics if len(sources) == 0, ctx is nil, or any source is nil.
func NewMerge[T any](ctx context.Context, sources ...<-chan rslt.Result[T]) *Merge[T] {
	if len(sources) == 0 {
		panic("toc.NewMerge: at least one source required")
	}
	if ctx == nil {
		panic("toc.NewMerge: ctx must not be nil")
	}
	for _, src := range sources {
		if src == nil {
			panic("toc.NewMerge: source must not be nil")
		}
	}

	n := len(sources)
	m := &Merge[T]{
		out:             make(chan rslt.Result[T]),
		done:            make(chan struct{}),
		sourceReceived:  make([]atomic.Int64, n),
		sourceForwarded: make([]atomic.Int64, n),
		sourceDropped:   make([]atomic.Int64, n),
	}

	var wg sync.WaitGroup
	wg.Add(n)

	for i, src := range sources {
		go m.forward(ctx, i, src, &wg)
	}

	// Closer goroutine: waits for all source goroutines, closes output,
	// closes done. Does NOT write ctxErr — that is latched by source
	// goroutines via noteCancel.
	go func() {
		wg.Wait()
		close(m.out)
		close(m.done)
	}()

	return m
}

// Out returns the receive-only output channel. All items from all sources
// appear on this single channel in nondeterministic order.
//
// The consumer MUST drain Out() to completion or cancel the shared context.
// If the consumer stops reading without canceling, all source goroutines
// block on the shared output send and cannot drain their sources.
//
// Out() is idempotent — it always returns the same channel.
func (m *Merge[T]) Out() <-chan rslt.Result[T] {
	return m.out
}

// Wait blocks until all source goroutines exit and the output channel is
// closed. Returns a latched context error if any source goroutine entered
// a cancel path (pre-send checkpoint or send-select), nil otherwise.
//
// Wait may return nil even if ctx was canceled. This happens when no
// goroutine observes cancellation on a checked path — e.g., all sources
// close before any goroutine loops back to the pre-send check, or a
// goroutine is blocked in range src when cancel fires and the source
// closes without sending. This is intentional: the operator completed its
// work, cancellation had no observable effect on forwarding, and reporting
// it would be a false positive.
//
// Multiple Wait() calls are safe — subsequent calls return immediately
// with the same value.
func (m *Merge[T]) Wait() error {
	<-m.done

	return m.ctxErr
}

// Stats returns approximate metrics. See [MergeStats] for caveats.
//
// Per-source counters are loaded once into plain int64 slices, then
// aggregates are computed from those copied values. This guarantees
// single-call coherence: Received == sum(SourceReceived) within one
// Stats() return, even mid-flight.
//
// Stats are only reliable as final values after [Merge.Wait] returns.
func (m *Merge[T]) Stats() MergeStats {
	n := len(m.sourceReceived)
	received := make([]int64, n)
	forwarded := make([]int64, n)
	dropped := make([]int64, n)

	for i := range n {
		received[i] = m.sourceReceived[i].Load()
		forwarded[i] = m.sourceForwarded[i].Load()
		dropped[i] = m.sourceDropped[i].Load()
	}

	var totalReceived, totalForwarded, totalDropped int64
	for i := range n {
		totalReceived += received[i]
		totalForwarded += forwarded[i]
		totalDropped += dropped[i]
	}

	return MergeStats{
		Received:        totalReceived,
		Forwarded:       totalForwarded,
		Dropped:         totalDropped,
		SourceReceived:  received,
		SourceForwarded: forwarded,
		SourceDropped:   dropped,
	}
}

// noteCancel latches ctxErr on the first cancel observation. Safe for
// concurrent calls from multiple source goroutines.
func (m *Merge[T]) noteCancel(err error) {
	if err == nil {
		return
	}

	m.cancelOnce.Do(func() {
		m.ctxErr = err
	})
}

// drainDiscard enters discard mode: drains src to completion, counting
// each item as received and dropped (source ownership rule).
func (m *Merge[T]) drainDiscard(idx int, src <-chan rslt.Result[T]) {
	for range src {
		m.sourceReceived[idx].Add(1)
		m.sourceDropped[idx].Add(1)
	}
}

// forward is the per-source goroutine. It reads from src, forwards items
// to the shared output channel, and handles cancellation.
func (m *Merge[T]) forward(ctx context.Context, idx int, src <-chan rslt.Result[T], wg *sync.WaitGroup) {
	defer wg.Done()

	for r := range src {
		m.sourceReceived[idx].Add(1)

		// Non-blocking cancellation checkpoint. Without this, the
		// send-select alone can win repeatedly against ctx.Done()
		// if downstream is always ready — unbounded post-cancel
		// forwarding. With it: at most one item per source goroutine.
		select {
		case <-ctx.Done():
			m.noteCancel(ctx.Err())
			m.sourceDropped[idx].Add(1)
			m.drainDiscard(idx, src)

			return
		default:
		}

		select {
		case m.out <- r:
			m.sourceForwarded[idx].Add(1)
		case <-ctx.Done():
			m.noteCancel(ctx.Err())
			m.sourceDropped[idx].Add(1)
			m.drainDiscard(idx, src)

			return
		}
	}
}
