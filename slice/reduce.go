package slice

import "github.com/binaryphile/fluentfp/option"

// Reduce combines elements left-to-right using the first element as the initial value.
// Returns not-ok if the slice is empty. For a single-element slice, returns that element
// without calling fn. Unlike [Fold], Reduce requires the result type to match the element type.
// Nil fn does not panic when the slice has fewer than two elements (fn is never invoked).
func Reduce[T any](ts []T, fn func(T, T) T) (_ option.Option[T]) {
	if len(ts) == 0 {
		return
	}

	acc := ts[0]

	for _, t := range ts[1:] {
		acc = fn(acc, t)
	}

	return option.Of(acc)
}
