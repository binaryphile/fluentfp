package base

// ToSet returns a map with each element as a key set to true.
// Membership is by key presence (use comma-ok or range), not by value.
// Requires comparable elements.
func ToSet[T comparable](ts []T) map[T]bool {
	set := make(map[T]bool, len(ts))
	for _, t := range ts {
		set[t] = true
	}
	return set
}
