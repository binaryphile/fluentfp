package slice

// AssociateIndexed builds a map by applying fn to each element's index and value
// to produce a key-value pair. If multiple elements produce the same key,
// the last one wins. Returns an empty writable map for nil or empty input.
// fn must not be nil.
func AssociateIndexed[T any, K comparable, V any](ts []T, fn func(int, T) (K, V)) map[K]V {
	result := make(map[K]V, len(ts))

	for i, t := range ts {
		k, v := fn(i, t)
		result[k] = v
	}

	return result
}
