package wrap_test

import . "github.com/binaryphile/fluentfp/wrap"

// Compile-time export verification.
func _() {
	// Circuit breaker infrastructure
	_ = ErrCircuitOpen
	_ = NewBreaker
	_ = ConsecutiveFailures
	_ = BreakerConfig{}
	_ = BreakerState(0)
	_ = Snapshot{}
	_ = Transition{}

	// Backoff
	_ = Backoff(nil)
	_ = ExpBackoff

	// Entry point and types
	_ = Fn[int, int](nil)
	_ = Func[int, int]
	_ = Decorator[int, int](nil)
	_ = Features{}
	_ = RetryConfig{}
	_ = Retry
	_ = ThrottleConfig{}
	_ = Throttle

	// With (Features-based) and Apply (custom decorators)
	_ = Fn[int, int].With
	_ = Fn[int, int].Apply
	_ = WithThrottleWeighted[int, int]
}
