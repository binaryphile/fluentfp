package slice

// FromMap extracts the values of m as a Mapper for further transformation.
// Order is not guaranteed (map iteration order).
func FromMap[K comparable, V any](m map[K]V) Mapper[V] {
	result := make([]V, 0, len(m))
	for _, v := range m {
		result = append(result, v)
	}

	return result
}
