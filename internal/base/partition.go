package base

// Partition splits ts into two slices: elements where fn returns true, and elements where it returns false.
// Single pass. Both results are independent slices.
func Partition[T any](ts []T, fn func(T) bool) ([]T, []T) {
	match := make([]T, 0, len(ts))
	rest := make([]T, 0, len(ts))

	for _, t := range ts {
		if fn(t) {
			match = append(match, t)
		} else {
			rest = append(rest, t)
		}
	}

	return match, rest
}
