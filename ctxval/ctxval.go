// Package ctxval provides typed helpers for storing and retrieving
// request-scoped values in context.Context.
//
// Values are keyed by their Go type. Each type can have at most one value
// in a context. Use named types to distinguish values of the same underlying
// type. Type aliases (=) share the same key as the aliased type:
//
//	type RequestID string
//	type TraceID string
//
//	ctx = ctxval.With(ctx, RequestID("abc"))
//	ctx = ctxval.With(ctx, TraceID("xyz"))
//	reqID := ctxval.Get[RequestID](ctx)  // Option[RequestID]("abc")
//	trID  := ctxval.Get[TraceID](ctx)    // Option[TraceID]("xyz")
//
// For code that only needs one value per type (e.g., middleware injecting
// a User into context), the package-level [With] and [Get] functions are
// convenient. For shared or public boundaries where multiple values of the
// same type coexist, prefer [Key] — it gives each value a unique identity
// without requiring a new named type:
//
//	var userKey = ctxval.NewKey[User]()
//
//	ctx = userKey.With(ctx, currentUser)
//	userOpt := userKey.From(ctx)  // Option[User]
//
// This package is for request-scoped data that crosses API boundaries
// (user IDs, trace IDs, auth tokens). It is not for optional parameters,
// dependency injection, or service locators.
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

// Get retrieves the value of type T from ctx, returning a not-ok option
// if no value of that type is present. T must match the exact static type
// used in [With]; interface and concrete types are distinct keys.
//
// Panics if ctx is nil (same as context.Context.Value).
func Get[T any](ctx context.Context) option.Option[T] {
	b, ok := ctx.Value(typeKey[T]{}).(box[T])

	return option.New(b.v, ok)
}

// Key is a named context key for values of type T. Unlike [With]/[Get],
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

// Get retrieves the value under this key from ctx, returning a not-ok
// option if not present.
//
// Panics if k is nil (use [NewKey] to create keys).
// Panics if ctx is nil (same as context.Context.Value).
func (k *Key[T]) From(ctx context.Context) option.Option[T] {
	if k == nil {
		panic("ctxval: Key must not be nil; use NewKey to create")
	}

	b, ok := ctx.Value(k).(box[T])

	return option.New(b.v, ok)
}
