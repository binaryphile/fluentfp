package slice

// MapIndexed returns the result of applying fn to each element's index and value.
// fn must not be nil.
//
// It is a standalone function because Go methods cannot introduce new type parameters.
func MapIndexed[T, R any](ts []T, fn func(int, T) R) Mapper[R] {
	results := make([]R, len(ts))

	for i, t := range ts {
		results[i] = fn(i, t)
	}

	return results
}
