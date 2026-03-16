package hof

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrCircuitOpen is returned when the circuit breaker is rejecting requests.
// This occurs when the breaker is open, or when half-open with a probe already in flight.
var ErrCircuitOpen = errors.New("hof: circuit breaker is open")

// BreakerState represents the current state of a circuit breaker.
type BreakerState int

const (
	StateClosed   BreakerState = iota
	StateOpen
	StateHalfOpen
)

func (s BreakerState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// Snapshot is a point-in-time view of breaker state and metrics.
// Successes and Failures reset when the breaker transitions to closed.
// ConsecutiveFailures resets on any success (including while closed).
// Rejected is a lifetime counter.
// OpenedAt is the zero time when State is StateClosed.
type Snapshot struct {
	State               BreakerState
	Successes           int
	Failures            int
	ConsecutiveFailures int
	Rejected            int
	OpenedAt            time.Time
}

// Transition describes a circuit breaker state change.
type Transition struct {
	From BreakerState
	To   BreakerState
	At   time.Time
}

// BreakerConfig configures a circuit breaker.
type BreakerConfig struct {
	// ResetTimeout is how long the breaker stays open before allowing a probe request.
	// Must be > 0.
	ResetTimeout time.Duration

	// ReadyToTrip decides whether the breaker should open based on current metrics.
	// Called outside the internal lock after each counted failure while closed.
	// The snapshot reflects the state including the current failure.
	//
	// Under concurrency, the breaker validates that no metric mutations occurred
	// between ReadyToTrip evaluation and the trip commit. If metrics changed
	// (concurrent success or failure), the trip is aborted; the next failure
	// will re-evaluate with a fresh snapshot. This means predicates should be
	// monotone with respect to failure accumulation (e.g., >= threshold) for
	// reliable tripping under contention. Non-monotone predicates (e.g., == N)
	// may miss a trip if a concurrent mutation changes the count between
	// evaluation and commit.
	//
	// Must be side-effect-free. May be called concurrently from multiple goroutines.
	// Nil defaults to ConsecutiveFailures(5).
	ReadyToTrip func(Snapshot) bool

	// ShouldCount decides whether an error counts as a failure for trip purposes.
	// Called outside the internal lock.
	// Nil means all errors count. context.Canceled never counts regardless of this setting.
	ShouldCount func(error) bool

	// OnStateChange is called after each state transition on the normal (non-panic) path,
	// outside the internal lock. Transitions caused by panic recovery (e.g., a half-open
	// probe fn panic reopening the breaker) do not trigger the callback to avoid masking
	// the original panic.
	// Under concurrency, callback delivery may lag or overlap and should not
	// be treated as a total order. Panics in this callback propagate to the caller.
	// Nil means no notification.
	OnStateChange func(Transition)

	// Clock returns the current time. Nil defaults to time.Now.
	// Must be non-blocking, must not panic, and must not call Breaker methods
	// (deadlock risk). Useful for deterministic testing.
	Clock func() time.Time
}

// ConsecutiveFailures returns a ReadyToTrip predicate that trips after n consecutive failures.
// Panics if n < 1.
func ConsecutiveFailures(n int) func(Snapshot) bool {
	if n < 1 {
		panic("hof.ConsecutiveFailures: n must be > 0")
	}

	return func(s Snapshot) bool {
		return s.ConsecutiveFailures >= n
	}
}

// Breaker is a circuit breaker that tracks failures and short-circuits requests
// when a dependency is unhealthy. Use NewBreaker to create and WithBreaker to
// wrap functions for composition with Retry, Throttle, and other hof wrappers.
//
// The breaker uses a standard three-state model:
//   - Closed: requests pass through, failures are counted
//   - Open: requests fail immediately with ErrCircuitOpen
//   - HalfOpen: one probe request is admitted; success closes, failure reopens,
//     uncounted error (context.Canceled or ShouldCount→false) releases the probe
//     slot without changing state
//
// State transitions are lazy (checked on admission, not timer-driven).
// One probe request is admitted in half-open; all others are rejected.
//
// Each state transition increments an internal generation counter. Calls that
// complete after the breaker has moved to a new generation are silently ignored,
// preventing stale in-flight results from corrupting the current epoch's metrics.
//
// A Breaker must represent a single dependency or failure domain. Sharing a
// breaker across unrelated dependencies causes pathological coupling: one
// dependency's failures can trip the breaker for all, and one dependency's
// successful probe can close it while others remain unhealthy.
type Breaker struct {
	mu               sync.Mutex
	state            BreakerState
	generation       uint64
	metricsVersion   uint64
	successes        int
	failures         int
	consecutiveFails int
	rejected         int
	openedAt         time.Time
	probeInFlight    bool

	resetTimeout  time.Duration
	readyToTrip   func(Snapshot) bool
	shouldCount   func(error) bool
	onStateChange func(Transition)
	now           func() time.Time
}

// NewBreaker creates a circuit breaker with the given configuration.
// Panics if ResetTimeout <= 0.
func NewBreaker(cfg BreakerConfig) *Breaker {
	if cfg.ResetTimeout <= 0 {
		panic("hof.NewBreaker: ResetTimeout must be > 0")
	}

	readyToTrip := cfg.ReadyToTrip
	if readyToTrip == nil {
		readyToTrip = ConsecutiveFailures(5)
	}

	now := cfg.Clock
	if now == nil {
		now = time.Now
	}

	return &Breaker{
		resetTimeout:  cfg.ResetTimeout,
		readyToTrip:   readyToTrip,
		shouldCount:   cfg.ShouldCount,
		onStateChange: cfg.OnStateChange,
		now:           now,
	}
}

// Snapshot returns a point-in-time view of the breaker's state and metrics.
// State is the committed state; lazy transitions (open to half-open after
// resetTimeout) are not reflected until the next admission check.
func (b *Breaker) Snapshot() Snapshot {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.snapshotLocked()
}

func (b *Breaker) snapshotLocked() Snapshot {
	return Snapshot{
		State:               b.state,
		Successes:           b.successes,
		Failures:            b.failures,
		ConsecutiveFailures: b.consecutiveFails,
		Rejected:            b.rejected,
		OpenedAt:            b.openedAt,
	}
}

// WithBreaker wraps fn with circuit breaker protection from b.
// The returned function has the standard hof signature for composition with
// Retry, Throttle, and other wrappers.
//
// If ctx is already cancelled or expired before admission, the error is returned
// immediately without affecting breaker state or metrics.
//
// Context cancellation (context.Canceled) does not count as a failure.
// context.DeadlineExceeded counts as a failure by default (controllable via ShouldCount).
// Errors where ShouldCount returns false and context.Canceled do not break
// the consecutive-failure streak; only a success resets it.
//
// If fn panics during a half-open probe, the breaker records a failure and
// reopens before the panic propagates. Panics during closed-state calls do not
// affect breaker state.
//
// Panics if b or fn is nil.
func WithBreaker[T, R any](b *Breaker, fn func(context.Context, T) (R, error)) func(context.Context, T) (R, error) {
	if b == nil {
		panic("hof.WithBreaker: breaker must not be nil")
	}
	if fn == nil {
		panic("hof.WithBreaker: fn must not be nil")
	}

	return func(ctx context.Context, t T) (R, error) {
		var zero R

		if err := ctx.Err(); err != nil {
			return zero, err
		}

		// Admission check (defer-protected against Clock panics).
		gen, admitted, isProbe, admitTransition := b.admit()
		if !admitted {
			return zero, ErrCircuitOpen
		}

		// Probe finalization guard: ensures the probe slot is cleaned up
		// if the result is not recorded before exit. Covers:
		//   - fn panics → records failure to reopen breaker
		//   - ShouldCount panics → releases probe slot
		//   - OnStateChange panics on admission → records failure to reopen
		var (
			resultRecorded bool
			fnReturned     bool
		)
		if isProbe {
			defer func() {
				if resultRecorded {
					return
				}

				if fnReturned {
					// Post-fn callback panic — release probe slot.
					b.releaseProbe(gen)
				} else {
					// fn panic or pre-fn callback panic — record failure.
					b.recordFailureForGen(gen)
				}
			}()
		}

		if admitTransition != nil {
			b.fireOnStateChange(*admitTransition)
		}

		r, err := fn(ctx, t)
		fnReturned = true

		transition := b.recordCompletion(gen, isProbe, err)
		resultRecorded = true

		if transition != nil {
			b.fireOnStateChange(*transition)
		}

		return r, err
	}
}

// admit performs the admission check with self-locking.
func (b *Breaker) admit() (gen uint64, admitted, isProbe bool, transition *Transition) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.admitLocked()
}

