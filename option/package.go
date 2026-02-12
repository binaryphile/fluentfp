package option

import (
	"os"
)

func Getenv(key string) String {
	result := os.Getenv(key)

	return IfNotZero(result)
}

func Map[T, R any](b Basic[T], fn func(T) R) (_ Basic[R]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// Lift transforms a function operating on T into one operating on Basic[T].
// The lifted function executes only when the option is ok.
func Lift[T any](fn func(T)) func(Basic[T]) {
	return func(opt Basic[T]) {
		opt.Call(fn)
	}
}

func NotOk[T any]() (_ Basic[T]) {
	return
}
