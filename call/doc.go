// Package call provides decorators for context-aware effectful functions.
//
// Every decorator in this package wraps func(context.Context, T) (R, error)
// and returns the same signature. This uniform shape is the organizing
// principle: decorators compose by stacking because the types match at
// every layer.
//
// For higher-order functions over plain signatures — func(A) B composition,
// partial application, debouncing — see the [hof] package. The seam between
// call and hof is the function signature: call operates on the context-aware
// error-returning call shape; hof operates on everything else.
package call

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
	_ = NewFunc[int, int]
	_ = Decorator[int, int](nil)
	_ = CircuitBreaker[int, int]
	_ = Retrier[int, int]
	_ = Throttler[int, int]
	_ = ThrottlerWeighted[int, int]
	_ = ErrMapper[int, int]
	_ = OnError[int, int]
}
