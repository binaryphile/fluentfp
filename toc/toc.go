// Package toc provides a constrained stage runner inspired by
// Drum-Buffer-Rope (Theory of Constraints).
//
// A Stage owns a bounded input queue and a serial (or limited-concurrency)
// worker. Producers submit items via [Stage.Submit]; the stage processes
// them through fn and emits results on [Stage.Out]. When the queue is
// full, Submit blocks — this is the "rope" limiting upstream WIP.
//
// The stage tracks constraint utilization, idle time, and output-blocked
// time via [Stage.Stats], enabling operators to verify the constraint is
// real and detect when the bottleneck shifts downstream.
//
// Lifecycle:
//  1. Start a stage with [Start]
//  2. Submit items with [Stage.Submit] from one or more goroutines
//  3. Call [Stage.CloseInput] when done submitting (use defer as safety net)
//  4. Read results from [Stage.Out] until closed — MUST drain to completion
//  5. Call [Stage.Wait] (or [Stage.Cause]) to block until shutdown completes
//     Or combine steps 4-5: [Stage.DiscardAndWait] / [Stage.DiscardAndCause]
//
// CloseInput is also called internally on fail-fast error or parent context
// cancellation, so the stage always shuts down cleanly. Callers should still
// defer CloseInput for the normal-completion path.
//
// Cardinality: every [Stage.Submit] that returns nil yields exactly one
// [rslt.Result] on [Stage.Out]. Submit calls that return an error produce
// no result.
//
// Output draining: callers MUST read Out() until it closes (e.g.,
// for result := range stage.Out()). See Liveness below for consequences
// of not draining.
//
// Cancellation and fail-fast are cooperative: items already dequeued by a
// worker may still call fn with a canceled context. See [Stage.Submit] for
// admission-closure race semantics.
//
// Terminal status — [Stage.Wait] vs [Stage.Cause]:
//
//	Scenario                          Wait()       Cause()
//	─────────────────────────────────────────────────────────
//	All items succeed                 nil          nil
//	Worker error (fail-fast)          error        error
//	Parent cancel, no worker error    nil          parent cause
//	Worker error + parent cancel      error*       error*
//	Success, parent cancel after done nil          nil
//
//	* nondeterministic: whichever is observed first by the closer
//
// Completion boundary: the stage is "complete" when the closer goroutine
// latches terminal cause and closes the done channel. This happens after
// all workers exit. Workers exit after the input channel closes and all
// dequeued items are processed and sent to [Stage.Out].
//
// Liveness: [Stage.Wait], [Stage.Cause], and the closer goroutine depend
// on all workers exiting. Two conditions can prevent worker exit:
//
//  1. Consumer stops reading Out — workers block on result handoff.
//  2. fn blocks forever or ignores context — the worker never returns.
//
// In either case, the closer never runs, [Stage.Wait] hangs, goroutines
// leak, context resources are retained, and the stage is effectively
// abandoned. fn must eventually return; honoring ctx.Done() is strongly
// recommended. Always drain Out or use [Stage.DiscardAndWait] /
// [Stage.DiscardAndCause].
//
// Total stage WIP is up to Capacity (buffered) + Workers (in-flight).
// With Capacity 0 (unbuffered), WIP equals Workers.
//
// This package is for pipelines with a known bottleneck stage. If you
// don't know your constraint, profile first.
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

// ErrClosed is returned by [Stage.Submit] when the stage is no longer
// accepting input — after [Stage.CloseInput], fail-fast shutdown, or
// parent context cancellation. See [Stage.Submit] for race semantics.
var ErrClosed = errors.New("toc: stage closed")

// Options configures a [Stage].
type Options[T any] struct {
	// Capacity is the number of items the input buffer can hold.
	// Submit blocks when the buffer is full (the "rope").
	// Zero means unbuffered: Submit blocks until a worker dequeues.
	// Negative values panic.
	Capacity int

	// Weight returns the cost of item t for stats tracking only
	// ([Stats.InFlightWeight]). Does not affect admission — capacity
	// is count-based. Called on the Submit path, so must be cheap.
	// Must be pure, non-negative, and safe for concurrent calls.
	// If nil, every item costs 1.
	Weight func(T) int64

	// Workers is the number of concurrent fn invocations.
	// Zero means default: 1 (serial constraint — the common case).
	// Negative values panic.
	Workers int

	// ContinueOnError, when true, keeps processing after fn errors
	// instead of cancelling the stage. Default: false (fail-fast).
	ContinueOnError bool
}

