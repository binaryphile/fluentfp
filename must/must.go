package must

import (
	"errors"
	"fmt"
	"os"
)

// ErrEnvUnset indicates the environment variable is not set.
var ErrEnvUnset = errors.New("environment variable unset")

// ErrEnvEmpty indicates the environment variable is set but empty.
var ErrEnvEmpty = errors.New("environment variable empty")

// ErrNilFunction indicates a nil function was passed where one is required.
var ErrNilFunction = errors.New("nil function")

// BeNil panics if err is not nil. The panic value is err itself,
// preserving error chains for errors.Is/errors.As after recovery.
func BeNil(err error) {
	if err != nil {
		panic(err)
	}
}

// Get returns the value of a (value, error) pair, or panics if error is non-nil.
// The panic value is the original error, preserving error chains for
// errors.Is/errors.As after recovery.
func Get[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}

	return t
}

// Get2 returns the values of a (value, value, error) triple, or panics if error is non-nil.
// The panic value is the original error, preserving error chains for
// errors.Is/errors.As after recovery.
func Get2[T, T2 any](t T, t2 T2, err error) (T, T2) {
	if err != nil {
		panic(err)
	}

	return t, t2
}

// NonEmptyEnv returns the value of the environment variable named by key.
// It panics if the variable is unset or empty.
// The panic value wraps [ErrEnvUnset] or [ErrEnvEmpty] for errors.Is matching.
func NonEmptyEnv(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		panic(fmt.Errorf("must.NonEmptyEnv(%q): %w", key, ErrEnvUnset))
	}
	if v == "" {
		panic(fmt.Errorf("must.NonEmptyEnv(%q): %w", key, ErrEnvEmpty))
	}

	return v
}

// From returns the "must" version of fn.
// fn must be a single-argument function.
// Panics immediately if fn is nil, wrapping [ErrNilFunction].
func From[T, R any](fn func(T) (R, error)) func(T) R {
	if fn == nil {
		panic(fmt.Errorf("must.From: %w", ErrNilFunction))
	}

	return func(t T) R {
		result, err := fn(t)
		BeNil(err)

		return result
	}
}
