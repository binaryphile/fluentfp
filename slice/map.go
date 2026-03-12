package slice

// Map returns the result of applying fn to each member of ts.
// It is a standalone function because Go methods cannot introduce new type parameters —
// the target type R must be inferred from the function argument rather than bound on the receiver.
func Map[T, R any](ts []T, fn func(T) R) Mapper[R] {
	results := make([]R, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}
