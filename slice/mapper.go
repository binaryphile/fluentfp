package slice

import (
	"github.com/binaryphile/fluentfp/internal/base"
	"github.com/binaryphile/fluentfp/option"
)

// From creates a Mapper for fluent operations on a slice.
func From[T any](ts []T) Mapper[T] {
	return ts
}

// FindAs returns the first element that type-asserts to R, or not-ok if none match.
// Useful for finding a specific concrete type in a slice of interfaces.
func FindAs[R, T any](ts Mapper[T]) option.Option[R] {
	return base.FindAs[R, T](ts)
}
