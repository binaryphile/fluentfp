package slice

import "github.com/binaryphile/fluentfp/option"

// IndexOf returns the index of the first occurrence of target, or not-ok if absent.
// Uses == comparison; for predicate-based search, use IndexWhere.
func IndexOf[T comparable](ts []T, target T) option.Option[int] {
	for i, t := range ts {
		if t == target {
			return option.Of(i)
		}
	}
	return option.NotOk[int]()
}

// LastIndexOf returns the index of the last occurrence of target, or not-ok if absent.
// Uses == comparison; for predicate-based search, use LastIndexWhere.
func LastIndexOf[T comparable](ts []T, target T) option.Option[int] {
	for i := len(ts) - 1; i >= 0; i-- {
		if ts[i] == target {
			return option.Of(i)
		}
	}
	return option.NotOk[int]()
}
