package slice

import "github.com/binaryphile/fluentfp/internal/base"

// Partition splits ts into two slices: elements where fn returns true, and elements where it returns false.
// Single pass. Both results are independent slices.
// For method form, use Mapper[T].Partition or MapperTo[R, T].Partition.
func Partition[T any](ts Mapper[T], fn func(T) bool) (Mapper[T], Mapper[T]) {
	match, rest := base.Partition(ts, fn)
	return match, rest
}
