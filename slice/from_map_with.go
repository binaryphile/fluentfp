package slice

// FromMapWith transforms each key-value pair in m using fn and returns
// the results as a Mapper for further transformation.
// Order is not guaranteed (map iteration order).
func FromMapWith[K comparable, V, T any](m map[K]V, fn func(K, V) T) Mapper[T] {
	result := make([]T, 0, len(m))
	for k, v := range m {
		result = append(result, fn(k, v))
	}

	return result
}
