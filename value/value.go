// Package value provides value-first conditional selection.
// Use this for selecting a value with a fallback, not for executing branches of logic.
package value

import "github.com/binaryphile/fluentfp/option"

// Cond holds a value pending a condition check.
type Cond[T any] struct {
	v T
}

// Of wraps a value for conditional selection.
func Of[T any](t T) Cond[T] {
	return Cond[T]{v: t}
}

// When returns an option: Ok(value) if condition true, NotOk if false.
func (c Cond[T]) When(ok bool) option.Basic[T] {
	return option.New(c.v, ok)
}

// Coalesce returns the first non-zero value, or zero if all are zero.
func Coalesce[T comparable](vals ...T) (_ T) {
	var zero T
	for _, v := range vals {
		if v != zero {
			return v
		}
	}

	return
}

// LazyCond holds a function for deferred value computation.
type LazyCond[T any] struct {
	fn func() T
}

// OfCall wraps a function for lazy conditional selection.
// The function is only called if the condition is true.
func OfCall[T any](fn func() T) LazyCond[T] {
	return LazyCond[T]{fn: fn}
}

// When evaluates fn only if condition true, returns option.
func (c LazyCond[T]) When(ok bool) option.Basic[T] {
	if ok {
		return option.Of(c.fn())
	}

	return option.NotOk[T]()
}
