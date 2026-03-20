package call

import "context"

// Func is the call shape all decorators operate on:
// a context-aware function that returns a value or an error.
type Func[T, R any] func(context.Context, T) (R, error)

// NewFunc wraps a plain function as a Func. Go infers the type
// parameters, so you don't need to specify them explicitly:
//
//	call.NewFunc(fetchUser).With(...)
func NewFunc[T, R any](fn func(context.Context, T) (R, error)) Func[T, R] {
	return fn
}

// Decorator wraps a Func, returning a Func with the same signature.
type Decorator[T, R any] func(Func[T, R]) Func[T, R]

// With applies decorators to f in order. Each decorator wraps
// the result of the previous one (innermost-first):
//
//	call.Func[string, User](fetchUser).With(A, B, C)
//
// produces C(B(A(fetchUser))). A is innermost, C is outermost.
func (f Func[T, R]) With(ds ...Decorator[T, R]) Func[T, R] {
	for _, d := range ds {
		if d == nil {
			panic("call.With: nil decorator")
		}
		f = Func[T, R](d(f))
	}
	return f
}

// CircuitBreaker returns a Decorator that wraps a Func with
// circuit breaker protection. See [WithBreaker] for details.
func CircuitBreaker[T, R any](b *Breaker) Decorator[T, R] {
	return func(fn Func[T, R]) Func[T, R] {
		return Func[T, R](WithBreaker(b, fn))
	}
}

// Retrier returns a Decorator that retries on error with the
// given backoff strategy. See [Retry] for details.
func Retrier[T, R any](maxAttempts int, backoff Backoff, shouldRetry func(error) bool) Decorator[T, R] {
	return func(fn Func[T, R]) Func[T, R] {
		return Func[T, R](Retry(maxAttempts, backoff, shouldRetry, fn))
	}
}

// Throttler returns a Decorator that bounds concurrent calls.
// See [Throttle] for details.
func Throttler[T, R any](n int) Decorator[T, R] {
	return func(fn Func[T, R]) Func[T, R] {
		return Func[T, R](Throttle(n, fn))
	}
}

// ThrottlerWeighted returns a Decorator that bounds concurrent
// calls by total cost. See [ThrottleWeighted] for details.
func ThrottlerWeighted[T, R any](capacity int, cost func(T) int) Decorator[T, R] {
	return func(fn Func[T, R]) Func[T, R] {
		return Func[T, R](ThrottleWeighted(capacity, cost, fn))
	}
}

// ErrMapper returns a Decorator that transforms errors.
// See [MapErr] for details.
func ErrMapper[T, R any](mapper func(error) error) Decorator[T, R] {
	return func(fn Func[T, R]) Func[T, R] {
		return Func[T, R](MapErr(fn, mapper))
	}
}

// OnError returns a Decorator that calls handler on error.
// See [OnErr] for details.
func OnError[T, R any](handler func(error)) Decorator[T, R] {
	return func(fn Func[T, R]) Func[T, R] {
		return Func[T, R](OnErr(fn, handler))
	}
}