// Stats holds approximate metrics for a [Stage].
//
// Fields are read from independent atomics, so a single Stats value
// is NOT a consistent snapshot — individual fields may reflect
// slightly different moments. Relationships like
// Submitted >= Completed + Canceled are not guaranteed mid-flight.
// Stats are only reliable as final values after all [Stage.Submit] calls
// have returned and [Stage.Wait] returns.
//
// All durations are cumulative across all workers since [Start].
// With Workers > 1, durations can exceed wall-clock time.
type Stats struct {
	Submitted int64 // items accepted by Submit (successful return)
	Completed int64 // items where fn returned (includes Failed and Panicked)
	Failed    int64 // subset of Completed where result is error (includes Panicked)
	Panicked  int64 // subset of Failed where fn panicked
	Canceled  int64 // items dequeued but not passed to fn because cancellation was observed first (not in Completed)

	ServiceTime       time.Duration // cumulative time fn was executing
	IdleTime          time.Duration // cumulative worker time waiting for input (includes startup and tail wait)
	OutputBlockedTime time.Duration // cumulative worker time blocked handing result to consumer (unbuffered out channel)

	BufferedDepth  int64 // approximate items in queue; may transiently be negative mid-flight; 0 when Capacity is 0 (unbuffered)
	InFlightWeight int64 // weighted cost of items currently in fn (stats-only, not admission)
	QueueCapacity  int   // configured capacity
}

// queued wraps an item with its pre-computed weight.
type queued[T any] struct {
	item   T
	weight int64
}

// Stage is a running constrained stage. Created by [Start].
// The zero value is not usable.
type Stage[T, R any] struct {
	in        chan queued[T]
	out       chan rslt.Result[R]
	done      chan struct{}   // closed after Out is closed
	closing   chan struct{}   // closed on shutdown to unblock Submit
	stageDone <-chan struct{} // closed when stageCtx is canceled (parent or fail-fast)
	cancel    context.CancelCauseFunc
	wg        sync.WaitGroup

	// closed guards input shutdown. Once set, Submit rejects.
	closed    atomic.Bool
	closeOnce sync.Once
	sendMu    sync.RWMutex // senders hold RLock; CloseInput holds Lock to close s.in safely

	// stats — atomic for lock-free reads
	submitted     atomic.Int64
	completed     atomic.Int64
	failed        atomic.Int64
	panicked      atomic.Int64
	canceled      atomic.Int64
	bufferedDepth atomic.Int64

	serviceNs       atomic.Int64
	idleNs          atomic.Int64
	outputBlockedNs atomic.Int64

	inFlightWeight atomic.Int64
	capacity       int

	weight func(T) int64

	// err is the first observed fail-fast error (not deterministic
	// by input order — first worker to acquire errMu wins).
	errMu sync.Mutex
	err   error

	// cause is the latched terminal status, written exactly once by the
	// closer goroutine before close(done). Safe to read without a lock
	// after <-s.done returns, per Go memory model: "The closing of a
	// channel is synchronized before a receive that returns a zero value
	// because the channel is closed." Do not write cause from any other
	// goroutine or after close(done).
	cause error
}

