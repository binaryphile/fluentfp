package pair

// Zip returns a slice of each pair of elements from the two input slices.
func Zip[V1, V2 any](v1s []V1, v2s []V2) []For[V1, V2] {
	if len(v1s) != len(v2s) {
		panic("zip: arguments must have same length")
	}

	result := make([]For[V1, V2], len(v1s))
	for i := range v1s {
		result[i] = New(v1s[i], v2s[i])
	}

	return result
}
