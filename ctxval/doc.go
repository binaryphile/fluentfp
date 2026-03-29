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
//	reqID := ctxval.Lookup[RequestID](ctx)  // Option[RequestID]("abc")
//	trID  := ctxval.Lookup[TraceID](ctx)    // Option[TraceID]("xyz")
//
// For code that only needs one value per type (e.g., middleware injecting
// a User into context), the package-level [With] and [Lookup] functions are
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
