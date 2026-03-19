package toc

import (
	"context"
	"errors"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/binaryphile/fluentfp/rslt"
)

// sideState tracks per-source progress in the Join collect phase.
type sideState int

const (
	sOpen        sideState = iota // waiting for first item
	sGotFirst                     // have first item, still reading extras
	sClosedEmpty                  // closed without producing any item
	sClosed                       // closed after producing at least one item
)

// hasItem reports whether the side has consumed at least one item.
func (s sideState) hasItem() bool { return s == sGotFirst || s == sClosed }

// resolved reports whether the side has reached a terminal state
// (closed with or without producing).
func (s sideState) resolved() bool { return s == sClosedEmpty || s == sClosed }

// MissingResultError indicates a source closed without producing a result.
// Callers can use [errors.As] to extract the Source field and determine
// which side was missing.
type MissingResultError struct {
	Source string // "A" or "B"
}

func (e *MissingResultError) Error() string {
	return "toc.Join: source " + e.Source + " closed without producing a result"
}

// JoinStats holds metrics for a [Join].
//
// Fields are read from independent atomics, so a mid-flight Stats value is
// NOT a consistent snapshot. Invariants are guaranteed only after [Join.Wait]
// returns.
//
// Conservation invariant (after Wait):
//   - ReceivedA = Combined + DiscardedA + ExtraA
//   - ReceivedB = Combined + DiscardedB + ExtraB
//   - Combined + Errors <= 1
//
// Counter precedence: DiscardedX counts only first items that were consumed
// but not successfully combined (error, cancel, panic, or other side missing).
// ExtraX counts items beyond the first, drained after the join decision is
// reached. Post-decision items are always classified as ExtraX, even if
// cancellation later prevents result delivery.
type JoinStats struct {
	ReceivedA  int64 // total items consumed from srcA
	ReceivedB  int64 // total items consumed from srcB
	Combined   int64 // successful fn(a,b) combinations (0 or 1)
	Errors     int64 // error results delivered to Out (0 or 1)
	DiscardedA int64 // A items consumed but not part of a successful combination (error, missing, cancel, panic)
	DiscardedB int64 // B items consumed but not part of a successful combination (error, missing, cancel, panic)
	ExtraA     int64 // A items beyond the first, drained after the join decision (contract violation)
	ExtraB     int64 // B items beyond the first, drained after the join decision (contract violation)

	OutputBlockedTime time.Duration // time blocked sending result to out
}

// Join is a strict branch recombination operator: it uses the first item
// from each of two source channels for join semantics (combine or error),
// then drains all remaining items from both sources (source ownership rule).
//
// Join is designed for recombining [Tee] branches — each branch is
// expected to produce exactly one result. Missing items (source closes
// without producing) and extra items (source produces more than one)
// are contract violations handled gracefully: missing items produce
// [MissingResultError], extra items are drained and counted in stats.
//
// Created by [NewJoin]. The zero value is not usable.
// A Join must not be copied after first use.
type Join[R any] struct {
	out  chan rslt.Result[R]
	done chan struct{}

	receivedA  atomic.Int64
	receivedB  atomic.Int64
	combined   atomic.Int64
	errors     atomic.Int64
	discardedA atomic.Int64
	discardedB atomic.Int64
	extraA     atomic.Int64
	extraB     atomic.Int64

	outputBlockedNs atomic.Int64

	cancelOnce sync.Once
	ctxErr     error
}