// admitLocked checks whether a request should be allowed through.
// Must be called with b.mu held.
func (b *Breaker) admitLocked() (gen uint64, admitted, isProbe bool, transition *Transition) {
	switch b.state {
	case StateClosed:
		return b.generation, true, false, nil

	case StateOpen:
		now := b.now()
		if now.Before(b.openedAt.Add(b.resetTimeout)) {
			b.rejected++

			return 0, false, false, nil
		}

		// Timeout elapsed — transition to half-open, admit as probe.
		b.generation++
		t := Transition{From: StateOpen, To: StateHalfOpen, At: now}
		b.state = StateHalfOpen
		b.probeInFlight = true

		return b.generation, true, true, &t

	case StateHalfOpen:
		if b.probeInFlight {
			b.rejected++

			return 0, false, false, nil
		}

		b.probeInFlight = true

		return b.generation, true, true, nil
	}

	return 0, false, false, nil
}

// recordCompletion handles all result-recording logic for a completed call.
// ShouldCount is called outside any lock; panics from it are caught by the
// caller's probe finalization guard.
func (b *Breaker) recordCompletion(gen uint64, isProbe bool, err error) *Transition {
	if err == nil {
		return b.recordSuccessForGen(gen)
	}

	// context.Canceled never counts as a failure.
	if errors.Is(err, context.Canceled) {
		if isProbe {
			b.releaseProbe(gen)
		}

		return nil
	}

	// ShouldCount filter for other errors.
	if b.shouldCount != nil && !b.shouldCount(err) {
		if isProbe {
			b.releaseProbe(gen)
		}

		return nil
	}

	return b.recordFailureForGen(gen)
}

