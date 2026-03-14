package option

import (
	"os"
)

// Env returns an ok option of the environment variable's value if set and non-empty,
// or not-ok if unset or empty.
func Env(key string) String {
	return NonEmpty(os.Getenv(key))
}

// Map returns an option of the result of applying fn to the option's value if ok, or not-ok otherwise.
// For same-type mapping, use the Convert method.
func Map[T, R any](b Option[T], fn func(T) R) (_ Option[R]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// FlatMap returns the result of applying fn to the option's value if ok, or not-ok otherwise.
func FlatMap[T, R any](b Option[T], fn func(T) Option[R]) (_ Option[R]) {
	if !b.ok {
		return
	}

	return fn(b.t)
}

// Lift transforms a function operating on T into one operating on Option[T].
// The lifted function executes only when the option is ok.
func Lift[T any](fn func(T)) func(Option[T]) {
	return func(opt Option[T]) {
		opt.IfOk(fn)
	}
}

// Lookup returns an ok option of the value at key in m, or not-ok if the key is absent.
func Lookup[K comparable, V any](m map[K]V, key K) (_ Option[V]) {
	v, ok := m[key]
	if !ok {
		return
	}

	return Of(v)
}

// OrFalse returns the option's value if ok, or false if not-ok.
// Standalone for type safety — Go methods cannot constrain T to bool.
func OrFalse(o Option[bool]) bool {
	return o.OrZero()
}

// NotOk returns a not-ok option of type T.
func NotOk[T any]() (_ Option[T]) {
	return
}