// NewJoin creates a Join that uses the first item from each source for
// join semantics, then drains remaining items. Ok/Ok pairs are combined
// via fn; errors are forwarded. The result is emitted on [Join.Out].
//
// Error matrix:
//   - Ok(a), Ok(b) → Ok(fn(a, b))
//   - Ok(a), Err(e) → Err(e); a discarded
//   - Err(e), Ok(b) → Err(e); b discarded
//   - Err(ea), Err(eb) → Err(errors.Join(ea, eb))
//   - Ok(a), missing → Err([MissingResultError]{Source: "B"}); a discarded
//   - missing, Ok(b) → Err([MissingResultError]{Source: "A"}); b discarded
//   - Err(e), missing → Err(errors.Join(e, [MissingResultError]{Source: "B"}))
//   - missing, Err(e) → Err(errors.Join([MissingResultError]{Source: "A"}, e))
//   - missing, missing → no output
//
// Each source is drained to completion (source ownership rule). Extra
// items beyond the first are counted in [JoinStats.ExtraA] / [JoinStats.ExtraB].
// Sources may close at different times.
//
// fn must be a pure, synchronous combiner — it must not block, perform
// I/O, or depend on cancellation. fn runs on the Join's only goroutine;
// a blocking fn prevents cancellation observation, source draining, and
// [Join.Wait] from returning. Similarly, if the consumer stops reading
// [Join.Out] without canceling, the goroutine blocks on the output send
// with the same consequences. If combining can fail or block, use a
// downstream [Pipe] for the error-capable, context-aware transform.
// Panics in fn are recovered as [rslt.PanicError].
//
// Cancellation is best-effort and observed only in the Phase 1 select and
// the output send. On ctx cancellation, consumed items are discarded and
// both sources are drained. A pre-send checkpoint catches
// already-observable cancellation before attempting the output send, but
// a result may still be emitted if the send races with cancellation
// (both select cases ready). [Join.Wait] returns the latched context
// error if cancellation was observed during collection or output;
// [Join.Wait] may return nil even if the context was canceled, if the
// goroutine completed without observing it. Drain is unconditional —
// sources must close for the goroutine to exit.
//
// Panics if srcA, srcB, ctx, or fn is nil.
func NewJoin[A, B, R any](
	ctx context.Context,
	srcA <-chan rslt.Result[A],
	srcB <-chan rslt.Result[B],
	fn func(A, B) R,
) *Join[R] {
	if ctx == nil {
		panic("toc.NewJoin: ctx must not be nil")
	}
	if srcA == nil {
		panic("toc.NewJoin: srcA must not be nil")
	}
	if srcB == nil {
		panic("toc.NewJoin: srcB must not be nil")
	}
	if fn == nil {
		panic("toc.NewJoin: fn must not be nil")
	}

	j := &Join[R]{
		out:  make(chan rslt.Result[R]),
		done: make(chan struct{}),
	}

	go func() {
		defer close(j.done)
		defer close(j.out)

		var a rslt.Result[A]
		var b rslt.Result[B]
		stateA, stateB := sOpen, sOpen
		chA, chB := srcA, srcB // local copies — nil only on channel close

		// drain drains both sources to completion. Each item is counted
		// as Received + the provided counter (Extra for post-decision
		// drain, Discarded for cancel drain).
		drain := func(counterA, counterB *atomic.Int64) {
			dA, dB := srcA, srcB

			for dA != nil || dB != nil {
				select {
				case _, ok := <-dA:
					if !ok {
						dA = nil

						continue
					}

					j.receivedA.Add(1)
					counterA.Add(1)
				case _, ok := <-dB:
					if !ok {
						dB = nil

						continue
					}

					j.receivedB.Add(1)
					counterB.Add(1)
				}
			}
		}

		// trySend sends r to out, respecting cancellation. Pre-send
		// checkpoint catches already-observable cancellation; the
		// blocking send-select catches cancellation during the send.
		// A race between send and cancel can still deliver.
		// Returns true if sent, false if canceled.
		trySend := func(r rslt.Result[R]) bool {
			// Pre-send checkpoint.
			select {
			case <-ctx.Done():
				return false
			default:
			}

			outStart := time.Now()

			select {
			case j.out <- r:
				j.outputBlockedNs.Add(int64(time.Since(outStart)))

				return true
			case <-ctx.Done():
				j.outputBlockedNs.Add(int64(time.Since(outStart)))

				return false
			}
		}

		// Phase 1: collect one from each + absorb extras.
		for {
			// Exit when both sides are resolved (closed with or without item).
			if stateA.resolved() && stateB.resolved() {
				break
			}

			// Exit when both sides have an item.
			if stateA.hasItem() && stateB.hasItem() {
				break
			}

			// Exit when one side closed empty and other has item.
			if stateA == sClosedEmpty && stateB.hasItem() {
				break
			}

			if stateB == sClosedEmpty && stateA.hasItem() {
				break
			}

			select {
			case item, ok := <-chA:
				if !ok {
					chA = nil

					if stateA == sOpen {
						stateA = sClosedEmpty
					} else {
						stateA = sClosed
					}

					continue
				}

				j.receivedA.Add(1)

				if stateA == sOpen {
					a = item
					stateA = sGotFirst
				} else {
					j.extraA.Add(1)
				}
			case item, ok := <-chB:
				if !ok {
					chB = nil

					if stateB == sOpen {
						stateB = sClosedEmpty
					} else {
						stateB = sClosed
					}

					continue
				}

				j.receivedB.Add(1)

				if stateB == sOpen {
					b = item
					stateB = sGotFirst
				} else {
					j.extraB.Add(1)
				}
			case <-ctx.Done():
				j.noteCancel(ctx.Err())

				if stateA.hasItem() {
					j.discardedA.Add(1)
				}

				if stateB.hasItem() {
					j.discardedB.Add(1)
				}

				drain(&j.discardedA, &j.discardedB)

				return
			}
		}

		// Phase 2: combine and emit.
		haveA := stateA.hasItem()
		haveB := stateB.hasItem()

		var result rslt.Result[R]
		emitResult := true

		switch {
		case haveA && haveB:
			valA, errA := a.Unpack()
			valB, errB := b.Unpack()

			switch {
			case errA == nil && errB == nil:
				// Ok/Ok → combine via fn with panic recovery.
				result = func() (r rslt.Result[R]) {
					defer func() {
						if p := recover(); p != nil {
							r = rslt.Err[R](&rslt.PanicError{Value: p, Stack: debug.Stack()})
						}
					}()

					return rslt.Ok(fn(valA, valB))
				}()

				if result.IsOk() {
					j.combined.Add(1)
				} else {
					// fn panicked.
					j.errors.Add(1)
					j.discardedA.Add(1)
					j.discardedB.Add(1)
				}
			case errA == nil && errB != nil:
				// Ok/Err → forward B's error, discard both.
				result = rslt.Err[R](errB)
				j.errors.Add(1)
				j.discardedA.Add(1)
				j.discardedB.Add(1)
			case errA != nil && errB == nil:
				// Err/Ok → forward A's error, discard both.
				result = rslt.Err[R](errA)
				j.errors.Add(1)
				j.discardedA.Add(1)
				j.discardedB.Add(1)
			default:
				// Err/Err → errors.Join preserving both.
				result = rslt.Err[R](errors.Join(errA, errB))
				j.errors.Add(1)
				j.discardedA.Add(1)
				j.discardedB.Add(1)
			}
		case haveA && !haveB:
			// A has item, B missing.
			missingB := &MissingResultError{Source: "B"}

			if _, errA := a.Unpack(); errA != nil {
				result = rslt.Err[R](errors.Join(errA, missingB))
			} else {
				result = rslt.Err[R](missingB)
			}

			j.errors.Add(1)
			j.discardedA.Add(1)
		case !haveA && haveB:
			// B has item, A missing.
			missingA := &MissingResultError{Source: "A"}

			if _, errB := b.Unpack(); errB != nil {
				result = rslt.Err[R](errors.Join(missingA, errB))
			} else {
				result = rslt.Err[R](missingA)
			}

			j.errors.Add(1)
			j.discardedB.Add(1)
		default:
			// Both missing → no output.
			emitResult = false
		}

		if emitResult {
			if !trySend(result) {
				j.noteCancel(ctx.Err())

				// Result not delivered — adjust stats.
				if result.IsOk() {
					// Combined result dropped.
					j.combined.Add(-1)
					j.discardedA.Add(1)
					j.discardedB.Add(1)
				} else {
					// Error not delivered — undo count.
					// Discards are already counted correctly.
					j.errors.Add(-1)
				}
			}
		}

		// Phase 3: drain remaining items from both sources.
		// Items here are genuinely beyond the first — contract violations.
		drain(&j.extraA, &j.extraB)
	}()

	return j
}

