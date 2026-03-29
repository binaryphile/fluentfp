package ctxval_test

import (
	"context"
	"fmt"

	"github.com/binaryphile/fluentfp/ctxval"
)

// ExRequestID is a distinct type for request IDs in context.
type ExRequestID string

func ExampleWith() {
	// Store a request ID in context, keyed by type.
	ctx := ctxval.With(context.Background(), ExRequestID("req-abc-123"))

	id := ctxval.Lookup[ExRequestID](ctx).Or("unknown")
	fmt.Println(id)
	// Output: req-abc-123
}

func ExampleLookup() {
	// Lookup returns not-ok when the type isn't in the context.
	ctx := context.Background()

	id := ctxval.Lookup[ExRequestID](ctx)
	fmt.Println(id.IsOk())
	fmt.Println(id.Or("fallback"))
	// Output:
	// false
	// fallback
}

func ExampleNewKey() {
	// Two keys of the same type don't collide.
	adminKey := ctxval.NewKey[string]()
	userKey := ctxval.NewKey[string]()

	ctx := context.Background()
	ctx = adminKey.With(ctx, "root")
	ctx = userKey.With(ctx, "alice")

	fmt.Println(adminKey.From(ctx).Or(""))
	fmt.Println(userKey.From(ctx).Or(""))
	// Output:
	// root
	// alice
}
