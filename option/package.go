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

// When returns an ok option of t if cond is true, or not-ok otherwise.
// It is the condition-first counterpart of [New] (which mirrors the comma-ok idiom).
//
// Note: t is evaluated eagerly by Go's call semantics. For expensive
// computations that should only run when cond is true, use [WhenFunc].
//
// Style guidance: prefer When for explicit boolean conditions.
// Prefer [New] when forwarding a comma-ok result (v, ok := m[k]).
func When[T any](cond bool, t T) Option[T] {
	return New(t, cond)
}

// WhenFunc returns an ok option of fn() if cond is true, or not-ok otherwise.
// The function is only called when the condition is true.
// Panics if fn is nil, even when cond is false.
func WhenFunc[T any](cond bool, fn func() T) Option[T] {
	if fn == nil {
		panic("option: WhenFunc called with nil function")
	}

	if !cond {
		return NotOk[T]()
	}

	return Of(fn())
}
