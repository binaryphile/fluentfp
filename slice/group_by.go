package slice

// GroupBy groups elements by the key returned by fn.
// Returns a map from key to the elements sharing that key, preserving order within each group.
func GroupBy[T any, K comparable](ts []T, fn func(T) K) map[K][]T {
	groups := make(map[K][]T)
	for _, t := range ts {
		key := fn(t)
		groups[key] = append(groups[key], t)
	}

	return groups
}
