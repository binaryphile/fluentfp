package slice

// Union returns the deduplicated combination of a and b, preserving
// first-occurrence order (all of a first, then extras from b).
func Union[T comparable](a, b Mapper[T]) Mapper[T] {
	if len(a) == 0 && len(b) == 0 {
		return Mapper[T]{}
	}

	seen := make(map[T]bool, len(a)+len(b))
	result := make(Mapper[T], 0, len(a)+len(b))

	for _, v := range a {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}

	for _, v := range b {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}

	return result
}
