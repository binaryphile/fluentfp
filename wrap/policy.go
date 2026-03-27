package wrap

import "context"

// Fn is the function shape all decorators operate on:
// a context-aware function that returns a value or an error.
type Fn[T, R any] func(context.Context, T) (R, error)

// Func wraps a plain function as an Fn for fluent decoration.
// Go infers the type parameters from fn:
//
//	wrap.Func(fetchUser).WithRetry(3, wrap.Exponential(time.Second), nil)
func Func[T, R any](fn func(context.Context, T) (R, error)) Fn[T, R] {
	return fn
}

// Decorator wraps an Fn, returning an Fn with the same signature.
// Use with [Fn.With] for advanced composition with first-class decorator values.
type Decorator[T, R any] func(Fn[T, R]) Fn[T, R]

// With applies decorators to f in order. Each decorator wraps
// the result of the previous one (innermost-first):
//
//	wrap.Func(fetchUser).With(A, B, C)
//
// produces C(B(A(fetchUser))). A is innermost, C is outermost.
// For the common case, prefer the chainable With* methods instead.
func (f Fn[T, R]) With(ds ...Decorator[T, R]) Fn[T, R] {
	for _, d := range ds {
		if d == nil {
			panic("wrap.With: nil decorator")
		}
		f = Fn[T, R](d(f))
	}
	return f
}

// WithRetry wraps f to retry on error up to maxAttempts total times.
// See [retry] for behavioral details.
func (f Fn[T, R]) WithRetry(maxAttempts int, backoff Backoff, shouldRetry func(error) bool) Fn[T, R] {
	return Fn[T, R](retry(maxAttempts, backoff, shouldRetry, f))
}

// WithBreaker wraps f with circuit breaker protection.
// See [withBreaker] for behavioral details.
func (f Fn[T, R]) WithBreaker(b *Breaker) Fn[T, R] {
	return Fn[T, R](withBreaker(b, f))
}

// WithThrottle wraps f with count-based concurrency control.
// At most n calls execute concurrently.
func (f Fn[T, R]) WithThrottle(n int) Fn[T, R] {
	return Fn[T, R](throttle(n, f))
}

// WithThrottleWeighted wraps f with cost-based concurrency control.
// The total cost of concurrently-executing calls never exceeds capacity.
func (f Fn[T, R]) WithThrottleWeighted(capacity int, cost func(T) int) Fn[T, R] {
	return Fn[T, R](throttleWeighted(capacity, cost, f))
}

// WithMapError wraps f so that any non-nil error is transformed by mapper.
func (f Fn[T, R]) WithMapError(mapper func(error) error) Fn[T, R] {
	return Fn[T, R](mapErr(f, mapper))
}

// WithOnError wraps f so that handler is called on non-nil errors.
// The error is not modified.
func (f Fn[T, R]) WithOnError(handler func(error)) Fn[T, R] {
	return Fn[T, R](onErr(f, handler))
}
