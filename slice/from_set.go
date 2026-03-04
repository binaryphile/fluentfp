package slice

// FromSet extracts the members of a set (keys where value is true) as a Mapper.
// Order is not guaranteed (map iteration order).
func FromSet[T comparable](m map[T]bool) Mapper[T] {
	result := make([]T, 0, len(m))
	for k, v := range m {
		if v {
			result = append(result, k)
		}
	}

	return result
}
