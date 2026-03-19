# Order Processing Service

```
go run ./examples/orders/
```

## Start Here: One Handler, Before and After

Every Go developer has written this handler. Decode JSON, validate, call a service, write the response. Here's what that looks like for creating an order:

**Conventional Go (50 lines):**

```go
func handleCreateOrder(w http.ResponseWriter, req *http.Request) {
    if req.Header.Get("Content-Type") != "application/json" {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(415)
        json.NewEncoder(w).Encode(map[string]string{"error": "expected application/json"})
        return
    }

    var order Order
    dec := json.NewDecoder(http.MaxBytesReader(w, req.Body, 1<<20))
    dec.DisallowUnknownFields()
    if err := dec.Decode(&order); err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(400)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }

    if order.Customer == "" {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(400)
        json.NewEncoder(w).Encode(map[string]string{"error": "customer is required"})
        return
    }
    // ... more validation ...

    if !breaker.allow() {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(503)
        json.NewEncoder(w).Encode(map[string]string{"error": "pricing service unavailable"})
        return
    }
    enriched, err := enrichOrder(req.Context(), order)
    if err != nil {
        breaker.recordFailure()
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(500)
        json.NewEncoder(w).Encode(map[string]string{"error": "internal error"})
        return
    }
    breaker.recordSuccess()

    store.put(enriched)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(201)
    json.NewEncoder(w).Encode(enriched)
}
```

Count the `w.Header().Set` / `w.WriteHeader` / `json.NewEncoder` blocks: six of them. Each one is 3 lines of identical response-writing machinery. The actual business logic — decode, validate, enrich, store — is buried between them.

**With fluentfp:**

```go
handleCreateOrder := func(req *http.Request) rslt.Result[web.Response] {
    reqID := ctxval.From[RequestID](req.Context()).Or("unknown")

    // named functions: assignID, enrich, logFailure, storeAndNotify (defined above)

    order := web.DecodeJSON[Order](req)
    validatedOrder := rslt.FlatMap(order, validateOrder)
    assignedOrder := validatedOrder.Transform(withNewID)
    enrichedOrder := rslt.FlatMap(assignedOrder, enrich)
    storedOrder := enrichedOrder.MapErr(logFailure).Tap(storeAndNotify)
    return rslt.Map(storedOrder, web.Created[Order])
}
```

No `if err != nil`. The handler is a pipeline: each line is one operation on a `Result`. `FlatMap` chains operations that can fail (decode→validate, enrich). `Transform` transforms the Ok value (assign ID, store+notify). `MapErr` handles the Err side (log failure). `Map` wraps the final value in a response.

The handler returns a value instead of mutating a `ResponseWriter`. All the response rendering — headers, status codes, JSON encoding, error formatting — happens once in `web.Adapt`.

That's the core idea. Everything below builds on it.

## How It Works, Piece by Piece

### Returning values instead of mutating (web, rslt)

The standard `http.HandlerFunc` takes a `ResponseWriter` and mutates it. That means every code path must remember to set headers, write the status, encode the body, and return. Miss any step and you get silent bugs (headers after body, missing Content-Type, double writes).

fluentfp's `web.Handler` returns a `rslt.Result[web.Response]`:

```go
type Handler = func(*http.Request) rslt.Result[web.Response]
```

A `Result` is either `Ok(value)` or `Err(error)`. The handler returns one or the other. `web.Adapt` converts it to a standard `http.HandlerFunc`:

```go
mux.HandleFunc("POST /orders", web.Adapt(handleCreateOrder, errorMapper))
```

The response constructors — `web.Created`, `web.OK`, `web.BadRequest`, `web.NotFound` — carry the status code with them. No more remembering to call `WriteHeader(201)` vs `WriteHeader(200)`.

### Chaining decode into validation (rslt.FlatMap, web.Steps)

Decoding a JSON body correctly in Go requires:

1. Check Content-Type
2. Wrap body with MaxBytesReader
3. Create decoder, set DisallowUnknownFields
4. Decode, check error
5. Run validations, check each error

