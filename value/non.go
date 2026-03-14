package value

import "github.com/binaryphile/fluentfp/option"

// NonZero returns an ok option if t is not the zero value, or not-ok otherwise.
func NonZero[T comparable](t T) option.Option[T] {
	return option.NonZero(t)
}

// NonEmpty returns an ok option if s is not empty, or not-ok otherwise.
var NonEmpty = option.NonEmpty

// NonNil returns an ok option containing the dereferenced value if t is not nil, or not-ok otherwise.
func NonNil[T any](t *T) option.Option[T] {
	return option.NonNil(t)
}

// NonZeroCall returns an ok option of fn(t) if t is not the zero value, or not-ok otherwise.
func NonZeroCall[T comparable, R any](t T, fn func(T) R) option.Option[R] {
	return option.NonZeroCall(t, fn)
}

// NonEmptyCall returns an ok option of fn(s) if s is not empty, or not-ok otherwise.
func NonEmptyCall[R any](s string, fn func(string) R) option.Option[R] {
	return option.NonEmptyCall(s, fn)
}

// NonNilCall returns an ok option of fn(dereferenced t) if t is not nil, or not-ok otherwise.
func NonNilCall[T any, R any](t *T, fn func(T) R) option.Option[R] {
	return option.NonNilCall(t, fn)
}
