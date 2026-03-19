# Order Processing Service

A curl-testable HTTP server that demonstrates how 7 fluentfp packages compose in a single application. See [COMPARISON.md](COMPARISON.md) for a side-by-side with conventional Go at every point where fluentfp is used.

```
go run ./examples/orders/
```

## Architecture

```
HTTP request (synchronous):
  decode (web) → validate (web.Steps) → enrich (hof.WithBreaker) → store → 201

Background (fire-and-forget):
  postCh → toc.Start → toc.Tee(2) → [audit Pipe | inventory Pipe]
```

The HTTP path is fully synchronous: the order is validated, enriched, and stored before the response is sent. This means validation errors return 400, pricing failures return 500, and a tripped circuit breaker returns 503 — all visible to the caller.

After storing, the order is sent to a background pipeline for post-processing (audit logging and inventory tracking). This work is best-effort: if the channel is full, it's skipped with a log message.

## Packages Used

### web — Typed HTTP Handlers

The `web` package replaces the standard `http.HandlerFunc(w, r)` pattern with handlers that return values:

```go
type Handler = func(*http.Request) rslt.Result[web.Response]
```

Instead of mutating a `ResponseWriter`, handlers return a `Result` — either an `Ok` response or an error. `web.Adapt` converts this back to a standard `http.HandlerFunc`, rendering the response or error as JSON.

**In this example:**

- `web.DecodeJSON[Order](req)` decodes the request body into a `Result[Order]`, returning structured errors (415 wrong content-type, 400 malformed JSON) without any manual error checking.

- `web.Steps(hasCustomer, hasItems, itemsHavePositiveQty)` chains validation functions that each return `Result[Order]`. The chain short-circuits on the first error — no nested `if err != nil` blocks.

- `web.Created(enriched)` and `web.OK(order)` construct typed responses with the right status code.

- `web.WithErrorMapper(mapDomainError)` maps domain errors to HTTP errors at the adapter boundary. Here it catches `hof.ErrCircuitOpen` and returns 503 instead of a generic 500.

- `web.BadRequest(msg)` and `web.NotFound(msg)` construct structured error responses with appropriate status codes and JSON bodies.

### rslt — Result Type for Error Composition

The `rslt` package provides `Result[T]` — a value that is either `Ok(value)` or `Err(error)`. It replaces the `(T, error)` return convention with something that composes.

**In this example:**

```go
validated := rslt.FlatMap(web.DecodeJSON[Order](req), validateOrder)
order, err := validated.Unpack()
```

`rslt.FlatMap` chains `DecodeJSON` (which returns `Result[Order]`) into `validateOrder` (which also returns `Result[Order]`). If decoding fails, validation is skipped. If validation fails, the error propagates. The entire decode-validate pipeline is one expression.

`Unpack()` converts back to Go's `(T, error)` convention when you need to branch.

### hof — Circuit Breaker

The `hof` package provides higher-order functions that wrap existing functions with cross-cutting behavior.

**In this example:**

```go
breaker := hof.NewBreaker(hof.BreakerConfig{
    ResetTimeout: 10 * time.Second,
    ReadyToTrip:  hof.ConsecutiveFailures(3),
})
enrichWithBreaker := hof.WithBreaker(breaker, enrichOrder)
```

`hof.WithBreaker` wraps `enrichOrder` (a `func(context.Context, Order) (Order, error)`) with circuit breaker protection. The wrapped function has the same signature — callers don't know it's protected. After 3 consecutive failures, the breaker opens and immediately rejects calls with `hof.ErrCircuitOpen` until the reset timeout expires.

The error mapper at the HTTP boundary catches `ErrCircuitOpen` and returns 503:

```go
func mapDomainError(err error) (*web.Error, bool) {
    if errors.Is(err, hof.ErrCircuitOpen) {
        return &web.Error{Status: 503, ...}, true
    }
    return nil, false
}
```

Try it: send 3 orders with SKU `"FAIL-PRICE"` to trip the breaker, then send a normal order — it gets 503.

### option — Safe Query Parameter Parsing

The `option` package provides `Option[T]` — a value that may or may not be present. It replaces `if s == ""` and `strconv.Atoi` error checking with composable operations.

**In this example:**

```go
status, hasStatus := option.NonEmpty(q.Get("status")).Get()
minTotalOpt := option.FlatMap(option.NonEmpty(q.Get("min_total")), option.Atoi)
mt, hasMinTotal := minTotalOpt.Get()
```

