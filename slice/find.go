package slice

import "github.com/binaryphile/fluentfp/option"

// Find returns the first element matching the predicate, or not-ok if none match.
func Find[T any](ts []T, fn func(T) bool) option.Basic[T] {
	for _, t := range ts {
		if fn(t) {
			return option.Of(t)
		}
	}
	return option.NotOk[T]()
}
