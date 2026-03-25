package slice

// GroupToMap builds a multi-value map by applying keyFn and valFn to each element.
// Values for each key preserve encounter order.
// Returns an empty writable map for nil or empty input.
// keyFn and valFn must not be nil.
func GroupToMap[T any, K comparable, V any](ts []T, keyFn func(T) K, valFn func(T) V) map[K][]V {
	result := make(map[K][]V)

	for _, t := range ts {
		k := keyFn(t)
		result[k] = append(result[k], valFn(t))
	}

	return result
}
