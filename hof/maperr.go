package hof

import "context"

// MapErr wraps fn so that any non-nil error returned by fn is transformed by
// mapper before being returned. The result value from fn is always preserved
// unchanged.
//
// Example — annotate errors from a repository call:
//
//	// annotateGetUser wraps err with get-user calling context.
//	annotateGetUser := func(err error) error {
//	    return fmt.Errorf("get user: %w", err)
//	}
//	annotated := hof.MapErr(repo.GetUser, annotateGetUser)
//
// mapper is only called for non-nil errors. For any non-nil input, mapper
// must return a non-nil error; the returned function panics otherwise because
// MapErr cannot safely convert failure into success — the wrapped function
// may not define a meaningful result on error.
//
// Composition order matters: the outer wrapper sees the inner wrapper's
// returned error. Use fmt.Errorf with %w to preserve error identity.
//
// Panics at construction time if fn is nil or mapper is nil.
func MapErr[T, R any](fn func(context.Context, T) (R, error), mapper func(error) error) func(context.Context, T) (R, error) {
	if fn == nil {
		panic("hof.MapErr: fn must not be nil")
	}
	if mapper == nil {
		panic("hof.MapErr: mapper must not be nil")
	}

	return func(ctx context.Context, t T) (R, error) {
		r, err := fn(ctx, t)
		if err != nil {
			mapped := mapper(err)
			if mapped == nil {
				panic("hof.MapErr: mapper must not return nil")
			}

			return r, mapped
		}

		return r, nil
	}
}
