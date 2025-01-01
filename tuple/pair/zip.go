package pair

// Zip returns a slice of each pair of elements from the two input slices.
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
