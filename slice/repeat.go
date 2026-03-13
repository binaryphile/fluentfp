package slice

// RepeatN returns n copies of v.
// Returns empty for n <= 0.
func RepeatN[T any](v T, n int) Mapper[T] {
	if n <= 0 {
		return Mapper[T]{}
	}

	result := make([]T, n)
	for i := range result {
		result[i] = v
	}

	return result
}