// Start launches a constrained stage that processes items through fn.
//
// The stage starts a cancel watcher goroutine that calls [Stage.CloseInput]
// when the context is canceled (either parent cancel or fail-fast). This
// ensures workers always eventually exit.
//
// Panics if ctx is nil, fn is nil, Capacity is negative, or Workers is negative.
func Start[T, R any](
	ctx context.Context,
	fn func(context.Context, T) (R, error),
	opts Options[T],
) *Stage[T, R] {
	if fn == nil {
		panic("toc.Start: fn must not be nil")
	}

	capacity := opts.Capacity
	if capacity < 0 {
		panic("toc.Start: Capacity must be non-negative")
	}

	workers := opts.Workers
	if workers < 0 {
		panic("toc.Start: Workers must be non-negative")
	}
	if workers == 0 {
		workers = 1
	}

	failFast := !opts.ContinueOnError

	weight := opts.Weight
	if weight == nil {
		weight = func(_ T) int64 { return 1 }
	}

	stageCtx, cancel := context.WithCancelCause(ctx)

	s := &Stage[T, R]{
		in:        make(chan queued[T], capacity),
		out:       make(chan rslt.Result[R]),
		done:      make(chan struct{}),
		closing:   make(chan struct{}),
		stageDone: stageCtx.Done(),
		cancel:    cancel,
		capacity:  capacity,
		weight:    weight,
	}

	s.wg.Add(workers)

	for i := 0; i < workers; i++ {
		go s.worker(stageCtx, fn, failFast)
	}

	// Cancel watcher: ensures input closes on any cancellation
	// (parent cancel or fail-fast), so workers always exit.
	go func() {
		<-stageCtx.Done()
		s.CloseInput()
	}()

	// Closer: waits for all workers, latches terminal cause, cleans up.
	// The stage is "complete" when close(s.done) fires. Parent context
	// state is sampled here — if parent cancels between wg.Wait and
	// the latch, Cause() reports parent cancellation (the stage had not
	// yet published completion).
	go func() {
		s.wg.Wait()

		// Latch terminal cause before signaling completion.
		// After this point, Cause() returns a stable, idempotent value.
		s.errMu.Lock()
		if s.err != nil {
			s.cause = s.err
		} else if ctx.Err() != nil {
			s.cause = context.Cause(ctx) // ctx is parent, captured by closure
		}
		s.errMu.Unlock()

		s.cancel(nil) // release context resources; triggers cancel watcher exit
		close(s.out)
		close(s.done)
	}()

	return s
}

// Submit sends item into the stage for processing. Blocks when the
// buffer is full (backpressure / "rope"). Returns [ErrClosed] after
// [Stage.CloseInput], fail-fast shutdown, or parent context cancellation.
//
// If cancellation or CloseInput has already occurred before Submit is
// called, Submit deterministically returns [ErrClosed] without blocking.
// A Submit that is blocked or entering concurrently when shutdown fires
// may nondeterministically succeed or return an error, per Go select
// semantics — even a blocked Submit can succeed if capacity becomes
// available at the same instant. Items admitted during this window are
// processed normally (or canceled if the stage context is already done).
//
// The ctx parameter controls only admission blocking — it is NOT passed
// to fn. The stage's own context (derived from the ctx passed to [Start])
// is what fn receives. This means canceling a submitter's context does
// not cancel the item's processing once admitted.
//
// Panics if ctx is nil (same as context.Context method calls).
// Panics if Weight returns a negative value.
// Note: a panic in Weight propagates to the caller (unlike fn panics,
// which are recovered and wrapped in [rslt.PanicError]).
// Safe for concurrent use from multiple goroutines. Safe to call
// concurrently with [Stage.CloseInput] (will not panic).
func (s *Stage[T, R]) Submit(ctx context.Context, item T) error {
	// Fast path: reject if stage is already closed.
	if s.closed.Load() {
		return ErrClosed
	}

	// Fast path: reject if stage context is already canceled.
	// This catches the case where parent cancel or fail-fast has
	// already fired but the async cancel watcher has not yet called
	// CloseInput. Without this check, the blocking select could
	// nondeterministically choose the send case over stageDone.
	select {
	case <-s.stageDone:
		return ErrClosed
	default:
	}

	// Reject if caller ctx is already canceled.
	if err := ctx.Err(); err != nil {
		return err
	}

	w := s.weight(item)
	if w < 0 {
		panic("toc: Weight must return a non-negative value")
	}

	q := queued[T]{item: item, weight: w}

	return s.trySend(ctx, q)
}

// trySend attempts to send q to the input channel under the sendMu
// read lock, preventing close(s.in) from racing with the send.
func (s *Stage[T, R]) trySend(ctx context.Context, q queued[T]) error {
	s.sendMu.RLock()
	defer s.sendMu.RUnlock()

	// Re-check closed under lock — after acquiring RLock, if closed is
	// true, CloseInput is waiting for us (or hasn't locked yet). Either
	// way, don't send.
	if s.closed.Load() {
		return ErrClosed
	}

	select {
	case s.in <- q:
		s.submitted.Add(1)
		s.bufferedDepth.Add(1)

		return nil
	case <-s.closing:
		return ErrClosed
	case <-s.stageDone:
		return ErrClosed
	case <-ctx.Done():
		return ctx.Err()
	}
}

