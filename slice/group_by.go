package slice

// GroupBy groups elements by the key returned by fn.
// Groups preserve first-seen key order; items within each group preserve input order.
// fn must not be nil.
func GroupBy[T any, K comparable](ts []T, fn func(T) K) Mapper[Group[K, T]] {
	if len(ts) == 0 {
		return Mapper[Group[K, T]]{}
	}

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