// Out returns the receive-only output channel. At most one result will
// appear on this channel.
//
// The consumer MUST drain Out() to completion or cancel the shared context.
// If the consumer stops reading without canceling, the goroutine blocks on
// the output send.
//
// Out() is idempotent — it always returns the same channel.
func (j *Join[R]) Out() <-chan rslt.Result[R] {
	return j.out
}

// Wait blocks until the goroutine exits and the output channel is closed.
// Returns a latched context error if cancellation was observed during
// collection or output. May return nil even if the context was canceled,
// if the goroutine completed without observing it (e.g., both sources
// closed before cancellation was checked).
//
// Multiple Wait() calls are safe — subsequent calls return immediately
// with the same value.
func (j *Join[R]) Wait() error {
	<-j.done

	return j.ctxErr
}

// Stats returns approximate metrics. See [JoinStats] for caveats.
// Stats are only reliable as final values after [Join.Wait] returns.
func (j *Join[R]) Stats() JoinStats {
	return JoinStats{
		ReceivedA:         j.receivedA.Load(),
		ReceivedB:         j.receivedB.Load(),
		Combined:          j.combined.Load(),
		Errors:            j.errors.Load(),
		DiscardedA:        j.discardedA.Load(),
		DiscardedB:        j.discardedB.Load(),
		ExtraA:            j.extraA.Load(),
		ExtraB:            j.extraB.Load(),
		OutputBlockedTime: time.Duration(j.outputBlockedNs.Load()),
	}
}

// noteCancel latches ctxErr on the first cancel observation. Safe for
// concurrent calls.
func (j *Join[R]) noteCancel(err error) {
	if err == nil {
		return
	}

	j.cancelOnce.Do(func() {
		j.ctxErr = err
	})
}
