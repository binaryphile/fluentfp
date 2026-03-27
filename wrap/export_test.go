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

	// Entry point and composition
	_ = Fn[int, int](nil)
	_ = Func[int, int]
	_ = Decorator[int, int](nil)

	// With* methods (verified via Fn method set)
	_ = Fn[int, int].WithRetry
	_ = Fn[int, int].WithBreaker
	_ = Fn[int, int].WithThrottle
	_ = Fn[int, int].WithThrottleWeighted
	_ = Fn[int, int].WithMapError
	_ = Fn[int, int].WithOnError
	_ = Fn[int, int].With
}
