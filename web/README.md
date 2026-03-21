# web

JSON HTTP handlers that return values instead of mutating ResponseWriter.

```go
// Before: mutation, manual headers, repeated error blocks
func handleGetUser(w http.ResponseWriter, req *http.Request) {
    user, err := db.FindUser(req.PathValue("id"))
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(500)
        json.NewEncoder(w).Encode(map[string]string{"error": "internal error"})
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(200)
    json.NewEncoder(w).Encode(user)
}

// After: return a value, let Adapt handle rendering
handleGetUser := func(req *http.Request) rslt.Result[web.Response] {
    return rslt.Map(
        rslt.Of(db.FindUser(req.PathValue("id"))),
        web.OK[User],
    )
}
mux.HandleFunc("GET /users/{id}", web.Adapt(handleGetUser))
```

Twelve lines become four. No `ResponseWriter`, no manual headers, no `json.NewEncoder`. The handler is a testable expression — call it with a request, inspect the Result.

## What It Looks Like

```go
// Decode JSON request body — Content-Type, MaxBytes, UnknownFields handled
order, err := web.DecodeJSON[Order](req)  // (Order, error)
```

```go
// Validation chain — short-circuits on first error, each step carries HTTP status
validateOrder := web.Steps(hasCustomer, hasItems, itemsHavePositiveQty)
validated := rslt.FlatMap(order, validateOrder)
```

```go
// Response constructors — status code travels with the body
return rslt.Ok(web.Created(order))   // 201
return rslt.Ok(web.OK(order))        // 200
return rslt.Ok(web.NoContent())      // 204
```

```go
// Error constructors — structured JSON errors with status codes
return rslt.Err[web.Response](web.BadRequest("customer is required"))   // 400
return rslt.Err[web.Response](web.NotFound("order not found"))          // 404
return rslt.Err[web.Response](web.Conflict("order already exists"))     // 409
return rslt.Err[web.Response](web.Forbidden("insufficient permissions")) // 403
```

```go
// Error mapping — domain errors → HTTP errors, defined once at the boundary
mapDomainError := func(err error) (*web.Error, bool) {
    if errors.Is(err, call.ErrCircuitOpen) {
        return &web.Error{Status: 503, Message: "service unavailable"}, true
    }
    return nil, false
}
mux.HandleFunc("POST /orders",
    web.Adapt(handleCreateOrder, web.WithErrorMapper(mapDomainError)))
```

```go
// Custom decode options
order := web.DecodeJSONWith[Order](req, web.DecodeOpts{
    MaxBytes:     5 << 20,  // 5 MB
    AllowUnknown: true,     // don't reject unknown fields
})
```

## Error Rendering Flow

When a handler returns `Err`, Adapt renders the error as JSON:

1. `errors.As(err, &webErr)` → render `*web.Error` directly (status, message, code, details)
2. Else if `WithErrorMapper` is set → call mapper
3. If mapper returns `(*Error, true)` → render that
4. Else → 500 `{"error": "internal error"}`

Handlers don't write errors — they return them. Adapt decides how to render.

## Operations

**Handler + Adapt**
- `Handler` = `func(*http.Request) rslt.Result[Response]` — the handler type
- `Adapt(h Handler, opts ...AdaptOption) http.HandlerFunc` — bridge to stdlib
- `WithErrorMapper(fn func(error) (*Error, bool)) AdaptOption` — map domain errors to HTTP errors

**Params**
- `PathParam(req, name) Option[string]` — named path parameter, not-ok if missing/empty

**Decode**
- `DecodeJSON[T](req) (T, error)` — decode with defaults (1 MB limit, reject unknown fields, require `application/json`)
- `DecodeJSONWith[T](req, opts) (T, error)` — decode with custom policy

**Validate**
- `Steps[T](fns ...func(T) Result[T]) func(T) Result[T]` — chain validations, short-circuit on first error

**Response**
- `JSON[T](status, body) Response` — any status + body
- `OK[T](body) Response` — 200
- `Created[T](body) Response` — 201
- `NoContent() Response` — 204

**Errors**
- `BadRequest(msg) error` — 400
- `Forbidden(msg) error` — 403
- `NotFound(msg) error` — 404
- `Conflict(msg) error` — 409
- `TooManyRequests(msg) error` — 429
- `StatusError(status, code, msg) error` — custom status

## Rate Limiting Pattern

Rate limiting lives in middleware, not in fluentfp. The web package provides the error constructor; your middleware provides the logic:

```go
// Rate limit middleware using golang.org/x/time/rate
func withRateLimit(limit rate.Limit, burst int, next http.Handler) http.Handler {
    limiter := rate.NewLimiter(limit, burst)
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            // Return a web.Error — Adapt's error rendering handles the rest
            writeRateLimitError(w)
            return
        }
        next.ServeHTTP(w, r)
    })
}

func writeRateLimitError(w http.ResponseWriter) {
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Retry-After", "1")
    w.WriteHeader(http.StatusTooManyRequests)
    json.NewEncoder(w).Encode(web.ClientError{
        Error: "rate limit exceeded",
        Code:  "TOO_MANY_REQUESTS",
    })
}
```

For handlers that return `Result[Response]`, use `web.TooManyRequests` in an error mapper:

```go
mapRateLimit := func(err error) (*web.Error, bool) {
    if errors.Is(err, errRateLimited) {
        return &web.Error{
            Status:  429,
            Message: "rate limit exceeded",
            Code:    "TOO_MANY_REQUESTS",
            Headers: http.Header{"Retry-After": {"1"}},
        }, true
    }
    return nil, false
}
```

This package is for the transport boundary — JSON in, JSON out. Domain logic belongs in separate functions that handlers call. Not a web framework.

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/web) for complete API documentation and the [orders example](../examples/orders/) for a full integration demo.
