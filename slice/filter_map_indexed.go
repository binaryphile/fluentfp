package slice

// FilterMapIndexed applies fn to each element's index and value, keeping only
// results where fn returns true. Combines filtering and type-changing
// transformation in a single pass. Preserves input order among kept elements.
// Returns nil for nil input. fn must not be nil.
//
// It is a standalone function because Go methods cannot introduce new type parameters.
func FilterMapIndexed[T, R any](ts []T, fn func(int, T) (R, bool)) Mapper[R] {
	if ts == nil {
		return nil
	}

	result := make(Mapper[R], 0, len(ts))

	for i, t := range ts {
		if r, ok := fn(i, t); ok {
			result = append(result, r)
		}
	}

	return result
}
