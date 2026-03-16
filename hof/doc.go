// Package hof provides function combinators for composition, partial application,
// independent application, concurrency control, side-effect wrapping, and
// call coalescing. Based on Stone's "Algorithms: A Functional Programming
// Approach" (pipe, sect, cross).
package hof

// Compile-time export verification. Every fluentfp package uses this pattern
// to ensure exported symbols remain available across refactors.
func _() {
	// Composition
	_ = Pipe[int, int, int]

	// Partial application
	_ = Bind[int, int, int]
	_ = BindR[int, int, int]

	// Independent application
	_ = Cross[int, int, int, int]

	// Building blocks
	_ = Eq[int]

	// Concurrency control
	_ = Throttle[int, int]
	_ = ThrottleWeighted[int, int]

	// Side-effect wrappers
	_ = OnErr[int, int]

	// Retry
	_ = Retry[int, int]
	_ = Backoff(nil)
	_ = ConstantBackoff
	_ = ExponentialBackoff

	// Circuit breaking
	_ = NewBreaker
	_ = WithBreaker[int, int]
	_ = ConsecutiveFailures
	_ = ErrCircuitOpen
	_ = BreakerState(0)
	_ = StateClosed
	_ = StateOpen
	_ = StateHalfOpen
	_ = Snapshot{}
	_ = Transition{}
	_ = BreakerConfig{}

	// Debouncing
	_ = NewDebouncer[int]
	_ = MaxWait
}
