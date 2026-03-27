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

	// Methods on Fn
	_ = Fn[int, int].Breaker
	_ = Fn[int, int].MapError
	_ = Fn[int, int].OnError
	_ = Fn[int, int].Retry
	_ = Fn[int, int].Apply
}
