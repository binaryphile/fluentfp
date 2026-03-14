package slice

import (
	"cmp"

	"github.com/binaryphile/fluentfp/option"
)

// MinBy returns the element with the smallest key, or not-ok if the slice is empty.
// Keys are compared using [cmp.Compare], which places NaN before -Inf for float types.
// If multiple elements share the smallest key, the first one is returned.
// The key function is called exactly once per element.
func MinBy[T any, K cmp.Ordered](ts []T, key func(T) K) (_ option.Option[T]) {
	if len(ts) == 0 {
		return
	}

	best := ts[0]
	bestKey := key(best)

	for _, t := range ts[1:] {
		k := key(t)
		if cmp.Compare(k, bestKey) < 0 {
			best = t
			bestKey = k
		}
	}

	return option.Of(best)
}

// MaxBy returns the element with the largest key, or not-ok if the slice is empty.
// Keys are compared using [cmp.Compare], which places NaN before -Inf for float types.
// If multiple elements share the largest key, the first one is returned.
// The key function is called exactly once per element.
func MaxBy[T any, K cmp.Ordered](ts []T, key func(T) K) (_ option.Option[T]) {
	if len(ts) == 0 {
		return
	}

	best := ts[0]
	bestKey := key(best)

	for _, t := range ts[1:] {
		k := key(t)
		if cmp.Compare(k, bestKey) > 0 {
			best = t
			bestKey = k
		}
	}

	return option.Of(best)
}
