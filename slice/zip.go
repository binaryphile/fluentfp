package slice

import (
	"github.com/binaryphile/fluentfp/tuple"
)

// Zip returns a slice of each pair of elements from the two input slices.
func Zip[T, T2 any](ts []T, t2s []T2) []tuple.Pair[T, T2] {
	if len(ts) != len(t2s) {
		panic("zip: arguments must have same length")
	}

	result := make([]tuple.Pair[T, T2], len(ts))
	for i := range ts {
		result[i] = tuple.Pair[T, T2]{
			V1: ts[i],
			V2: t2s[i],
		}
	}

	return result
}
