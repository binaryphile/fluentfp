package combo

// Permutations returns all orderings of items.
// Returns n! results — use only for small inputs.
func Permutations[T any](items []T) [][]T {
	if len(items) == 0 {
		return [][]T{{}}
	}

	var result [][]T

	for i := range items {
		rest := make([]T, 0, len(items)-1)
		rest = append(rest, items[:i]...)
		rest = append(rest, items[i+1:]...)

		for _, perm := range Permutations(rest) {
			result = append(result, append([]T{items[i]}, perm...))
		}
	}

	return result
}
