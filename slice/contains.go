package slice

// Contains reports whether ts contains target. Uses == comparison;
// NaN does not equal itself, so values containing NaN are never found.
func Contains[T comparable](ts []T, target T) bool {
	for _, t := range ts {
		if t == target {
			return true
		}
	}
	return false
}
