package wrap

import "context"

// Fn is the function shape all decorators operate on:
// a context-aware function that returns a value or an error.
type Fn[T, R any] func(context.Context, T) (R, error)

// Func wraps a plain function as an Fn for fluent decoration.
// Go infers the type parameters from fn:
//
//	wrap.Func(fetchUser).With(wrap.Features{
//	    Retry: wrap.Retry(3, wrap.ExpBackoff(time.Second), nil),
//	})
func Func[T, R any](fn func(context.Context, T) (R, error)) Fn[T, R] {
	return fn
}

// Decorator wraps an Fn, returning an Fn with the same signature.
// Use with [Fn.Apply] for custom decorators not covered by [Features].
type Decorator[T, R any] func(Fn[T, R]) Fn[T, R]

// RetryConfig configures the retry feature.
type RetryConfig struct {
	Backoff     Backoff
	Max         int
	ShouldRetry func(error) bool
}

// Retry returns a RetryConfig for use in [Features].
func Retry(max int, backoff Backoff, shouldRetry func(error) bool) *RetryConfig {
	return &RetryConfig{Max: max, Backoff: backoff, ShouldRetry: shouldRetry}
}

// ThrottleConfig configures count-based concurrency control.
type ThrottleConfig struct {
	N int
}

// Throttle returns a ThrottleConfig for use in [Features].
func Throttle(n int) *ThrottleConfig {
	return &ThrottleConfig{N: n}
}

// Features configures which decorators to apply. Nil fields are skipped.
//
// Decorators are applied in a fixed order (innermost to outermost):
// OnError → MapError → Retry → Breaker → Throttle.
type Features struct {
	Breaker  *Breaker
	MapError func(error) error
	OnError  func(error)
	Retry    *RetryConfig
	Throttle *ThrottleConfig
}

// With applies the features to f in a fixed order. Nil fields are skipped.
// Innermost to outermost: OnError → MapError → Retry → Breaker → Throttle.
func (f Fn[T, R]) With(feat Features) Fn[T, R] {
	if feat.OnError != nil {
		f = Fn[T, R](onErr(f, feat.OnError))
	}

	if feat.MapError != nil {
		f = Fn[T, R](mapErr(f, feat.MapError))
	}

	if feat.Retry != nil {
		f = Fn[T, R](retry(feat.Retry.Max, feat.Retry.Backoff, feat.Retry.ShouldRetry, f))
	}

	if feat.Breaker != nil {
		f = Fn[T, R](withBreaker(feat.Breaker, f))
	}

	if feat.Throttle != nil {
		f = Fn[T, R](throttle(feat.Throttle.N, f))
	}

	return f
}

// Apply applies custom decorators to f in order (innermost-first).
// Use for decorators not covered by [Features], such as [WithThrottleWeighted]
// or application-specific wrappers.
func (f Fn[T, R]) Apply(ds ...Decorator[T, R]) Fn[T, R] {
	for _, d := range ds {
		if d == nil {
			panic("wrap.Apply: nil decorator")
		}
		f = Fn[T, R](d(f))
	}
	return f
}

// WithThrottleWeighted returns a Decorator for cost-based concurrency control.
// Not available via Features because the cost function requires type T.
func WithThrottleWeighted[T, R any](capacity int, cost func(T) int) Decorator[T, R] {
	return func(f Fn[T, R]) Fn[T, R] {
		return Fn[T, R](throttleWeighted(capacity, cost, f))
	}
}
