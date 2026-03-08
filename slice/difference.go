package slice

// Difference returns elements in a that are not in b, deduplicated,
// preserving first-occurrence order from a.
func Difference[T comparable](a, b Mapper[T]) Mapper[T] {
	if len(a) == 0 {
		return Mapper[T]{}
	}

	bSet := ToSet(b)
	seen := make(map[T]bool, len(a))
	result := make(Mapper[T], 0, len(a))

	for _, v := range a {
		if !bSet[v] && !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}

	return result
}
