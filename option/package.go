package option

import (
	"os"
)

func Getenv(key string) String {
	result := os.Getenv(key)

	return IfProvided(result)
}

func Map[T, U any](b Basic[T], TToU func(T) U) (_ Basic[U]) {
	if !b.ok {
		return
	}

	return Of(TToU(b.t))
}
