package toc

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/binaryphile/fluentfp/rslt"
)

// TeeStats holds metrics for a [Tee].
//
// Fields are read from independent atomics, so a mid-flight Stats value is
// NOT a consistent snapshot — individual fields may reflect different moments.
// Invariants are guaranteed only after [Tee.Wait] returns.
//
// Invariant (after Wait): Received = FullyDelivered + PartiallyDelivered + Undelivered.
//
// PartiallyDelivered is at most 1 per Tee lifetime: once cancellation
// interrupts delivery mid-item, the goroutine enters discard mode and
// does not attempt delivery on subsequent items.
//
// BranchBlockedTime[i] measures direct send-wait time on branch i. It does
// not measure end-to-end latency imposed by other branches. Because branches
// are sent in index order, earlier branches' blocked time reflects their
// consumer's speed directly; later branches' blocked time is near zero even
// if they are throttled by earlier branches.
type TeeStats struct {
	Received           int64 // items consumed from src
	FullyDelivered     int64 // items sent to ALL branches
	PartiallyDelivered int64 // items sent to some branches before cancel (≥1, <N)
	Undelivered        int64 // items not sent to any branch (cancel before first send, or discard mode)

	// Per-branch stats. Len == N (number of branches).
	// BranchDelivered[i] = items successfully sent to branch i.
	// BranchBlockedTime[i] = cumulative time blocked sending to branch i.
	BranchDelivered   []int64
	BranchBlockedTime []time.Duration
}

// Tee is a synchronous lockstep broadcast from one source to N branches.
//
// Created by [NewTee]. The zero value is not usable.
// A Tee must not be copied after first use.
type Tee[T any] struct {
	branches []chan rslt.Result[T]
	done     chan struct{}

	received           atomic.Int64
	fullyDelivered     atomic.Int64
	partiallyDelivered atomic.Int64
	undelivered        atomic.Int64

	// Per-branch atomics. Indexed by branch number.
	branchDelivered []atomic.Int64
	branchBlockedNs []atomic.Int64

	ctxErr error // latched after goroutine exits, before close(done)
}

// NewTee creates a Tee that reads from src and broadcasts each item to n
// unbuffered output branches. Items are sent to branches sequentially in
// index order — branch 0 first, then branch 1, etc. The slowest consumer
// governs pace (synchronous lockstep).
//
// The Tee drains src to completion (source ownership rule), provided all
// branch consumers drain their branches or ctx is canceled, and src
// eventually closes. Cancellation unblocks branch sends but does not
// force-close src. After ctx cancellation, it switches to discard mode:
// continues reading src but discards all items. Branch sends are
// best-effort during shutdown — cancel may race with a successful send,
// so output may still appear after cancellation. All drops are reflected
// in stats. If a branch consumer stops reading and ctx is never canceled,
// the Tee blocks on that branch's send and stalls all branches.
//
// Tee does not clone payloads. Reference-containing payloads (pointers,
// slices, maps) may alias across branches. Consumers must treat received
// values as immutable; mutation after receipt is a data race.
//
// Panics if n <= 0, src is nil, or ctx is nil.
func NewTee[T any](ctx context.Context, src <-chan rslt.Result[T], n int) *Tee[T] {
	if n <= 0 {
		panic("toc.NewTee: n must be positive")
	}
	if src == nil {
		panic("toc.NewTee: src must not be nil")
	}
	if ctx == nil {
		panic("toc.NewTee: ctx must not be nil")
	}

	branches := make([]chan rslt.Result[T], n)
	for i := range branches {
		branches[i] = make(chan rslt.Result[T])
	}

	t := &Tee[T]{
		branches:        branches,
		done:            make(chan struct{}),
		branchDelivered: make([]atomic.Int64, n),
		branchBlockedNs: make([]atomic.Int64, n),
	}

	go t.run(ctx, src)

	return t
}

// Branch returns the receive-only output channel for branch i.
//
// Callers MUST drain all branches to completion or cancel the shared
// context. An undrained branch blocks the Tee and stalls all branches.
//
// Panics if i is out of range [0, n).
func (t *Tee[T]) Branch(i int) <-chan rslt.Result[T] {
	if i < 0 || i >= len(t.branches) {
		panic("toc.Tee.Branch: index out of range")
	}

	return t.branches[i]
}

// Wait blocks until the Tee goroutine exits. Returns ctx.Err() if context
// cancellation caused items to be dropped (discard mode or interrupted
// branch send), nil otherwise.
func (t *Tee[T]) Wait() error {
	<-t.done

	return t.ctxErr
}

// Stats returns approximate metrics. See [TeeStats] for caveats.
// Stats are only reliable as final values after [Tee.Wait] returns.
func (t *Tee[T]) Stats() TeeStats {
	n := len(t.branches)
	delivered := make([]int64, n)
	blocked := make([]time.Duration, n)

	for i := range n {
		delivered[i] = t.branchDelivered[i].Load()
		blocked[i] = time.Duration(t.branchBlockedNs[i].Load())
	}

	return TeeStats{
		Received:           t.received.Load(),
		FullyDelivered:     t.fullyDelivered.Load(),
		PartiallyDelivered: t.partiallyDelivered.Load(),
		Undelivered:        t.undelivered.Load(),
		BranchDelivered:    delivered,
		BranchBlockedTime:  blocked,
	}
}

// run is the Tee's single goroutine.
func (t *Tee[T]) run(ctx context.Context, src <-chan rslt.Result[T]) {
	defer close(t.done)
	defer func() {
		for _, ch := range t.branches {
			close(ch)
		}
	}()

	// drainAndDiscard enters discard mode: drains src to completion
	// (source ownership rule).
	drainAndDiscard := func() {
		for range src {
			t.received.Add(1)
			t.undelivered.Add(1)
		}

		t.ctxErr = ctx.Err()
	}

	for r := range src {
		t.received.Add(1)

		// Check for cancellation — switch to discard mode.
		select {
		case <-ctx.Done():
			t.undelivered.Add(1)
			drainAndDiscard()

			return
		default:
		}

		deliveredTo := 0

	branches:
		for i, ch := range t.branches {
			outStart := time.Now()

			select {
			case ch <- r:
				t.branchBlockedNs[i].Add(int64(time.Since(outStart)))
				t.branchDelivered[i].Add(1)
				deliveredTo++
			case <-ctx.Done():
				t.branchBlockedNs[i].Add(int64(time.Since(outStart)))

				break branches
			}
		}

		switch {
		case deliveredTo == len(t.branches):
			t.fullyDelivered.Add(1)
		case deliveredTo > 0:
			t.partiallyDelivered.Add(1)
			drainAndDiscard()

			return
		default:
			t.undelivered.Add(1)
			drainAndDiscard()

			return
		}
	}
}
