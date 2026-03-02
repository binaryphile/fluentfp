package slice

// Contains returns true if ts contains target.
func Contains[T comparable](ts []T, target T) bool {
	for _, t := range ts {
		if t == target {
			return true
		}
	}
	return false
}
