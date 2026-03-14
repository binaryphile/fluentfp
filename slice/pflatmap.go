package slice

import "github.com/binaryphile/fluentfp/internal/base"

// PFlatMap applies fn to each element of ts concurrently using the specified number of
// worker goroutines, then flattens the results into a single slice. Order is preserved.
// The fn must be safe for concurrent use.
//
// It is a standalone function because Go methods cannot introduce new type parameters —
// the target type R must be inferred from the function argument rather than bound on the receiver.
//
// Panics in fn are recovered, converted to *[rslt.PanicError] with a stack captured
// during recovery, and re-panicked on the calling goroutine after all workers exit.
//
// Panics if workers <= 0.
func PFlatMap[T, R any](ts []T, workers int, fn func(T) []R) Mapper[R] {
	return base.PFlatMap(ts, workers, fn)
}
