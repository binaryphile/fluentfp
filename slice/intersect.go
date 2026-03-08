package slice

// Intersect returns elements present in both a and b, deduplicated,
// preserving first-occurrence order from a.
func Intersect[T comparable](a, b Mapper[T]) Mapper[T] {
	if len(a) == 0 || len(b) == 0 {
		return Mapper[T]{}
	}

	bSet := ToSet(b)
	cap := min(len(a), len(b))
	seen := make(map[T]bool, cap)
	result := make(Mapper[T], 0, cap)

	for _, v := range a {
		if bSet[v] && !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}

	return result
}
