package slice

import "github.com/binaryphile/fluentfp/internal/base"

// ParallelMap returns the result of applying fn to each member of m, using the specified
// number of worker goroutines. Order is preserved. The fn must be safe for concurrent use.
func ParallelMap[T, R any](m Mapper[T], workers int, fn func(T) R) Mapper[R] {
	return base.ParallelMap(m, workers, fn)
}