// CloseInput signals that no more items will be submitted.
// Workers finish processing buffered items, then shut down.
//
// Blocks briefly until all in-flight [Stage.Submit] calls exit, then
// closes the input channel.
//
// Idempotent — safe to call multiple times (use defer as safety net).
// Also called internally on fail-fast or parent context cancellation.
func (s *Stage[T, R]) CloseInput() {
	s.closeOnce.Do(func() {
		s.closed.Store(true)
		close(s.closing) // unblock any in-flight trySend selects

		// Acquire write lock to wait for all in-flight senders to
		// release their read locks, then close s.in safely.
		s.sendMu.Lock()
		close(s.in)
		s.sendMu.Unlock()
	})
}

// Out returns the receive-only output channel. It closes after all
// workers finish and all results have been sent.
//
// Cardinality: every successful [Stage.Submit] produces exactly one result.
//
// Ordering: with Workers == 1, results are delivered in submit order.
// With Workers > 1, result order is nondeterministic.
//
// Callers MUST drain Out to completion (or use [Stage.DiscardAndWait]
// / [Stage.DiscardAndCause]):
//
//	for result := range stage.Out() {
//	    val, err := result.Unpack()
//	    // handle result — do NOT break out of this loop
//	}
//
// After cancellation or fail-fast, Out may still emit: success results
// from work already in fn, ordinary error results from in-flight work,
// and canceled results for buffered items drained post-cancel. With
// Workers > 1, the fail-fast triggering error is not guaranteed to appear
// before cancellation results; use [Stage.Wait] or [Stage.Cause] for
// stage-level terminal status, not stream order.
//
// If the consumer stops reading, workers block sending results, which
// prevents shutdown and causes [Stage.Wait] to hang — leaking goroutines,
// context resources, and the stage itself. This is the same contract as
// consuming from any Go channel-based pipeline.
func (s *Stage[T, R]) Out() <-chan rslt.Result[R] { return s.out }

// Wait blocks until all workers have finished and Out is closed.
// Wait does not initiate shutdown — call [Stage.CloseInput] first.
// Without CloseInput, Wait blocks forever if no more items are submitted.
// Requires that Out is drained concurrently (see [Stage.Out]).
//
// Returns the first observed fail-fast error, or nil. Specifically:
//   - In fail-fast mode: returns the first fn error that caused shutdown
//     (nondeterministic among concurrent workers — first to acquire lock)
//   - In ContinueOnError mode: always returns nil (check individual results)
//   - On parent context cancellation: returns nil (caller already knows;
//     errors returned by fn after parent cancel are not stored)
//
// If parent cancellation races with a worker error, Wait may return
// either nil or the error depending on observation order. Use [Stage.Cause]
// for terminal status that distinguishes all three outcomes.
func (s *Stage[T, R]) Wait() error {
	<-s.done

	// s.err is stable after <-s.done: the closer goroutine is the last
	// writer (under errMu), and close(s.done) happens-after that unlock.
	return s.err
}

// Cause returns the latched terminal cause of the stage, blocking until
// shutdown completes. Like [Stage.Wait], Cause does not initiate shutdown.
// Unlike [Stage.Wait], Cause distinguishes all three outcomes:
//   - nil: all items completed successfully (or ContinueOnError with no cancel)
//   - fail-fast error: the first fn error that caused shutdown
//   - parent cancel cause: [context.Cause] of the parent context at completion
//
// The terminal cause is latched when the last worker finishes, so Cause
// is stable and idempotent — it returns the same value regardless of later
// parent context changes. If parent cancellation races with a worker error,
// Cause may return either depending on observation order.
//
// A parent cancellation that occurs after fn returns but before the worker
// completes result handoff is reported as stage cancellation, even though
// all business-level computations succeeded. Use [Stage.Wait] if only
// fail-fast errors matter.
//
// Requires Out to be drained (see [Stage.DiscardAndCause] when individual
// results are not needed).
func (s *Stage[T, R]) Cause() error {
	<-s.done

	return s.cause
}

