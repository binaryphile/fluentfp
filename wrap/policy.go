package wrap

import "context"

// Fn is the function shape all decorators operate on:
// a context-aware function that returns a value or an error.
type Fn[T, R any] func(context.Context, T) (R, error)

// Func wraps a plain function as an Fn for fluent decoration.
// Go infers the type parameters from fn:
//
//	wrap.Func(fetchUser).
//	    Retry(3, wrap.ExpBackoff(time.Second), nil).
//	    Breaker(breaker)
func Func[T, R any](fn func(context.Context, T) (R, error)) Fn[T, R] {
	return fn
}

// Decorator wraps an Fn, returning an Fn with the same signature.
// Use with [Fn.Apply] for custom decorators.
type Decorator[T, R any] func(Fn[T, R]) Fn[T, R]

// Breaker wraps f with circuit breaker protection.
// The breaker is shared state — pass the same *Breaker to multiple
// wrapped functions to have them trip together.
func (f Fn[T, R]) Breaker(b *Breaker) Fn[T, R] {
	return Fn[T, R](withBreaker(b, f))
}

// MapError wraps f so that any non-nil error is transformed by mapper.
func (f Fn[T, R]) MapError(mapper func(error) error) Fn[T, R] {
	return Fn[T, R](mapErr(f, mapper))
}

// OnError wraps f so that handler is called on non-nil errors.
// The error is not modified.
func (f Fn[T, R]) OnError(handler func(error)) Fn[T, R] {
	return Fn[T, R](onErr(f, handler))
}

// Retry wraps f to retry on error up to max total attempts.
func (f Fn[T, R]) Retry(max int, backoff Backoff, shouldRetry func(error) bool) Fn[T, R] {
	return Fn[T, R](retry(max, backoff, shouldRetry, f))
}

// Throttle wraps f with count-based concurrency control.
// At most n calls execute concurrently.
func (f Fn[T, R]) Throttle(n int) Fn[T, R] {
	return Fn[T, R](throttle(n, f))
}

// Weighted wraps f with cost-based concurrency control.
// The total cost of concurrently-executing calls never exceeds capacity.
func (f Fn[T, R]) Weighted(capacity int, cost func(T) int) Fn[T, R] {
	return Fn[T, R](throttleWeighted(capacity, cost, f))
}

// Apply applies custom decorators to f in order (innermost-first).
func (f Fn[T, R]) Apply(ds ...Decorator[T, R]) Fn[T, R] {
	for _, d := range ds {
		if d == nil {
			panic("wrap.Apply: nil decorator")
		}
		f = Fn[T, R](d(f))
	}
	return f
}
