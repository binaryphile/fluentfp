package slice

// Unique removes duplicate elements, preserving first occurrence.
// For non-comparable types or key-based deduplication, use [UniqueBy].
// Note: for float types, NaN != NaN, so NaN values are never deduplicated.
// Returns nil for nil input.
func Unique[T comparable](ts []T) Mapper[T] {
	if ts == nil {
		return nil
	}

	seen := make(map[T]bool, len(ts))
	result := make(Mapper[T], 0, len(ts))

	for _, t := range ts {
		if !seen[t] {
			seen[t] = true
			result = append(result, t)
		}
	}

	return result
}
