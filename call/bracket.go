package call

import "context"

// Bracket returns a [Decorator] that wraps a [Func] with acquire-before and
// release-after semantics.
//
// acquire is called with the input context and value. If it returns an error,
// fn does not run and the release closure is not called. If acquire succeeds,
// its returned release closure runs via defer — even if fn returns an error
// or panics.
//
// The release closure captures whatever state acquire needs to clean up
// (semaphore token, connection handle, admission weight):
//
//	semAcquire := func(ctx context.Context, _ int) (func(), error) {
//	    sem <- struct{}{}
//	    return func() { <-sem }, nil
//	}
//	guarded := call.From(process).With(call.Bracket(semAcquire))
//
// Decorator ordering matters:
//   - fn.With(Retrier(...), Bracket(acq)) = acquire/release per retry attempt
//   - fn.With(Bracket(acq), Retrier(...)) = one acquire across all retries
//
// The release closure must not panic. If it does, the panic propagates and
// any fn result or error is lost (standard Go defer semantics).
//
// acquire must not be nil.
func Bracket[T, R any](acquire func(context.Context, T) (func(), error)) Decorator[T, R] {
	if acquire == nil {
		panic("call.Bracket: acquire must not be nil")
	}

	return func(fn Func[T, R]) Func[T, R] {
		if fn == nil {
			panic("call.Bracket: fn must not be nil")
		}

		return func(ctx context.Context, t T) (R, error) {
			var zero R

			release, err := acquire(ctx, t)
			if err != nil {
				return zero, err
			}

			defer release()

			return fn(ctx, t)
		}
	}
}
