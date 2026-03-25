package slice

// UniqueBy returns a new slice with duplicate elements removed, where duplicates
// are determined by the key returned by fn. Preserves first occurrence and maintains order.
// fn must not be nil.
func UniqueBy[T any, K comparable](ts []T, fn func(T) K) Mapper[T] {
	seen := make(map[K]bool, len(ts))
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		key := fn(t)
		if !seen[key] {
			seen[key] = true
			results = append(results, t)
		}
	}

	return results
}
