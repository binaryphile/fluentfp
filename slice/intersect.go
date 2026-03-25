package slice

// Intersect returns elements present in both a and b, deduplicated,
// preserving first-occurrence order from a. Returns empty for nil/empty input.
func Intersect[T comparable](a, b []T) Mapper[T] {
	if len(a) == 0 || len(b) == 0 {
		return Mapper[T]{}
	}

	bSet := ToSet(b)
	hint := min(len(a), len(b))
	seen := make(map[T]bool, hint)
	result := make(Mapper[T], 0, hint)

	for _, v := range a {
		if bSet[v] && !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}

	return result
}
