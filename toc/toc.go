package toc

import (
	"context"
	"errors"
	"runtime/debug"
	"runtime/metrics"
	"sync"
	"sync/atomic"
	"time"

	"github.com/binaryphile/fluentfp/rslt"
)

// Metric names for runtime/metrics allocation sampling.
const (
	metricAllocBytes   = "/gc/heap/allocs:bytes"
	metricAllocObjects = "/gc/heap/allocs:objects"
)

// allocMetricsProbe checks whether both allocation metric names exist
// exactly once with KindUint64 in the given descriptions. Pure function
// for testability; called once via allocMetricsOnce.
func allocMetricsProbe(descs []metrics.Description) bool {
	var haveBytes, haveObjects bool
	for _, d := range descs {
		switch d.Name {
		case metricAllocBytes:
			if d.Kind != metrics.KindUint64 || haveBytes {
				return false // wrong kind or duplicate
			}
			haveBytes = true
		case metricAllocObjects:
			if d.Kind != metrics.KindUint64 || haveObjects {
				return false // wrong kind or duplicate
			}
			haveObjects = true
		}
	}
	return haveBytes && haveObjects
}

var (
	allocMetricsOnce      sync.Once
	allocMetricsSupported bool
)

func isAllocMetricsSupported() bool {
	allocMetricsOnce.Do(func() {
		allocMetricsSupported = allocMetricsProbe(metrics.All())
	})
	return allocMetricsSupported
}

// ErrClosed is returned by [Stage.Submit] when the stage is no longer
// accepting input — after [Stage.CloseInput], fail-fast shutdown, or
// parent context cancellation. See [Stage.Submit] for race semantics.
var ErrClosed = errors.New("toc: stage closed")

// Options configures a [Stage].
//
// Total stage WIP (item count) is up to Capacity (buffered) + Workers
// (in-flight). Capacity is always an item-count bound; [Options.Weight]
// affects stats only, not admission.
type Options[T any] struct {
	// Capacity is the number of items the input buffer can hold.
	// Submit blocks when the buffer is full.
	// Zero means unbuffered: Submit blocks until a worker dequeues.
	// Negative values panic.
	Capacity int

	// MaxWIP is the maximum number of admitted items (buffered +
	// in-flight). Submit blocks when the limit is reached (the "rope").
	// Zero means default: Capacity + Workers (backward compatible).
	// Clamped to max(1, min(MaxWIP, Capacity + Workers)) at construction.
	// Adjust at runtime with [Stage.SetMaxWIP].
	MaxWIP int

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

	// TrackAllocations, when true, samples process-wide heap allocation
	// counters (runtime/metrics /gc/heap/allocs:bytes and :objects) before
	// and after each fn invocation and accumulates the deltas into
	// [Stats.ObservedAllocBytes] and [Stats.ObservedAllocObjects].
	//
	// Scope: each sample captures the invocation window of a single fn
	// call. Counters are process-global — they include allocations by any
	// goroutine during that window, not just the stage's own work.
	//
	// Concurrent over-attribution: with Workers > 1, overlapping
	// invocation windows can each capture the same unrelated allocation,
	// so per-stage totals can exceed actual process allocations. Totals
	// are also not additive across stages for the same reason.
	//
	// Overhead: on the order of 1µs per item in single-worker throughput
	// benchmarks (two runtime/metrics.Read calls plus counter extraction
	// and atomic accumulation). Negligible when fn does real work;
	// roughly doubles overhead for no-op or sub-microsecond fns.
	// Multi-worker contention on shared atomic counters may add cost.
	//
	// Default: false (disabled). Enable when diagnosing allocation-heavy
	// stages. Silently disabled if the runtime does not support the
	// required metrics (validated on first use via sync.Once).
	TrackAllocations bool
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

	// Pipe-specific counters. Zero for Start-created stages.
	// Invariant (after Wait): Received = Submitted + Forwarded + Dropped.
	Received  int64 // items consumed from src by feeder (any Result)
	Forwarded int64 // upstream Err items sent directly to out (bypassed fn)
	Dropped   int64 // items seen but neither submitted nor forwarded (shutdown/cancel)

	ServiceTime       time.Duration // cumulative time fn was executing
	IdleTime          time.Duration // cumulative worker time waiting for input (includes startup and tail wait)
	OutputBlockedTime time.Duration // cumulative worker time blocked handing result to consumer (unbuffered out channel)

	BufferedDepth  int64 // approximate items in queue; may transiently be negative mid-flight; 0 when Capacity is 0 (unbuffered)
	InFlightWeight int64 // weighted cost of items currently in fn (stats-only, not admission)
	QueueCapacity  int   // configured capacity

	MaxWIP        int   // current WIP limit (>= 1)
	Admitted      int64 // reserved permits, not active workers. Includes buffered, in-flight, and reserved-but-not-yet-enqueued. Can exceed current MaxWIP after SetMaxWIP shrinks the limit (existing permits are not revoked) or during the brief grant-to-enqueue window. Not a hard invariant gauge — use for observability, not alerting thresholds.
	WaiterCount    int // current number of Submits blocked on rope
	MaxWaiterCount int // high-water mark for waiter queue depth
	RopeWaitCount int64 // cumulative: total submissions that blocked on rope (not current blocked count)
	RopeWaitNs    int64 // cumulative rope wait time (nanoseconds)

	// AllocTrackingActive reports whether allocation sampling is
	// effectively enabled for this stage. False when
	// [Options.TrackAllocations] was not set, or when the runtime does
	// not support the required metrics. Allows callers to distinguish
	// "tracking requested but unsupported" from "tracking not requested"
	// or "tracking active but fn allocated zero."
	AllocTrackingActive bool

	// ObservedAllocBytes and ObservedAllocObjects are cumulative heap
	// allocation counters sampled via runtime/metrics around each fn
	// invocation. Zero when AllocTrackingActive is false.
	//
	// Process-global, not stage-exclusive: includes allocations by any
	// goroutine during each fn invocation window. With Workers > 1,
	// overlapping windows can capture the same unrelated allocation in
	// multiple workers, so per-stage totals can exceed actual process
	// allocations over the same period. Not additive across stages.
	// Biased upward by longer service times (more background noise).
	// Best used as a directional signal under stable workload where the
	// stage dominates allocations, not for precise attribution. For
	// exact allocation profiling, use go tool pprof.
	ObservedAllocBytes   uint64
	ObservedAllocObjects uint64
}

