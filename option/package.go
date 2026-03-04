package option

import (
	"os"
)

func Getenv(key string) String {
	result := os.Getenv(key)

	return NonZero(result)
}

func Map[T, R any](b Option[T], fn func(T) R) (_ Option[R]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
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

func NotOk[T any]() (_ Option[T]) {
	return
}
