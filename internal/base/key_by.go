package base

// KeyBy indexes elements by a key derived from fn, returning a map from key to element.
// If multiple elements produce the same key, the last one wins.
func KeyBy[T any, K comparable](ts []T, fn func(T) K) map[K]T {
	result := make(map[K]T, len(ts))
	for _, t := range ts {
		result[fn(t)] = t
	}
	return result
}