// testHooks holds optional barrier callbacks for deterministic testing.
// nil in production. Assigned before goroutines launch, read-only after.
type testHooks struct {
	afterAdmitFastPath  func() // after admitted++ and Unlock in acquireAdmission
	afterWaiterQueued   func() // after waiters append and Unlock, before select
	onGrant             func() // in grantWaitersLocked after close(w.ready) — UNDER LOCK, must not block
	afterRelease        func() // after Unlock in releaseAdmission
	afterCloseAdmission func() // in CloseInput after closedForAdmission+Unlock, before close(s.closing)
	beforeRevokeOrHonor func() // before admissionMu.Lock in revokeOrHonor
}

// notifyNonBlocking sends to ch without blocking. Safe to call under a mutex.
func notifyNonBlocking(ch chan<- struct{}) {
	select {
	case ch <- struct{}{}:
	default:
	}
}

// waiter represents a blocked Submit waiting for an admission slot.
// grantWaitersLocked closes ready to wake exactly this waiter.
type waiter struct {
	ready chan struct{}
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
	hooks     *testHooks     // nil in production; set by tests for deterministic barriers
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
	received      atomic.Int64 // Pipe feeder: items consumed from src
	forwarded     atomic.Int64 // Pipe feeder: upstream Err items sent to out
	dropped       atomic.Int64 // Pipe feeder: items neither submitted nor forwarded

	serviceNs       atomic.Int64
	idleNs          atomic.Int64
	outputBlockedNs atomic.Int64

	inFlightWeight atomic.Int64
	allocBytes     atomic.Uint64
	allocObjects   atomic.Uint64

	trackAllocs bool
	capacity    int
	workers     int

	weight func(T) int64

	// Rope: admission control via per-waiter channel queue.
	admissionMu        sync.Mutex
	admitted           int64    // items admitted but not yet completed
	maxWIP             int      // current WIP limit; >= 1
	closedForAdmission bool     // authoritative gate for permit creation; protected by admissionMu. Set by CloseInput before closed/closing. All permit paths (fast path, slow path, grantWaitersLocked) check this under lock.
	waiters            []waiter // FIFO queue of blocked Submits
	maxWaiterCount     int      // high-water mark for waiter queue depth
	ropeWaitCnt        atomic.Int64
	ropeWaitNs         atomic.Int64

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

// feederFunc is the signature for Pipe's feeder goroutine.
// The feeder reads from an upstream source and feeds items into the stage.
// It is registered in wg before launch and must defer wg.Done().
type feederFunc[T, R any] func(s *Stage[T, R], stageCtx context.Context)

// start is the internal constructor shared by Start and Pipe.
// All goroutines that send to out (workers + optional feeder) are registered
// in wg before any goroutine launches. Invariant: out closes only after
// every possible sender is done (wg.Wait returns).
func start[T, R any](
	ctx context.Context,
	fn func(context.Context, T) (R, error),
	opts Options[T],
	feeder feederFunc[T, R],
) *Stage[T, R] {
	capacity := opts.Capacity
	if capacity < 0 {
		panic("toc: Capacity must be non-negative")
	}

	workers := opts.Workers
	if workers < 0 {
		panic("toc: Workers must be non-negative")
	}
	if workers == 0 {
		workers = 1
	}

	failFast := !opts.ContinueOnError

	weight := opts.Weight
	if weight == nil {
		weight = func(_ T) int64 { return 1 }
	}

	ceiling := capacity + workers
	maxWIP := opts.MaxWIP
	if maxWIP <= 0 {
		maxWIP = ceiling // default: no additional constraint
	}
	if maxWIP > ceiling {
		maxWIP = ceiling
	}
	if maxWIP < 1 {
		maxWIP = 1
	}

	stageCtx, cancel := context.WithCancelCause(ctx)

	s := &Stage[T, R]{
		in:          make(chan queued[T], capacity),
		out:         make(chan rslt.Result[R]),
		done:        make(chan struct{}),
		closing:     make(chan struct{}),
		stageDone:   stageCtx.Done(),
		cancel:      cancel,
		trackAllocs: opts.TrackAllocations && isAllocMetricsSupported(),
		capacity:    capacity,
		workers:     workers,
		maxWIP:      maxWIP,
		weight:      weight,
	}

	// INVARIANT: every goroutine that sends to s.out must be counted here.
	// close(s.out) happens only after wg.Wait returns. Adding a new sender
	// without incrementing this count causes a send-on-closed-channel race.
	senders := workers
	if feeder != nil {
		senders++
	}
	s.wg.Add(senders)

	for i := 0; i < workers; i++ {
		go s.worker(stageCtx, fn, failFast)
	}

	if feeder != nil {
		go feeder(s, stageCtx)
	}

	// Cancel watcher: ensures input closes on any cancellation
	// (parent cancel or fail-fast), so workers always exit.
	go func() {
		<-stageCtx.Done()
		s.CloseInput()
	}()

	// Closer: waits for all senders (workers + feeder), latches terminal
	// cause, cleans up. The stage is "complete" when close(s.done) fires.
	go func() {
		s.wg.Wait()

		// Latch terminal cause before signaling completion.
		s.errMu.Lock()
		if s.err != nil {
			s.cause = s.err
		} else if ctx.Err() != nil {
			s.cause = context.Cause(ctx)
		}
		s.errMu.Unlock()

		s.cancel(nil)
		close(s.out)
		close(s.done)
	}()

	return s
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

	return start(ctx, fn, opts, nil)
}

// Pipe creates a stage that reads from an upstream Result channel,
// forwarding Ok values to fn via workers and passing Err values directly
// to the output (error passthrough). The feeder goroutine drains src
// to completion, provided the consumer drains [Stage.Out] or ctx is
// canceled, and src eventually closes. Cancellation unblocks output
// sends but does not force-close src.
//
// Error passthrough is best-effort during shutdown: if ctx is canceled
// or fail-fast fires, upstream Err values may be dropped instead of
// forwarded (reflected in [Stats.Dropped]). During normal operation,
// all upstream errors are forwarded.
//
// The returned stage's input side is owned by the feeder — do not call
// [Stage.Submit] or [Stage.CloseInput] directly. Both are handled gracefully
// (no panic, no deadlock) but are misuse. External Submit calls void the
// stats invariant (Received will not account for externally submitted items).
//
// Stats: [Stats.Received] = [Stats.Submitted] + [Stats.Forwarded] + [Stats.Dropped].
// Forwarded errors do not trigger fail-fast and do not affect [Stage.Wait].
//
// Panics if ctx is nil, src is nil, or fn is nil.
func Pipe[T, R any](
	ctx context.Context,
	src <-chan rslt.Result[T],
	fn func(context.Context, T) (R, error),
	opts Options[T],
) *Stage[T, R] {
	if fn == nil {
		panic("toc.Pipe: fn must not be nil")
	}
	if src == nil {
		panic("toc.Pipe: src must not be nil")
	}

	feeder := func(s *Stage[T, R], stageCtx context.Context) {
		defer s.wg.Done()
		defer s.CloseInput()

		for r := range src {
			s.received.Add(1)

			if v, err := r.Unpack(); err != nil {
				// Error passthrough: forward to out, bypass workers.
				select {
				case s.out <- rslt.Err[R](err):
					s.forwarded.Add(1)
				case <-stageCtx.Done():
					s.dropped.Add(1)
				}
			} else {
				if submitErr := s.Submit(stageCtx, v); submitErr != nil {
					s.dropped.Add(1)
				}
			}
		}
	}

	return start(ctx, fn, opts, feeder)
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

	// Rope: acquire admission slot before sending to channel.
	if err := s.acquireAdmission(ctx); err != nil {
		return err
	}

	if err := s.trySend(ctx, q); err != nil {
		// Rollback: release the admission slot we acquired.
		s.releaseAdmission()

		return err
	}

	return nil
}

// acquireAdmission blocks until an admission slot is available or
// the context/stage is canceled. Returns nil on success, error on
// cancel/close.
func (s *Stage[T, R]) acquireAdmission(ctx context.Context) error {
	s.admissionMu.Lock()

	// Admission gate: closedForAdmission is the authoritative cutoff,
	// checked under admissionMu. No permit is ever created after this
	// point — both fast and slow paths are behind this check.
	if s.closedForAdmission {
		s.admissionMu.Unlock()

		return ErrClosed
	}

	// Fast path: slot available.
	if s.admitted < int64(s.maxWIP) {
		s.admitted++
		s.admissionMu.Unlock()

		if h := s.hooks; h != nil && h.afterAdmitFastPath != nil {
			h.afterAdmitFastPath()
		}

		return nil
	}

	// Slow path: enqueue a waiter and block.
	w := waiter{ready: make(chan struct{})}
	s.waiters = append(s.waiters, w)
	if len(s.waiters) > s.maxWaiterCount {
		s.maxWaiterCount = len(s.waiters)
	}
	s.admissionMu.Unlock()

	if h := s.hooks; h != nil && h.afterWaiterQueued != nil {
		h.afterWaiterQueued()
	}

	s.ropeWaitCnt.Add(1)
	waitStart := time.Now()

	select {
	case <-w.ready:
		// Slot granted by grantWaitersLocked (admitted already incremented).
		s.ropeWaitNs.Add(int64(time.Since(waitStart)))

		return nil

	case <-ctx.Done():
		s.ropeWaitNs.Add(int64(time.Since(waitStart)))

		return s.revokeOrHonor(w, ctx.Err())

	case <-s.closing:
		s.ropeWaitNs.Add(int64(time.Since(waitStart)))

		return s.revokeOrHonor(w, ErrClosed)

	case <-s.stageDone:
		s.ropeWaitNs.Add(int64(time.Since(waitStart)))

		return s.revokeOrHonor(w, ErrClosed)
	}
}

// revokeOrHonor attempts to remove a waiter from the queue.
// If still queued, removes it and returns rejectErr.
// If already granted, honors the grant (returns nil so the caller
// proceeds to trySend, which will fail and rollback).
//
// Key invariant: a waiter removed from the queue by grantWaitersLocked has
// irrevocably consumed a permit (admitted++ under admissionMu).
// w.ready is notification only — queue membership under admissionMu
// is the source of truth for slot ownership.
//
// Must NOT be called with admissionMu held.
func (s *Stage[T, R]) revokeOrHonor(w waiter, rejectErr error) error {
	if h := s.hooks; h != nil && h.beforeRevokeOrHonor != nil {
		h.beforeRevokeOrHonor()
	}

	s.admissionMu.Lock()
	if s.removeWaiter(w) {
		s.admissionMu.Unlock()

		return rejectErr
	}
	// Already granted — honor the grant, don't leak the slot.
	s.admissionMu.Unlock()

	return nil
}

// removeWaiter removes w from the waiters queue. Returns true if found
// and removed, false if already granted (not in queue).
// Must be called with admissionMu held.
func (s *Stage[T, R]) removeWaiter(w waiter) bool {
	for i, candidate := range s.waiters {
		if candidate.ready == w.ready {
			s.waiters = append(s.waiters[:i], s.waiters[i+1:]...)
			if len(s.waiters) == 0 {
				s.waiters = nil
			}

			return true
		}
	}

	return false
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

	s.bufferedDepth.Add(1)

	select {
	case s.in <- q:
		s.submitted.Add(1)

		return nil
	case <-s.closing:
		s.bufferedDepth.Add(-1)

		return ErrClosed
	case <-s.stageDone:
		s.bufferedDepth.Add(-1)

		return ErrClosed
	case <-ctx.Done():
		s.bufferedDepth.Add(-1)

		return ctx.Err()
	}
}

// CloseInput signals that no more items will be submitted.
// Workers finish processing buffered items, then shut down.
//
// Linearization point: closedForAdmission=true under admissionMu.
// Any Submit whose acquireAdmission serializes after this point
// returns ErrClosed without creating a permit. A Submit already
// granted before this point may still succeed or roll back via
// trySend — that is a concurrent operation that linearized before
// close, not a post-close admission.
//
// Close state is split across three signals:
//   - closedForAdmission (admissionMu): authoritative gate for permit creation
//   - closed (atomic): fast-path rejection in Submit and trySend
//   - closing (channel): shutdown broadcast to blocked selects
//
// These are set in order; closed/closing may lag closedForAdmission.
//
// Blocks briefly until all in-flight [Stage.Submit] calls exit, then
// closes the input channel.
//
// Idempotent — safe to call multiple times (use defer as safety net).
// Also called internally on fail-fast or parent context cancellation.
func (s *Stage[T, R]) CloseInput() {
	s.closeOnce.Do(func() {
		// Mark admission closed under admissionMu first — serializes
		// with grantWaitersLocked so no grants can happen after this point.
		s.admissionMu.Lock()
		s.closedForAdmission = true
		s.admissionMu.Unlock()

		if h := s.hooks; h != nil && h.afterCloseAdmission != nil {
			h.afterCloseAdmission()
		}

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
// The stage is complete when all workers have finished, terminal status
// is latched, and the done channel closes.
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

	s.admissionMu.Lock()
	admitted := s.admitted
	maxWIP := s.maxWIP
	waiterCount := len(s.waiters)
	maxWaiterCount := s.maxWaiterCount
	s.admissionMu.Unlock()

	return Stats{
		Submitted:            s.submitted.Load(),
		Completed:            s.completed.Load(),
		Failed:               s.failed.Load(),
		Panicked:             s.panicked.Load(),
		Canceled:             s.canceled.Load(),
		Received:             s.received.Load(),
		Forwarded:            s.forwarded.Load(),
		Dropped:              s.dropped.Load(),
		ServiceTime:          time.Duration(s.serviceNs.Load()),
		IdleTime:             time.Duration(s.idleNs.Load()),
		OutputBlockedTime:    time.Duration(s.outputBlockedNs.Load()),
		BufferedDepth:        depth,
		InFlightWeight:       s.inFlightWeight.Load(),
		QueueCapacity:        s.capacity,
		MaxWIP:               maxWIP,
		Admitted:             admitted,
		WaiterCount:          waiterCount,
		MaxWaiterCount:       maxWaiterCount,
		RopeWaitCount:        s.ropeWaitCnt.Load(),
		RopeWaitNs:           s.ropeWaitNs.Load(),
		AllocTrackingActive:  s.trackAllocs,
		ObservedAllocBytes:   s.allocBytes.Load(),
		ObservedAllocObjects: s.allocObjects.Load(),
	}
}

// grantWaitersLocked wakes blocked Submits when slots are available.
// Does not grant once closedForAdmission is set — waiters will wake via
// s.closing/s.stageDone channels and be rejected there.
// Must be called with admissionMu held. closedForAdmission is also
// protected by admissionMu, so this check has no TOCTOU race.
func (s *Stage[T, R]) grantWaitersLocked() {
	if s.closedForAdmission {
		return
	}

	for len(s.waiters) > 0 && s.admitted < int64(s.maxWIP) {
		w := s.waiters[0]
		s.waiters[0] = waiter{} // clear for GC
		s.waiters = s.waiters[1:]
		s.admitted++
		close(w.ready)
		if h := s.hooks; h != nil && h.onGrant != nil {
			h.onGrant() // UNDER LOCK — must not block. Use notifyNonBlocking.
		}
	}
	// Reclaim backing array when queue empties.
	if len(s.waiters) == 0 {
		s.waiters = nil
	}
}

// releaseAdmission decrements the admitted counter and wakes waiters.
// Called by workers on completion (including panic recovery).
func (s *Stage[T, R]) releaseAdmission() {
	s.admissionMu.Lock()
	s.admitted--
	s.grantWaitersLocked()
	s.admissionMu.Unlock()

	if h := s.hooks; h != nil && h.afterRelease != nil {
		h.afterRelease()
	}
}

// SetMaxWIP dynamically adjusts the WIP limit. Wakes blocked Submits
// if the new limit is higher than the current one. The value is clamped
// to [1, Capacity+Workers]. Returns the applied value.
// Concurrency-safe.
func (s *Stage[T, R]) SetMaxWIP(n int) int {
	ceiling := s.capacity + s.workers
	if n < 1 {
		n = 1
	}
	if n > ceiling {
		n = ceiling
	}

	s.admissionMu.Lock()
	s.maxWIP = n
	s.grantWaitersLocked()
	s.admissionMu.Unlock()

	return n
}

// MaxWIP returns the current WIP limit.
func (s *Stage[T, R]) MaxWIP() int {
	s.admissionMu.Lock()
	n := s.maxWIP
	s.admissionMu.Unlock()

	return n
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

	var samples [2]metrics.Sample
	if s.trackAllocs {
		samples[0].Name = metricAllocBytes
		samples[1].Name = metricAllocObjects
	}

	for {
		idleStart := time.Now()
		q, ok := <-s.in
		s.idleNs.Add(int64(time.Since(idleStart)))

		if !ok {
			return
		}

		s.processItem(ctx, fn, failFast, q, samples[:])
	}
}

// processItem handles a single dequeued item. The admission permit is
// released via defer, guaranteeing exactly-once release regardless of
// panics, errors, cancellation, or future code changes.
//
// Permit lifetime covers processing AND output publish (s.out <- result).
// If the consumer stops draining Out(), the permit is held until the
// send unblocks. This is the same liveness contract as the stage itself:
// consumer must drain Out() or cancel the context.
func (s *Stage[T, R]) processItem(
	ctx context.Context,
	fn func(context.Context, T) (R, error),
	failFast bool,
	q queued[T],
	samples []metrics.Sample,
) {
	defer s.releaseAdmission()

	s.bufferedDepth.Add(-1)

	// Check cancellation before calling fn.
	if err := ctx.Err(); err != nil {
		s.canceled.Add(1)

		outStart := time.Now()
		s.out <- rslt.Err[R](context.Cause(ctx))
		s.outputBlockedNs.Add(int64(time.Since(outStart)))

		return
	}

	s.inFlightWeight.Add(q.weight)

	var bytesBefore, objsBefore uint64
	if s.trackAllocs {
		metrics.Read(samples)
		bytesBefore = samples[0].Value.Uint64()
		objsBefore = samples[1].Value.Uint64()
	}

	serviceStart := time.Now()
	result := s.safeCall(ctx, fn, q.item)
	s.serviceNs.Add(int64(time.Since(serviceStart)))

	if s.trackAllocs {
		metrics.Read(samples)
		bytesAfter := samples[0].Value.Uint64()
		objsAfter := samples[1].Value.Uint64()
		if bytesAfter >= bytesBefore {
			s.allocBytes.Add(bytesAfter - bytesBefore)
		}
		if objsAfter >= objsBefore {
			s.allocObjects.Add(objsAfter - objsBefore)
		}
	}

	s.inFlightWeight.Add(-q.weight)
	s.completed.Add(1)

	if _, err := result.Unpack(); err != nil {
		s.failed.Add(1)

		if failFast {
			s.errMu.Lock()
			first := s.err == nil && ctx.Err() == nil
			if first {
				s.err = err
			}
			s.errMu.Unlock()

			if first {
				s.cancel(err)
				s.CloseInput()
			}
		}
	}

	outStart := time.Now()
	s.out <- result
	s.outputBlockedNs.Add(int64(time.Since(outStart)))
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
