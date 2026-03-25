package slice

// ContainsAll reports whether all elements of targets appear in ts.
// Returns true if targets is nil or empty (vacuous truth).
// Duplicate values in targets do not require duplicate occurrences in ts.
// Equality uses Go == semantics.
func ContainsAll[T comparable](ts, targets []T) bool {
	if len(targets) == 0 {
		return true
	}

	set := ToSet(ts)

	for _, t := range targets {
		if !set[t] {
			return false
		}
	}

	return true
}
