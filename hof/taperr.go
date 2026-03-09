package hof

import "context"

// TapErr wraps fn so that onErr is called after fn returns a non-nil error.
// The returned function calls fn, checks for error, calls onErr() if present,
// then returns fn's original results unchanged.
//
// onErr must be safe for concurrent use when the returned function is called
// from multiple goroutines (e.g., context.CancelFunc is safe).
//
// Panics if fn is nil or onErr is nil.
func TapErr[T, R any](fn func(context.Context, T) (R, error), onErr func()) func(context.Context, T) (R, error) {
	if fn == nil {
		panic("hof.TapErr: fn must not be nil")
	}
	if onErr == nil {
		panic("hof.TapErr: onErr must not be nil")
	}

	return func(ctx context.Context, t T) (R, error) {
		r, err := fn(ctx, t)
		if err != nil {
			onErr()
		}

		return r, err
	}
}
