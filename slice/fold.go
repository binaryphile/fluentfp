package slice

// Fold reduces a slice to a single value by applying fn to each element.
// It starts with initial and applies fn(accumulator, element) for each element from left to right.
// Returns initial if the slice is empty.
func Fold[T, R any](ts []T, initial R, fn func(R, T) R) R {
	acc := initial
	for _, t := range ts {
		acc = fn(acc, t)
	}

	return acc
}

// TryFold folds ts from left to right, starting with initial.
// For each element t, it calls fn(acc, t) and uses the returned
// accumulator as the input to the next step.
//
// If fn returns an error, TryFold stops immediately and returns the
// accumulator value returned by that failing call together with the
// error. Remaining elements are not visited.
//
// For an empty or nil slice, TryFold returns initial and nil.
//
// Note: if R is mutable or contains references, the returned
// accumulator may reflect partial mutation performed by fn before
// it returned an error. TryFold does not attempt rollback.
func TryFold[T, R any](ts []T, initial R, fn func(R, T) (R, error)) (R, error) {
	acc := initial
	for _, t := range ts {
		var err error
		acc, err = fn(acc, t)
		if err != nil {
			return acc, err
		}
	}

	return acc, nil
}
