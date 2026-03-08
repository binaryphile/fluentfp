package slice

// GroupBy groups elements by the key returned by fn.
// Returns Mapper[Group[K, T]] — groups preserve first-seen key order and chain directly.
func GroupBy[T any, K comparable](ts Mapper[T], fn func(T) K) Mapper[Group[K, T]] {
	index := make(map[K]int)
	var result []Group[K, T]

	for _, t := range ts {
		key := fn(t)
		if idx, exists := index[key]; exists {
			result[idx].Items = append(result[idx].Items, t)
		} else {
			index[key] = len(result)
			result = append(result, Group[K, T]{Key: key, Items: []T{t}})
		}
	}

	return result
}