// DiscardAndWait drains all remaining results from [Stage.Out] and returns
// [Stage.Wait]'s error. Use when individual results are not needed.
//
// Requires exclusive ownership of [Stage.Out] — must not be called while
// another goroutine is reading Out. Mixing DiscardAndWait with direct
// Out consumption causes a consumption race (results go to the wrong reader).
func (s *Stage[T, R]) DiscardAndWait() error {
	for range s.Out() {
	}

	return s.Wait()
}

// DiscardAndCause drains all remaining results from [Stage.Out] and returns
// [Stage.Cause]'s latched terminal cause. Use when individual results are
// not needed but composable terminal status is required.
//
// Requires exclusive ownership of [Stage.Out] — must not be called while
// another goroutine is reading Out. Mixing DiscardAndCause with direct
// Out consumption causes a consumption race (results go to the wrong reader).
func (s *Stage[T, R]) DiscardAndCause() error {
	for range s.Out() {
	}

	return s.Cause()
}

// Stats returns approximate metrics. See [Stats] for caveats.
// Stats are only reliable as final values after all [Stage.Submit] calls
// have returned and [Stage.Wait] returns.
func (s *Stage[T, R]) Stats() Stats {
	var depth int64
	if s.capacity > 0 {
		depth = s.bufferedDepth.Load()
	}

	return Stats{
		Submitted:         s.submitted.Load(),
		Completed:         s.completed.Load(),
		Failed:            s.failed.Load(),
		Panicked:          s.panicked.Load(),
		Canceled:          s.canceled.Load(),
		ServiceTime:       time.Duration(s.serviceNs.Load()),
		IdleTime:          time.Duration(s.idleNs.Load()),
		OutputBlockedTime: time.Duration(s.outputBlockedNs.Load()),
		BufferedDepth:     depth,
		InFlightWeight:    s.inFlightWeight.Load(),
		QueueCapacity:     s.capacity,
	}
}

// worker pulls from in, calls fn, and sends every result to out.
//
// Cancellation is cooperative: a worker that has already dequeued an item
// and passed the pre-call ctx.Err() check may still call fn with a context
// that becomes canceled during execution. Buffered items dequeued after
// cancellation are emitted as canceled results without calling fn.
func (s *Stage[T, R]) worker(
	ctx context.Context,
	fn func(context.Context, T) (R, error),
	failFast bool,
) {
	defer s.wg.Done()

	for {
		idleStart := time.Now()
		q, ok := <-s.in
		s.idleNs.Add(int64(time.Since(idleStart)))

		if !ok {
			return
		}

		s.bufferedDepth.Add(-1)

		// Check cancellation before calling fn.
		if err := ctx.Err(); err != nil {
			s.canceled.Add(1)

			outStart := time.Now()
			s.out <- rslt.Err[R](context.Cause(ctx))
			s.outputBlockedNs.Add(int64(time.Since(outStart)))

			continue
		}

		s.inFlightWeight.Add(q.weight)

		serviceStart := time.Now()
		result := s.safeCall(ctx, fn, q.item)
		s.serviceNs.Add(int64(time.Since(serviceStart)))

		s.inFlightWeight.Add(-q.weight)
		s.completed.Add(1)

		if _, err := result.Unpack(); err != nil {
			s.failed.Add(1)

			if failFast {
				s.errMu.Lock()
				// Only store if this is the triggering error, not a
				// consequence of prior cancellation.
				first := s.err == nil && ctx.Err() == nil
				if first {
					s.err = err
				}
				s.errMu.Unlock()

				if first {
					s.cancel(err)
					s.CloseInput() // synchronous admission closure
				}
			}
		}

		outStart := time.Now()
		s.out <- result
		s.outputBlockedNs.Add(int64(time.Since(outStart)))
	}
}

// safeCall invokes fn with panic recovery, returning a Result.
func (s *Stage[T, R]) safeCall(
	ctx context.Context,
	fn func(context.Context, T) (R, error),
	item T,
) (result rslt.Result[R]) {
	defer func() {
		if r := recover(); r != nil {
			s.panicked.Add(1)
			result = rslt.Err[R](&rslt.PanicError{Value: r, Stack: debug.Stack()})
		}
	}()

	val, err := fn(ctx, item)
	if err != nil {
		return rslt.Err[R](err)
	}

	return rslt.Ok(val)
}
