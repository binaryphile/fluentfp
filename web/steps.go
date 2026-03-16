package web

import (
	"strconv"

	"github.com/binaryphile/fluentfp/rslt"
)

// Steps chains functions in order, short-circuiting on first error.
// Each step may validate, normalize, or transform the value.
// Zero steps returns identity (rslt.Ok). Panics if any fn is nil.
// For aggregated validation (collecting all errors), write a single step
// that runs checks and returns a structured *Error with Details.
func Steps[T any](fns ...func(T) rslt.Result[T]) func(T) rslt.Result[T] {
	for i, fn := range fns {
		if fn == nil {
			panic("web.Steps: fn at index " + strconv.Itoa(i) + " must not be nil")
		}
	}

	return func(t T) rslt.Result[T] {
		result := rslt.Ok(t)
		for _, fn := range fns {
			result = result.FlatMap(fn)
		}

		return result
	}
}
