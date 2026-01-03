package pair

// Zip returns a slice of each pair of elements from the two input slices.
// Panics if the slices have different lengths.
func Zip[V1, V2 any](v1s []V1, v2s []V2) []X[V1, V2] {
	if len(v1s) != len(v2s) {
		panic("zip: arguments must have same length")
	}

	result := make([]X[V1, V2], len(v1s))
	for i := range v1s {
		result[i] = Of(v1s[i], v2s[i])
	}

	return result
}

// ZipWith applies fn to corresponding elements of the two input slices.
// Panics if the slices have different lengths.
func ZipWith[A, B, R any](as []A, bs []B, fn func(A, B) R) []R {
	if len(as) != len(bs) {
		panic("zipWith: arguments must have same length")
	}

	result := make([]R, len(as))
	for i := range as {
		result[i] = fn(as[i], bs[i])
	}

	return result
}
