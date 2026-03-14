package slice

import "github.com/binaryphile/fluentfp/internal/base"

// PMap returns the result of applying fn to each member of ts, using the specified
// number of worker goroutines. Order is preserved. The fn must be safe for concurrent use.
//
// Panics in fn are recovered, converted to *[rslt.PanicError] with a stack captured
// during recovery, and re-panicked on the calling goroutine after all workers exit.
//
// Panics if workers <= 0.
func PMap[T, R any](ts []T, workers int, fn func(T) R) Mapper[R] {
	return base.PMap(ts, workers, fn)
}
