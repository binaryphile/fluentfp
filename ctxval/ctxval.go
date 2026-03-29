package ctxval

import (
	"context"

	"github.com/binaryphile/fluentfp/option"
)

type typeKey[T any] struct{}
type box[T any] struct{ v T }

// With returns a child context carrying val, keyed by its type T.
// A subsequent call with the same T shadows the parent's value.
//
// For distinct semantic keys of the same type, use [NewKey] instead.
//
// Panics if ctx is nil (same as context.WithValue).
func With[T any](ctx context.Context, val T) context.Context {
	return context.WithValue(ctx, typeKey[T]{}, box[T]{v: val})
}

// Lookup retrieves the value of type T from ctx, returning a not-ok option
// if no value of that type is present. T must match the exact static type
// used in [With]; interface and concrete types are distinct keys.
//
// Panics if ctx is nil (same as context.Context.Value).
func Lookup[T any](ctx context.Context) option.Option[T] {
	b, ok := ctx.Value(typeKey[T]{}).(box[T])

	return option.New(b.v, ok)
}

// Key is a named context key for values of type T. Unlike [With]/[Lookup],
// which key by type alone, each Key is a unique identity — multiple keys
// can carry the same type T without collision.
//
// Typically create with [NewKey]. Declare as package-level variables.
type Key[T any] struct{ _ byte }

// NewKey returns a new unique key for values of type T. Each call returns
// a distinct key (pointer identity).
func NewKey[T any]() *Key[T] { return &Key[T]{} }

// With returns a child context carrying val under this key.
//
// Panics if k is nil (use [NewKey] to create keys).
// Panics if ctx is nil (same as context.WithValue).
func (k *Key[T]) With(ctx context.Context, val T) context.Context {
	if k == nil {
		panic("ctxval: Key must not be nil; use NewKey to create")
	}

	return context.WithValue(ctx, k, box[T]{v: val})
}

// Lookup retrieves the value under this key from ctx, returning a not-ok
// option if not present.
//
// Panics if k is nil (use [NewKey] to create keys).
// Panics if ctx is nil (same as context.Context.Value).
func (k *Key[T]) Lookup(ctx context.Context) option.Option[T] {
	if k == nil {
		panic("ctxval: Key must not be nil; use NewKey to create")
	}

	b, ok := ctx.Value(k).(box[T])

	return option.New(b.v, ok)
}
