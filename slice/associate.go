package slice

// Associate builds a map by applying fn to each element to produce a key-value pair.
// If multiple elements produce the same key, the last one wins.
// Returns nil for nil or empty input.
func Associate[T any, K comparable, V any](ts []T, fn func(T) (K, V)) map[K]V {
	if len(ts) == 0 {
		return nil
	}

	result := make(map[K]V, len(ts))

	for _, t := range ts {
		k, v := fn(t)
		result[k] = v
	}

	return result
}
