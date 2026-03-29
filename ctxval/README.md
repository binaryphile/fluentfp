# ctxval

Type-safe context values without sentinel keys or type assertions.

```go
// Before: sentinel key + WithValue + type assertion + nil check
type contextKey struct{}
var requestIDKey = contextKey{}

ctx = context.WithValue(ctx, requestIDKey, "req-123")

reqID, ok := ctx.Value(requestIDKey).(string)
if !ok {
    reqID = "unknown"
}

// After
type RequestID string

ctx = ctxval.With(ctx, RequestID("req-123"))

reqID := ctxval.Lookup[RequestID](ctx).Or("unknown")
```

Seven lines become two. No sentinel type to define, no type assertion to get wrong, no nil check to forget.

## What It Looks Like

```go
// Store a value by its type
ctx := ctxval.With(ctx, RequestID("req-123"))

// Retrieve with fallback
reqID := ctxval.Lookup[RequestID](ctx).Or("unknown")
```

```go
// Multiple named types coexist — each type is its own key
type RequestID string
type TraceID string

ctx = ctxval.With(ctx, RequestID("req-123"))
ctx = ctxval.With(ctx, TraceID("trace-abc"))
reqID := ctxval.Lookup[RequestID](ctx)  // Option("req-123")
trID  := ctxval.Lookup[TraceID](ctx)    // Option("trace-abc")
```

```go
// Comma-ok extraction when you need to branch
if user, ok := ctxval.Lookup[User](ctx).Get(); ok {
    log.Printf("request from %s", user.Name)
}
```

```go
// Named keys — when multiple values of the same type coexist
var authTokenKey = ctxval.NewKey[string]()
var csrfTokenKey = ctxval.NewKey[string]()

ctx = authTokenKey.With(ctx, "bearer-xyz")
ctx = csrfTokenKey.With(ctx, "csrf-abc")
auth, _ := authTokenKey.Lookup(ctx).Get()  // "bearer-xyz"
csrf, _ := csrfTokenKey.Lookup(ctx).Get()  // "csrf-abc"
```

## When to Use What

**Type-keyed** (`With`/`Lookup`): When each Go type naturally maps to one value per context. Middleware injecting a `User`, `RequestID`, or `TraceID` — one value per type, no collision possible.

**Named keys** (`NewKey`/`Key.With`/`Key.Lookup`): When multiple values share the same underlying type at a shared API boundary. Two different `string` tokens, two different `int` IDs. Each `NewKey` call creates a unique key by pointer identity.

Type aliases (`type Alias = string`) share the key with the aliased type. Named types (`type RequestID string`) get their own key.

## Operations

- **Store**: `With[T](ctx, val)` — store by type; `Key.With(ctx, val)` — store by named key
- **Retrieve**: `Lookup[T](ctx)` — returns `Option[T]`; `Key.Lookup(ctx)` — returns `Option[T]`
- **Keys**: `NewKey[T]()` — create unique named key (pointer identity)

`Lookup` returns an `Option`, so all option operations apply: `.Or("default")`, `.Get()`, `.IsOk()`, `.IfOk(fn)`.

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/ctxval) for complete API documentation and the [orders example](../examples/orders/) for a full integration demo with middleware.
