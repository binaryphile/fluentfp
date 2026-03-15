package combo

import "github.com/binaryphile/fluentfp/slice"

// Combinations returns all k-element subsets of items, preserving order.
// Returns C(n,k) results.
func Combinations[T any](items []T, k int) slice.Mapper[[]T] {
	if k < 0 || k > len(items) {
		return nil
	}

	if k == 0 {
		return [][]T{{}}
	}

	// Include items[0] in the subset
	var result [][]T
	for _, c := range Combinations(items[1:], k-1) {
		result = append(result, append([]T{items[0]}, c...))
	}

	// Exclude items[0] from the subset
	result = append(result, Combinations(items[1:], k)...)

	return result
}
