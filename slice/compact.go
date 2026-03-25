package slice

// Compact removes consecutive duplicate elements, like Unix uniq.
// For global deduplication, use [Unique].
// Returns a new slice; does not compact in-place.
// Returns nil for nil input. Equality uses Go == semantics.
func Compact[T comparable](ts []T) Mapper[T] {
	if ts == nil {
		return nil
	}

	if len(ts) == 0 {
		return Mapper[T]{}
	}

	result := make([]T, 0, len(ts))
	result = append(result, ts[0])

	for i := 1; i < len(ts); i++ {
		if ts[i] != ts[i-1] {
			result = append(result, ts[i])
		}
	}

	return result
}