`web.DecodeJSON[Order](req)` does steps 1-4 in one call, returning `Result[Order]`. Then `rslt.FlatMap` chains it into validation:

```go
order := web.DecodeJSON[Order](req)
validatedOrder := rslt.FlatMap(order, validateOrder)
```

`FlatMap` takes a `Result` and a function that also returns a `Result`. If the input is `Ok`, it unwraps the value and passes it to the function. If the input is `Err`, it skips the function entirely and returns the error. The "flat" part: since `validateOrder` itself returns `Result[Order]`, a plain `Map` would give you `Result[Result[Order]]` — nested results. `FlatMap` flattens that to a single `Result[Order]`.

In practice: if decoding fails (wrong Content-Type → 415, malformed JSON → 400), validation is never called. If validation fails, that error propagates. Two lines replace 20 lines of decode-check-validate-check.

The validation chain is a list of named functions:

```go
validateOrder := web.Steps(hasCustomer, hasItems, itemsHavePositiveQty)
```

Each validator has the signature `func(Order) rslt.Result[Order]` and carries its own HTTP status code:

```go
func itemsHavePositiveQty(o Order) rslt.Result[Order] {
    if !slice.From(o.Items).Every(LineItem.HasPositiveQty) {
        return rslt.Err[Order](web.BadRequest("all items must have positive quantity"))
    }
    return rslt.Ok(o)
}
```

Adding a validation means adding a name to `web.Steps(...)`. Each validator is independently testable. In conventional Go, this is a monolithic function where every check returns a bare `error` that loses the HTTP status code.

### Wrapping functions with resilience (hof)

The pricing service might be slow or failing. In conventional Go, adding a circuit breaker means writing 40+ lines of mutex-protected state (open/closed/half-open), then threading check/record/branch logic through the handler. The breaker becomes more code than the call it protects.

With fluentfp, it's a decorator:

```go
breaker := hof.NewBreaker(hof.BreakerConfig{
    ResetTimeout: 10 * time.Second,
    ReadyToTrip:  hof.ConsecutiveFailures(3),
})
enrichWithBreaker := hof.WithBreaker(breaker, enrichOrder)
```

`enrichWithBreaker` has the same signature as `enrichOrder`. The handler calls it without knowing a breaker exists. After 3 consecutive failures, it rejects with `hof.ErrCircuitOpen`.

At the HTTP boundary, one error mapper catches it:

```go
func mapDomainError(err error) (*web.Error, bool) {
    if errors.Is(err, hof.ErrCircuitOpen) {
        return &web.Error{Status: 503, Message: "pricing service unavailable"}, true
    }
    return nil, false
}

errorMapper := web.WithErrorMapper(mapDomainError)
mux.HandleFunc("POST /orders", web.Adapt(handleCreateOrder, errorMapper))
```

Defined once, applied to every handler. In conventional Go, every handler that calls the breaker-wrapped function needs its own `if errors.Is` block with response writing.

**Try it:** send 3 orders with SKU `"FAIL-PRICE"`, then one with a normal SKU — it gets 503.

### Replacing if-not-ok with Option→Result (option)

Looking up a resource and returning 404 on miss is two branches with identical boilerplate in conventional Go. With fluentfp, it's a pipeline:

```go
handleGetOrder := func(req *http.Request) rslt.Result[web.Response] {
    id := req.PathValue("id")
    if id == "" {
        return rslt.Err[web.Response](web.BadRequest("missing order id"))
    }

    return rslt.Map(
        option.New(s.get(id)).OkOr(web.NotFound("order not found")),
        web.OK[Order],
    )
}
```

Three operations, no intermediate variables:

1. `option.New(s.get(id))` — wraps Go's `(Order, bool)` return into `Option[Order]`
2. `.OkOr(web.NotFound(...))` — present → `Ok(order)`, absent → `Err(404)`
3. `rslt.Map(..., web.OK[Order])` — wraps the Ok value in a 200 response

If the lookup missed, the 404 propagates through `Map` untouched.

The same `option` package handles map lookups with fallbacks:

```go
price := option.Lookup(prices, item.SKU).Or(100) // unknown SKU: $1.00
```

