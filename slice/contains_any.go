package slice

// ContainsAny reports whether any element of targets appears in ts.
// Returns false if targets is nil or empty.
// Duplicates in targets do not affect the result.
// Equality uses Go == semantics.
func ContainsAny[T comparable](ts, targets []T) bool {
	if len(targets) == 0 {
		return false
	}

	set := ToSet(targets)

	for _, t := range ts {
		if set[t] {
			return true
		}
	}

	return false
}
