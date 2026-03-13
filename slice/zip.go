package slice

import "github.com/binaryphile/fluentfp/tuple/pair"

// Zip combines corresponding elements from two slices into pairs.
// Truncates to the length of the shorter slice.
// Note: pair.Zip panics on length mismatch; slice.Zip truncates instead,
// which is safer as a default in collection pipelines.
func Zip[A, B any](as []A, bs []B) Mapper[pair.Pair[A, B]] {
	n := min(len(as), len(bs))
	results := make([]pair.Pair[A, B], n)
	for i := range n {
		results[i] = pair.Of(as[i], bs[i])
	}

	return results
}

// ZipWith combines corresponding elements from two slices using fn.
// Truncates to the length of the shorter slice (same as Zip).
func ZipWith[A, B, R any](as []A, bs []B, fn func(A, B) R) Mapper[R] {
	n := min(len(as), len(bs))
	results := make([]R, n)
	for i := range n {
		results[i] = fn(as[i], bs[i])
	}

	return results
}