`option.NonEmpty` converts an empty string to `None` and a non-empty string to `Some(s)`. `option.FlatMap` chains it with `option.Atoi` — if the string is empty, parsing is skipped entirely. If the string is present but not a valid integer, the result is `None` and we return 400.

`Get()` returns `(value, ok)` — the same pattern as map lookup, but for optional values.

### slice — Conditional Filtering and Sorting

The `slice` package provides fluent operations on slices. Its key feature is method chaining: each operation returns a value you can call the next method on.

**In this example:**

```go
orders := slice.SortBy(s.list(), Order.GetID).
    KeepIfWhen(hasStatus, hasMatchingStatus).
    KeepIfWhen(hasMinTotal, totalAtLeast)
```

`slice.SortBy` sorts orders by ID using a method expression (`Order.GetID`) as the key function. The result chains directly into `KeepIfWhen`, which conditionally applies a filter: if `hasStatus` is true, keep only orders matching the status; if false, return all orders unchanged. This avoids `if`/`else` blocks that break the chain.

**Method expressions** are a key pattern: `Order.GetID` is a `func(Order) string` created from the `GetID` method. fluentfp APIs accept these directly, so you don't need to write `func(o Order) string { return o.ID }`.

### ctxval — Typed Context Values

The `ctxval` package provides type-safe `context.Value` storage. Instead of defining sentinel key types and doing type assertions, you store and retrieve values by their Go type.

**In this example:**

```go
// Middleware stores a request ID:
ctx := ctxval.With(r.Context(), RequestID(fmt.Sprintf("req-%d", reqCounter.Add(1))))

// Handler retrieves it:
reqID := ctxval.From[RequestID](req.Context()).Or("unknown")
```

`ctxval.With` stores a value keyed by its type (`RequestID`). `ctxval.From[RequestID]` retrieves it as an `Option[RequestID]` — if the middleware didn't run, `Or("unknown")` provides a fallback. No type assertions, no key collisions.

### toc — Bounded Concurrency Pipeline

The `toc` package provides composable pipeline stages with bounded concurrency, backpressure, and observability.

**In this example:**

```go
stage := toc.Start(ctx, passthrough, toc.Options[Order]{Capacity: 10, Workers: 1})
tee := toc.NewTee(ctx, stage.Out(), 2)
auditPipe := toc.Pipe(ctx, tee.Branch(0), logOrder, toc.Options[Order]{})
inventoryPipe := toc.Pipe(ctx, tee.Branch(1), countItems, toc.Options[Order]{})
```

`toc.Start` creates a stage that accepts orders via `Submit` and processes them through a function. `toc.NewTee` broadcasts each output to N branches — here, every order goes to both audit and inventory. `toc.Pipe` chains a processing function onto a branch's output channel.

The pipeline is long-lived: it starts once at server startup and processes orders continuously. Each stage has bounded capacity (backpressure), configurable worker count, and built-in stats tracking.

**Shutdown propagation:** When the server shuts down, the post-processing channel is closed, which causes the feeder goroutine to call `CloseInput()`, which drains through the stage, tee, and pipes in order. Context cancellation provides a backstop.

## Try It

```bash
go run ./examples/orders/

# Create an order
curl -s -X POST http://localhost:3000/orders \
  -H 'Content-Type: application/json' \
  -d '{"customer":"Alice","items":[{"sku":"WIDGET-1","quantity":3},{"sku":"GADGET-2","quantity":1}]}'

# Retrieve it
curl -s http://localhost:3000/orders/ord-1

# List with filters (total_cents: WIDGET-1 is 999 cents)
curl -s 'http://localhost:3000/orders?status=enriched'
curl -s 'http://localhost:3000/orders?min_total=3000'

# Validation error
curl -s -X POST http://localhost:3000/orders \
  -H 'Content-Type: application/json' \
  -d '{"customer":"","items":[]}'

# Trip the circuit breaker (3 failures)
for i in 1 2 3; do
  curl -s -X POST http://localhost:3000/orders \
    -H 'Content-Type: application/json' \
    -d '{"customer":"Bob","items":[{"sku":"FAIL-PRICE","quantity":1}]}'
done

# Next request gets 503 (breaker open)
curl -s -X POST http://localhost:3000/orders \
  -H 'Content-Type: application/json' \
  -d '{"customer":"Carol","items":[{"sku":"WIDGET-1","quantity":1}]}'

# Wait 10 seconds for breaker reset, then try again
```
