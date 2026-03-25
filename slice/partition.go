package slice

import "github.com/binaryphile/fluentfp/internal/base"

// Partition splits ts into two slices: elements where fn returns true, and elements where it returns false.
// Input order is preserved in both results. Single pass. Both results are independent slices.
// fn must not be nil.
func Partition[T any](ts []T, fn func(T) bool) (Mapper[T], Mapper[T]) {
	match, rest := base.Partition(ts, fn)
	return match, rest
}
