package option

import (
	"os"
)

func Getenv(key string) String {
	result := os.Getenv(key)

	return IfProvided(result)
}

func Map[T, R any](b Basic[T], fn func(T) R) (_ Basic[R]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

func NotOk[T any]() (_ Basic[T]) {
	return
}
