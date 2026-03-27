package wrap

import (
	"context"

	"github.com/binaryphile/fluentfp/option"
)

// Fn is the function shape all decorators operate on:
// a context-aware function that returns a value or an error.
type Fn[T, R any] func(context.Context, T) (R, error)

// Func wraps a plain function as an Fn for fluent decoration.
// Go infers the type parameters from fn:
//
//	wrap.Func(fetchUser).With(wrap.Modes{
//	    RetryOpt: &wrap.RetryConfig{Max: 3, Backoff: wrap.ExpBackoff(time.Second)},
//	})
func Func[T, R any](fn func(context.Context, T) (R, error)) Fn[T, R] {
	return fn
}

// Decorator wraps an Fn, returning an Fn with the same signature.
// For custom decorators not covered by [Modes], apply manually:
//
//	decorated := wrap.Fn[T, R](myDecorator(fn))
type Decorator[T, R any] func(Fn[T, R]) Fn[T, R]

// RetryConfig configures the retry mode.
type RetryConfig struct {
	Backoff     Backoff
	Max         int
	ShouldRetry func(error) bool
}

// Modes configures which decorators to apply. Nil fields are skipped.
// The Opt suffix signals that nil is the expected way to omit a mode.
//
// Decorators are applied in a fixed order (innermost to outermost):
// OnError → MapError → Retry → Breaker → Throttle.
type Modes struct {
	BreakerOpt  *Breaker
	MapErrorOpt func(error) error
	OnErrorOpt  func(error)
	RetryOpt    *RetryConfig
	ThrottleOpt *int
}

// With applies the modes to f in a fixed order. Nil fields are skipped.
// Innermost to outermost: OnError → MapError → Retry → Breaker → Throttle.
func (f Fn[T, R]) With(m Modes) Fn[T, R] {
	if handler, ok := option.New(m.OnErrorOpt, m.OnErrorOpt != nil).Get(); ok {
		f = Fn[T, R](onErr(f, handler))
	}

	if mapper, ok := option.New(m.MapErrorOpt, m.MapErrorOpt != nil).Get(); ok {
		f = Fn[T, R](mapErr(f, mapper))
	}

	if cfg, ok := option.NonNil(m.RetryOpt).Get(); ok {
		f = Fn[T, R](retry(cfg.Max, cfg.Backoff, cfg.ShouldRetry, f))
	}

	if m.BreakerOpt != nil {
		f = Fn[T, R](withBreaker(m.BreakerOpt, f))
	}

	if n, ok := option.NonNil(m.ThrottleOpt).Get(); ok {
		f = Fn[T, R](throttle(n, f))
	}

	return f
}

// WithRetry is shorthand for With(Modes{RetryOpt: &RetryConfig{...}}).
func (f Fn[T, R]) WithRetry(maxAttempts int, backoff Backoff, shouldRetry func(error) bool) Fn[T, R] {
	return Fn[T, R](retry(maxAttempts, backoff, shouldRetry, f))
}

// WithBreaker is shorthand for With(Modes{BreakerOpt: b}).
func (f Fn[T, R]) WithBreaker(b *Breaker) Fn[T, R] {
	return Fn[T, R](withBreaker(b, f))
}

// WithThrottle is shorthand for With(Modes{ThrottleOpt: &n}).
func (f Fn[T, R]) WithThrottle(n int) Fn[T, R] {
	return Fn[T, R](throttle(n, f))
}

// WithThrottleWeighted wraps f with cost-based concurrency control.
// Not available via Modes because the cost function requires type T.
func (f Fn[T, R]) WithThrottleWeighted(capacity int, cost func(T) int) Fn[T, R] {
	return Fn[T, R](throttleWeighted(capacity, cost, f))
}

// WithMapError is shorthand for With(Modes{MapErrorOpt: mapper}).
func (f Fn[T, R]) WithMapError(mapper func(error) error) Fn[T, R] {
	return Fn[T, R](mapErr(f, mapper))
}

// WithOnError is shorthand for With(Modes{OnErrorOpt: handler}).
func (f Fn[T, R]) WithOnError(handler func(error)) Fn[T, R] {
	return Fn[T, R](onErr(f, handler))
}
