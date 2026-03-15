package hof

import (
	"sync"
	"time"
)

// debounceConfig holds optional configuration for a Debouncer.
type debounceConfig struct {
	maxWait time.Duration
}

// DebounceOption configures a Debouncer.
type DebounceOption func(*debounceConfig)

// MaxWait caps the maximum delay under continuous activity.
// When continuous calls keep resetting the trailing timer, MaxWait guarantees
// execution after this duration from the first call in a burst.
// Zero (default) means no cap — trailing edge only, which can defer
// indefinitely under continuous activity.
// Panics if d < 0.
func MaxWait(d time.Duration) DebounceOption {
	if d < 0 {
		panic("hof.MaxWait: duration must be >= 0")
	}

	return func(c *debounceConfig) {
		c.maxWait = d
	}
}

// Debouncer coalesces rapid calls, executing fn with the latest value
// after a quiet period of at least wait. At most one fn execution runs
// at a time; calls during execution queue the latest value for a fresh
// timer cycle after completion.
//
// A single owner goroutine manages all state — no mutex contention,
// no stale timer callbacks. Call, Cancel, and Flush communicate via
// channels; the owner processes events sequentially.
//
// Value capture: Call stores the latest T by value. No deep copy is
// performed. If T contains pointers, slices, or maps, the caller must
// not mutate their contents after Call.
//
// Panic behavior: fn runs in a spawned goroutine. If fn panics, the
// owner goroutine's state is preserved via deferred completion signaling,
// and the panic propagates normally (typically crashing the process).
//
// Reentrancy: Call and Cancel are safe to invoke from within fn on the
// same Debouncer. Flush and Close from within fn will deadlock — fn
// completion must signal before either can proceed.
//
// Close must be called when the Debouncer is no longer needed to stop
// the owner goroutine. Use-after-Close panics. Close is idempotent.
// Operations concurrent with Close may block until Close completes,
// then panic.
type Debouncer[T any] struct {
	callCh    chan T
	cancelCh  chan chan bool
	flushCh   chan chan bool
	closeCh   chan struct{}
	done      chan struct{}
	closeOnce sync.Once
}

// NewDebouncer creates a trailing-edge debouncer that executes fn with
// the latest value after wait elapses with no new calls.
// Panics if wait <= 0 or fn is nil.
func NewDebouncer[T any](wait time.Duration, fn func(T), opts ...DebounceOption) *Debouncer[T] {
	if wait <= 0 {
		panic("hof.NewDebouncer: wait must be > 0")
	}
	if fn == nil {
		panic("hof.NewDebouncer: fn must not be nil")
	}

	var cfg debounceConfig

	for _, o := range opts {
		o(&cfg)
	}

	d := &Debouncer[T]{
		callCh:   make(chan T),
		cancelCh: make(chan chan bool),
		flushCh:  make(chan chan bool),
		closeCh:  make(chan struct{}),
		done:     make(chan struct{}),
	}

	go d.run(wait, cfg.maxWait, fn)

	return d
}

// Call schedules fn with v. If a previous call is pending, its value is
// replaced with v and the trailing timer resets. If fn is currently
// executing, v is queued for a fresh timer cycle after completion.
func (d *Debouncer[T]) Call(v T) {
	select {
	case <-d.done:
		panic("hof.Debouncer: Call after Close")
	case d.callCh <- v:
	}
}

// Cancel stops any pending execution. Returns true if pending work was
// canceled, false if there was nothing pending.
// If a Flush is blocked waiting for pending work, Cancel unblocks it
// and the Flush returns false.
func (d *Debouncer[T]) Cancel() bool {
	resp := make(chan bool, 1)

	select {
	case <-d.done:
		panic("hof.Debouncer: Cancel after Close")
	case d.cancelCh <- resp:
	}

	return <-resp
}

// Flush executes pending work immediately. Returns true if fn was
// executed as a result of this call, false if there was nothing pending.
//
// When fn is already running with pending work queued, Flush blocks until
// the current fn completes and the pending work executes. New Calls that
// arrive during a flushed execution do not extend the Flush — they are
// scheduled normally via timer after Flush returns.
//
// Only one Flush waiter is supported at a time. If a Flush is already
// waiting, subsequent Flush calls return false immediately.
//
// Flush must not be called from within fn on the same Debouncer — this
// will deadlock because fn completion must signal before Flush can proceed.
func (d *Debouncer[T]) Flush() bool {
	resp := make(chan bool, 1)

	select {
	case <-d.done:
		panic("hof.Debouncer: Flush after Close")
	case d.flushCh <- resp:
	}

	return <-resp
}

// Close stops the owner goroutine. Any pending work is discarded.
// If fn is currently executing, Close waits for it to complete.
// If a Flush triggered the currently running execution, Flush returns
// true (the execution completes). If a Flush is waiting for pending
// work that Close discards, Flush returns false.
// Close is idempotent — subsequent calls return immediately.
// After Close, Call, Cancel, and Flush will panic.
// Operations concurrent with Close may block until Close completes.
//
// Close must not be called from within fn on the same Debouncer — this
// will deadlock because fn completion must signal before Close can proceed.
func (d *Debouncer[T]) Close() {
	d.closeOnce.Do(func() {
		select {
		case <-d.done:
			// already closed (shouldn't happen with Once, but defensive)
		case d.closeCh <- struct{}{}:
		}

		<-d.done
	})
}

