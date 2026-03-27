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
// For custom decorators not covered by [Features], apply manually:
//
//	decorated := wrap.Fn[T, R](myDecorator(fn))
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

// Features configures which decorators to apply. Nil/zero fields are skipped.
//
// Decorators are applied in a fixed order (innermost to outermost):
// OnError → MapError → Retry → Breaker → Throttle.
type Features struct {
	Breaker  *Breaker
	MapError func(error) error
	OnError  func(error)
	Retry    *RetryConfig
	Throttle int
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

	if feat.Throttle > 0 {
		f = Fn[T, R](throttle(feat.Throttle, f))
	}

	return f
}

// WithRetry is shorthand for With(Features{Retry: Retry(...)}).
func (f Fn[T, R]) WithRetry(maxAttempts int, backoff Backoff, shouldRetry func(error) bool) Fn[T, R] {
	return Fn[T, R](retry(maxAttempts, backoff, shouldRetry, f))
}

// WithBreaker is shorthand for With(Features{Breaker: b}).
func (f Fn[T, R]) WithBreaker(b *Breaker) Fn[T, R] {
	return Fn[T, R](withBreaker(b, f))
}

// WithThrottle is shorthand for With(Features{Throttle: Throttle(n)}).
func (f Fn[T, R]) WithThrottle(n int) Fn[T, R] {
	return Fn[T, R](throttle(n, f))
}

// WithThrottleWeighted wraps f with cost-based concurrency control.
// Not available via Features because the cost function requires type T.
func (f Fn[T, R]) WithThrottleWeighted(capacity int, cost func(T) int) Fn[T, R] {
	return Fn[T, R](throttleWeighted(capacity, cost, f))
}

// WithMapError is shorthand for With(Features{MapError: mapper}).
func (f Fn[T, R]) WithMapError(mapper func(error) error) Fn[T, R] {
	return Fn[T, R](mapErr(f, mapper))
}

// WithOnError is shorthand for With(Features{OnError: handler}).
func (f Fn[T, R]) WithOnError(handler func(error)) Fn[T, R] {
	return Fn[T, R](onErr(f, handler))
}
