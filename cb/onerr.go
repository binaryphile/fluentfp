package cb

import "context"

// OnErr wraps fn so that onErr is called with the error after fn returns a
// non-nil error. The returned function calls fn, checks for error, calls
// onErr(err) if present, then returns fn's original results unchanged.
//
// onErr must be safe for concurrent use when the returned function is called
// from multiple goroutines.
//
// Panics if fn is nil or onErr is nil.
func OnErr[T, R any](fn func(context.Context, T) (R, error), onErr func(error)) func(context.Context, T) (R, error) {
	if fn == nil {
		panic("cb.OnErr: fn must not be nil")
	}
	if onErr == nil {
		panic("cb.OnErr: onErr must not be nil")
	}

	return func(ctx context.Context, t T) (R, error) {
		r, err := fn(ctx, t)
		if err != nil {
			onErr(err)
		}

		return r, err
	}
}