// run is the owner goroutine. It owns all mutable state and processes
// events sequentially via select.
func (d *Debouncer[T]) run(wait, maxWait time.Duration, fn func(T)) {
	defer close(d.done)

	var (
		value   T
		pending bool
		running bool

		trailTimer *time.Timer
		maxTimer   *time.Timer

		// flushWaiter is non-nil when a Flush caller is waiting for
		// an execution to complete — either one it triggered directly
		// (from waiting state) or one spawned on its behalf (from
		// running+pending state).
		//
		// Invariants (maintained by owner goroutine):
		//   flushTarget => flushWaiter != nil && running
		//   flushWaiter == nil => flushTarget == false
		//   After responding to flushWaiter, both are cleared together.
		flushWaiter chan<- bool

		// flushTarget is true when the currently running execution is
		// the one the flushWaiter is waiting for. When false and
		// flushWaiter is set, the current execution was not triggered
		// by Flush — the waiter is waiting for pending work to be
		// spawned after this execution completes.
		flushTarget bool

		// doneCh receives when a spawned fn goroutine completes.
		doneCh = make(chan struct{}, 1)

		// Channel accessors — nil when timer is not active, enabling
		// select to skip inactive timer cases.
		trailC <-chan time.Time
		maxC   <-chan time.Time
	)

	// stopTrail stops the trailing timer and clears its select case.
	stopTrail := func() {
		if trailTimer != nil {
			if !trailTimer.Stop() {
				select {
				case <-trailTimer.C:
				default:
				}
			}

			trailC = nil
		}
	}

	// stopMax stops the maxWait timer and clears its select case.
	stopMax := func() {
		if maxTimer != nil {
			if !maxTimer.Stop() {
				select {
				case <-maxTimer.C:
				default:
				}
			}

			maxC = nil
		}
	}

	// startTrail starts or resets the trailing timer.
	startTrail := func() {
		if trailTimer == nil {
			trailTimer = time.NewTimer(wait)
		} else {
			trailTimer.Reset(wait)
		}

		trailC = trailTimer.C
	}

	// startMax starts or resets the maxWait timer.
	startMax := func() {
		if maxWait <= 0 {
			return
		}

		if maxTimer == nil {
			maxTimer = time.NewTimer(maxWait)
		} else {
			maxTimer.Reset(maxWait)
		}

		maxC = maxTimer.C
	}

	// spawn starts fn in a goroutine with deferred completion signal.
	spawn := func(v T) {
		running = true

		go func() {
			defer func() { doneCh <- struct{}{} }()
			fn(v)
		}()
	}

	// fire handles a timer expiration — either trailing or maxWait.
	fire := func() {
		stopTrail()
		stopMax()

		captured := value

		var zero T

		value = zero
		pending = false

		spawn(captured)
	}

	for {
		select {
		case v := <-d.callCh:
			if running {
				// running or running+pending: queue for after completion.
				value = v
				pending = true
			} else if pending {
				// waiting: update value, reset trailing timer.
				value = v
				stopTrail()
				startTrail()
			} else {
				// idle → waiting.
				value = v
				pending = true

				startTrail()
				startMax()
			}

		case resp := <-d.cancelCh:
			if pending && !running {
				// waiting → idle.
				stopTrail()
				stopMax()

				var zero T

				value = zero
				pending = false
				resp <- true
			} else if pending && running {
				// running+pending → running.
				var zero T

				value = zero
				pending = false

				if flushWaiter != nil && !flushTarget {
					// Flush was waiting for pending work that was just canceled.
					flushWaiter <- false
					flushWaiter = nil
				}
				// If flushTarget is true, the flushed execution is already running
				// and will complete normally — flushWaiter gets true on doneCh.

				resp <- true
			} else {
				resp <- false
			}

		case resp := <-d.flushCh:
			if pending && !running {
				// waiting → running: execute immediately.
				stopTrail()
				stopMax()

				captured := value

				var zero T

				value = zero
				pending = false
				flushWaiter = resp
				flushTarget = true

				spawn(captured)
			} else if pending && running && flushWaiter == nil {
				// running+pending: register waiter for pending work.
				flushWaiter = resp
				flushTarget = false
			} else {
				// idle, running (no pending), or flush slot taken.
				resp <- false
			}

		case <-trailC:
			fire()

		case <-maxC:
			fire()

		case <-doneCh:
			running = false

			if flushWaiter != nil && flushTarget {
				// The execution Flush triggered has completed.
				flushWaiter <- true
				flushWaiter = nil
				flushTarget = false

				if pending {
					// New work arrived during flushed execution.
					// Schedule via normal timer cycle.
					startTrail()
					startMax()
				}
			} else if pending {
				if flushWaiter != nil {
					// Flush is waiting for pending work (flushTarget is false).
					// Execute pending immediately — this IS the target.
					captured := value

					var zero T

					value = zero
					pending = false
					flushTarget = true

					spawn(captured)
				} else {
					// Start fresh timer cycle for pending value.
					startTrail()
					startMax()
				}
			} else if flushWaiter != nil {
				// Unreachable if invariants hold: flushWaiter with
				// !flushTarget requires pending (set by Flush in
				// running+pending), and Cancel always nils flushWaiter
				// when clearing pending in that state. Kept as a
				// safety net — silent corruption is worse than a
				// spurious false.
				flushWaiter <- false
				flushWaiter = nil
				flushTarget = false
			}

		case <-d.closeCh:
			stopTrail()
			stopMax()

			pending = false

			var zero T

			value = zero

			if running {
				// If Flush is waiting for pending work (not the running
				// execution), respond false now — the pending was discarded.
				if flushWaiter != nil && !flushTarget {
					flushWaiter <- false
					flushWaiter = nil
				}

				<-doneCh

				// If the flushed execution was running, it completed.
				if flushWaiter != nil {
					flushWaiter <- true
					flushWaiter = nil
				}
			} else if flushWaiter != nil {
				flushWaiter <- false
				flushWaiter = nil
			}

			return
		}
	}
}
