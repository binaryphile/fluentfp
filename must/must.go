package must

import (
	"fmt"
	"os"
)

// AssertNil panics if err is not nil.
func AssertNil(err error) {
	if err != nil {
		panic(err)
	}
}

// Must returns the value of a (value, error) pair of arguments unless error is non-nil.
// In that case, it panics.
func Must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}

	return t
}

// MustGetenv returns the value in the environment variable named by key.
// It panics if the environment variable doesn't exist or is empty.
func MustGetenv(key string) string {
	result := os.Getenv(key)

	if result == "" {
		panic(fmt.Sprintf("expected value for environment variable %s", key))
	}

	return result
}
