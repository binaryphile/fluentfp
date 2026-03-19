// Package cb provides resilience decorators for communicating with runtime
// dependencies. All decorators wrap func(context.Context, T) (R, error),
// preserving the signature so they compose freely: stack retry inside a
// breaker, throttle inside retry, etc.
//
// Named after citizen band radio — communication over unreliable channels.
// "Breaker, breaker" is literally the metaphor.
package cb

import "time"

// Compile-time export verification.
func _() {
	// Circuit breaker
	_ = ErrOpen
	_ = NewBreaker
	_ = WithBreaker[int, int]
	_ = ConsecutiveFailures
	_ = BreakerConfig{}
	_ = BreakerState(0)
	_ = Snapshot{}
	_ = Transition{}

	// Retry
	_ = Retry[int, int]
	_ = Backoff(nil)
	_ = ConstantBackoff
	_ = ExponentialBackoff

	// Throttle
	_ = Throttle[int, int]
	_ = ThrottleWeighted[int, int]

	// Side-effect wrappers
	_ = OnErr[int, int]
	_ = MapErr[int, int]

	// Debounce
	_ = NewDebouncer[int]
	_ = MaxWait
	_ = DebounceOption(nil)

	// Suppress unused import.
	_ = time.Second
}
