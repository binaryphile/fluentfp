package slice

// FilterMap applies fn to each element and keeps only the results where fn returns true.
// It combines filtering and type-changing transformation in a single pass.
// The inclusion decision and transformed output are derived from a single callback
// invocation per element. Preserves input order among kept elements.
// Returns nil for nil input.
func FilterMap[T, R any](ts []T, fn func(T) (R, bool)) Mapper[R] {
	if ts == nil {
		return nil
	}

	result := make(Mapper[R], 0, len(ts))

	for _, t := range ts {
		if r, ok := fn(t); ok {
			result = append(result, r)
		}
	}

	return result
}
