package slice

// GroupBy groups elements by the key returned by fn.
// Returns Entries[K, []T] — a defined map type with methods for chaining.
// Preserves order within each group.
func GroupBy[T any, K comparable](ts []T, fn func(T) K) Entries[K, []T] {
	groups := make(Entries[K, []T])
	for _, t := range ts {
		key := fn(t)
		groups[key] = append(groups[key], t)
	}

	return groups
}
