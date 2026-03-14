// Package combo provides combinatorial constructions: Cartesian products,
// permutations, combinations, and power sets.
package combo

import "github.com/binaryphile/fluentfp/tuple/pair"

// CartesianProduct returns all pairs from a and b.
func CartesianProduct[A, B any](a []A, b []B) []pair.Pair[A, B] {
	if len(a) == 0 || len(b) == 0 {
		return nil
	}

	result := make([]pair.Pair[A, B], 0, len(a)*len(b))
	for _, x := range a {
		for _, y := range b {
			result = append(result, pair.Of(x, y))
		}
	}

	return result
}

// CartesianProductWith applies fn to every (a, b) pair.
// Avoids intermediate Pair allocation when the caller transforms immediately.
// Panics if fn is nil.
func CartesianProductWith[A, B, R any](a []A, b []B, fn func(A, B) R) []R {
	if fn == nil {
		panic("combo.CartesianProductWith: fn must not be nil")
	}

	if len(a) == 0 || len(b) == 0 {
		return nil
	}

	result := make([]R, 0, len(a)*len(b))
	for _, x := range a {
		for _, y := range b {
			result = append(result, fn(x, y))
		}
	}

	return result
}
