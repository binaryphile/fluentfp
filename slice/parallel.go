package slice

import "github.com/binaryphile/fluentfp/internal/base"

// PMap returns the result of applying fn to each member of m, using the specified
// number of worker goroutines. Order is preserved. The fn must be safe for concurrent use.
func PMap[T, R any](m []T, workers int, fn func(T) R) Mapper[R] {
	return base.PMap(m, workers, fn)
}
