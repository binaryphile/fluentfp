package must

import (
	"fmt"
	"os"
)

// BeNil panics if err is not nil.
func BeNil(err error) {
	if err != nil {
		panic(err)
	}
}

// Get returns the value of a (value, error) pair of arguments unless error is non-nil.
// In that case, it panics.
func Get[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}

	return t
}

// Get2 returns the value of a (value, value, error) set of arguments unless error is non-nil.
// In that case, it panics.
func Get2[T, T2 any](t T, t2 T2, err error) (T, T2) {
	if err != nil {
		panic(err)
	}

	return t, t2
}

// Getenv returns the value in the environment variable named by key.
// It panics if the environment variable doesn't exist or is empty.
func Getenv(key string) string {
	result := os.Getenv(key)

	if result == "" {
		panic(fmt.Sprintf("expected value for environment variable %s", key))
	}

	return result
}

// Of returns the "must" version of fn.
// fn must be a single-argument function.
func Of[T, R any](fn func(T) (R, error)) func(T) R {
	return func(t T) R {
		result, err := fn(t)
		BeNil(err)

		return result
	}
}