// recordSuccessForGen records a success if generation matches.
// Stale completions (gen mismatch) are silently ignored.
func (b *Breaker) recordSuccessForGen(gen uint64) *Transition {
	b.mu.Lock()
	defer b.mu.Unlock()

	if gen != b.generation {
		return nil
	}

	return b.recordSuccessLocked()
}

// recordFailureForGen records a failure if generation matches.
// ReadyToTrip is called outside the lock to avoid reentrancy/deadlock risk.
// Uses metricsVersion to detect TOCTOU: if metrics changed between ReadyToTrip
// and the trip attempt (e.g., a concurrent success reset consecutiveFails),
// the trip is aborted. The next failure will re-evaluate with a fresh snapshot.
// Stale completions (gen mismatch) are silently ignored.
func (b *Breaker) recordFailureForGen(gen uint64) *Transition {
	transition, snap, mv, checkTrip := b.applyFailure(gen)
	if transition != nil || !checkTrip {
		return transition
	}

	// ReadyToTrip called outside the lock — safe for blocking/reentrancy.
	if !b.readyToTrip(snap) {
		return nil
	}

	return b.tripIfCurrent(gen, mv)
}

// applyFailure increments failure counters and handles the half-open→open transition.
// For the closed case, returns the snapshot and metricsVersion for the caller to
// check ReadyToTrip outside the lock.
func (b *Breaker) applyFailure(gen uint64) (transition *Transition, snap Snapshot, mv uint64, checkTrip bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if gen != b.generation {
		return nil, Snapshot{}, 0, false
	}

	b.failures++
	b.consecutiveFails++

	switch b.state {
	case StateHalfOpen:
		now := b.now()
		b.generation++
		t := Transition{From: StateHalfOpen, To: StateOpen, At: now}
		b.state = StateOpen
		b.openedAt = now
		b.probeInFlight = false

		return &t, Snapshot{}, 0, false

	case StateClosed:
		b.metricsVersion++

		return nil, b.snapshotLocked(), b.metricsVersion, true
	}

	return nil, Snapshot{}, 0, false
}

// tripIfCurrent transitions closed→open if generation and metricsVersion still match.
// If either changed (concurrent success/failure altered metrics or state), the trip
// is aborted — the next failure's own recordFailureForGen will re-evaluate with
// a fresh snapshot. Abort-on-mismatch avoids re-evaluation loops (which would change
// callback semantics and risk starvation).
func (b *Breaker) tripIfCurrent(gen, mv uint64) *Transition {
	b.mu.Lock()
	defer b.mu.Unlock()

	if gen != b.generation || b.state != StateClosed || b.metricsVersion != mv {
		return nil
	}

	now := b.now()
	b.generation++
	t := Transition{From: StateClosed, To: StateOpen, At: now}
	b.state = StateOpen
	b.openedAt = now

	return &t
}

// releaseProbe releases the half-open probe slot if generation matches.
func (b *Breaker) releaseProbe(gen uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if gen == b.generation && b.probeInFlight {
		b.probeInFlight = false
	}
}

// recordSuccessLocked records a successful call and returns any state transition.
// Must be called with b.mu held.
func (b *Breaker) recordSuccessLocked() *Transition {
	b.successes++
	b.consecutiveFails = 0
	b.metricsVersion++

	if b.state == StateHalfOpen {
		now := b.now()
		b.generation++
		t := Transition{From: StateHalfOpen, To: StateClosed, At: now}
		b.state = StateClosed
		b.probeInFlight = false

		// Reset counts and openedAt for the new closed period.
		b.successes = 0
		b.failures = 0
		b.openedAt = time.Time{}

		return &t
	}

	return nil
}

func (b *Breaker) fireOnStateChange(t Transition) {
	if b.onStateChange != nil {
		b.onStateChange(t)
	}
}
