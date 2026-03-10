package combo

// PowerSet returns all subsets of items.
// Returns 2^n results — use only for small inputs.
func PowerSet[T any](items []T) [][]T {
	if len(items) == 0 {
		return [][]T{{}}
	}

	rest := PowerSet(items[1:])

	var withFirst [][]T
	for _, subset := range rest {
		withFirst = append(withFirst, append([]T{items[0]}, subset...))
	}

	return append(rest, withFirst...)
}
