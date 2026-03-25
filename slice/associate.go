package slice

// Associate builds a map by applying fn to each element to produce a key-value pair.
// If multiple elements produce the same key, the last one wins.
// Returns an empty writable map for nil or empty input.
// fn must not be nil.
func Associate[T any, K comparable, V any](ts []T, fn func(T) (K, V)) map[K]V {
	result := make(map[K]V, len(ts))

	for _, t := range ts {
		k, v := fn(t)
		result[k] = v
	}

	return result
}
