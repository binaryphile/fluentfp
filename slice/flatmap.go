package slice

// FlatMap applies fn to each element of ts and flattens the results into a single slice.
// Nil slices returned by fn are treated as empty. fn must not be nil.
//
// It is a standalone function because Go methods cannot introduce new type parameters —
// the target type R must be inferred from the function argument rather than bound on the receiver.
func FlatMap[T, R any](ts []T, fn func(T) []R) Mapper[R] {
	results := make([]R, 0, len(ts))
	for _, t := range ts {
		results = append(results, fn(t)...)
	}
	return results
}