One line replaces `price, ok := m[key]; if !ok { price = fallback }`.

### Parsing and filtering without if-blocks (option, slice)

Query parameters are optional. In conventional Go, each one requires checking emptiness, parsing, handling parse errors, and conditionally filtering — each adding an `if` block that breaks the flow.

With fluentfp, parsing is a pipeline and filtering is a chain:

```go
status, hasStatus := option.NonEmpty(q.Get("status")).Get()
minTotalOption := option.FlatMap(option.NonEmpty(q.Get("min_total")), option.Atoi)
mt, hasMinTotal := minTotalOption.Get()

hasMatchingStatus := func(o Order) bool { return o.Status == status }
totalAtLeast := func(o Order) bool { return o.TotalCents >= mt }

orders := slice.SortBy(s.list(), Order.GetID).
    KeepIfWhen(hasStatus, hasMatchingStatus).
    KeepIfWhen(hasMinTotal, totalAtLeast)
```

`option.NonEmpty` → `option.FlatMap` → `option.Atoi` chains empty-check into integer parsing. The option handles the absent case automatically at every step.

`slice.SortBy(list, Order.GetID)` sorts using a method expression — "sort by ID" is the entire expression. `KeepIfWhen(cond, fn)` applies a filter only when the condition is true. Two optional filters chain without `if` blocks breaking the pipeline.

The conventional equivalent: `var filtered []Order` declared twice, two `for` loops with `append`, and `sort.Slice` with an `i, j` closure that accesses the outer slice by index.

### Type-safe context values (ctxval)

Go's `context.WithValue` requires a private key type, a type assertion on retrieval, and a nil check. Three lines of ceremony per value.

```go
// Middleware — store by type, no key needed:
ctx := ctxval.With(r.Context(), RequestID("req-1"))

// Handler — retrieve with fallback:
reqID := ctxval.From[RequestID](req.Context()).Or("unknown")
```

`ctxval.From` returns an `Option`, so `.Or("unknown")` is the fallback. No sentinel types, no type assertions.

### Bounded background pipelines (toc)

After storing the order, the handler sends it to a background pipeline for audit logging and inventory tracking. In conventional Go this means creating channels, writing fan-out goroutines, and manually sequencing channel closes. None of it has metrics.

With toc, the pipeline is four lines:

```go
stage := toc.Start(ctx, passthrough, toc.Options[Order]{Capacity: 10, Workers: 1})
tee := toc.NewTee(ctx, stage.Out(), 2)
auditPipe := toc.Pipe(ctx, tee.Branch(0), logOrder, toc.Options[Order]{})
inventoryPipe := toc.Pipe(ctx, tee.Branch(1), countItems, toc.Options[Order]{})
```

`Start` creates a bounded stage. `Tee` broadcasts each order to both branches. `Pipe` chains a function onto each branch. The pipeline runs for the lifetime of the server.

What toc gives you that bare goroutines don't:

- **Backpressure** — `Submit` blocks when the buffer is full, so you don't silently drop work
- **Stats** — `stage.Stats()` tracks submitted, completed, failed, service time, idle time, blocked time — ready for a `/debug` endpoint
- **Shutdown ordering** — closing input propagates through Stage → Tee → Pipe; context cancellation is a backstop
- **Error propagation** — errors flow as `Result` values instead of disappearing into goroutine logs

## Architecture

```
HTTP request (synchronous):
  decode (web) → validate (web.Steps) → enrich (hof.WithBreaker) → store → 201

Background (fire-and-forget):
  postCh → toc.Start → toc.Tee(2) → [audit Pipe | inventory Pipe]
```

The HTTP path is fully synchronous: the order is validated, enriched, and stored before the response is sent. Validation errors return 400, pricing failures return 500, and a tripped circuit breaker returns 503.

After storing, the order is sent to a background pipeline for post-processing. This is best-effort: if the channel is full, it's skipped with a log message.

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

## Detailed Comparison

See [COMPARISON.md](COMPARISON.md) for a side-by-side with conventional Go at every point where fluentfp is used.
