package pair

// Zip returns a slice of each pair of elements from the two input slices.
// Panics if the slices have different lengths.
func Zip[A, B any](as []A, bs []B) []X[A, B] {
	if len(as) != len(bs) {
		panic("zip: arguments must have same length")
	}

	result := make([]X[A, B], len(as))
	for i := range as {
		result[i] = Of(as[i], bs[i])
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
