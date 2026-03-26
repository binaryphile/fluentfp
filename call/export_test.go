package call_test

import . "github.com/binaryphile/fluentfp/call"

// Compile-time export verification.
func _() {
	// Circuit breaker
	_ = ErrCircuitOpen
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

	// Decorator composition
	_ = Func[int, int](nil)
	_ = From[int, int]
	_ = Decorator[int, int](nil)
	_ = Bracket[int, int]
	_ = CircuitBreaker[int, int]
	_ = Retrier[int, int]
	_ = Throttler[int, int]
	_ = ThrottlerWeighted[int, int]
	_ = ErrMapper[int, int]
	_ = OnError[int, int]
}
