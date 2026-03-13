package slice

// Scan reduces a slice like Fold, but collects all intermediate accumulator values.
// It includes the initial value as the first element (Haskell scanl semantics),
// so the result has len(ts)+1 elements. Returns [initial] for an empty or nil slice.
// Law: the last element equals Fold(ts, initial, fn).
func Scan[T, R any](ts []T, initial R, fn func(R, T) R) Mapper[R] {
	results := make([]R, len(ts)+1)
	results[0] = initial

	acc := initial
	for i, t := range ts {
		acc = fn(acc, t)
		results[i+1] = acc
	}

	return results
}
