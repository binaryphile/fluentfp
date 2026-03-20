# Order Processing Service

```
go run ./examples/orders/
```

## Start Here: One Handler, Before and After

Every Go developer has written this handler. Decode JSON, validate, call a service, write the response. Here's what that looks like for creating an order:

**Conventional Go (50 lines):**

```go
func handleCreateOrder(
    w http.ResponseWriter, req *http.Request,
) {
    if req.Header.Get("Content-Type") != "application/json" {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(415)
        json.NewEncoder(w).Encode(
            map[string]string{
                "error": "expected application/json",
            })
        return
    }

    var order Order
    dec := json.NewDecoder(
        http.MaxBytesReader(w, req.Body, 1<<20))
    dec.DisallowUnknownFields()
    if err := dec.Decode(&order); err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(400)
        json.NewEncoder(w).Encode(
            map[string]string{"error": err.Error()})
        return
    }

    if order.Customer == "" {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(400)
        json.NewEncoder(w).Encode(
            map[string]string{"error": "customer is required"})
        return
    }
    // ... more validation ...

    if !breaker.allow() {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(503)
        json.NewEncoder(w).Encode(
            map[string]string{
                "error": "pricing service unavailable",
            })
        return
    }
    enriched, err := enrichOrder(req.Context(), order)
    if err != nil {
        breaker.recordFailure()
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(500)
        json.NewEncoder(w).Encode(
            map[string]string{"error": "internal error"})
        return
    }
    breaker.recordSuccess()

    store.put(enriched)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(201)
    json.NewEncoder(w).Encode(enriched)
}
```

Count the `w.Header().Set` / `w.WriteHeader` / `json.NewEncoder` blocks: six of them. Each one is 3 lines of identical response-writing machinery. The actual business logic -- decode, validate, enrich, store -- is buried between them.

**With fluentfp:**

```go
handleCreateOrder := func(
    req *http.Request,
) rslt.Result[web.Response] {
    // Get request ID from context (set by middleware)
    reqID := ctxval.From[RequestID](req.Context()).Or("unknown")

    // Bind context to breaker-wrapped call for pipeline use
    enrich := rslt.LiftCtx(req.Context(), enrichWithBreaker)
    // Side effects: log errors, persist + notify on success
    logFailure := func(err error) { ... }
    storeAndNotify := func(o Order) { ... }

    // Pipeline: each step operates on the Result from the
    // previous step. If any step fails, the rest are skipped.
    orderResult := web.DecodeJSON[Order](req) // -> Result
    storedResult := orderResult.
        FlatMap(validateOrder).      // validate (-> 400)
        Transform(withNewID).            // assign ID + status
        FlatMap(enrich).             // pricing (-> 502/503)
        TapErr(logFailure).              // on error: log it
        Tap(storeAndNotify)              // on success: persist
    return rslt.Map(storedResult, web.Created[Order]) // 201
}
```

No `if err != nil`. Each line in the pipeline does one thing:

- **FlatMap** -- run a function that can fail; skip the rest on error
- **Transform** -- same-type transform: `func(T) T` (can be pure or side-effecting)
- **TapErr** -- side effect on error (doesn't change the error)
- **Tap** -- side effect on success (doesn't change the value)
- **rslt.Map** -- change the type (Order -> Response); standalone because Go methods can't change the generic type

The handler returns a value instead of mutating a `ResponseWriter`. All the response rendering -- headers, status codes, JSON encoding, error formatting -- happens once in `web.Adapt`.

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
mux.HandleFunc("POST /orders",
    web.Adapt(handleCreateOrder, errorMapper))
```

The response constructors -- `web.Created`, `web.OK`, `web.BadRequest`, `web.NotFound` -- carry the status code with them. No more remembering to call `WriteHeader(201)` vs `WriteHeader(200)`.

### Chaining decode into validation (rslt.FlatMap, web.Steps)

Decoding a JSON body correctly in Go requires:

1. Check Content-Type
2. Wrap body with MaxBytesReader
3. Create decoder, set DisallowUnknownFields
4. Decode, check error
5. Run validations, check each error

`web.DecodeJSON[Order](req)` does steps 1-4 in one call, returning `Result[Order]`. Then `rslt.FlatMap` chains it into validation:

```go
orderResult := web.DecodeJSON[Order](req)
validatedResult := orderResult.FlatMap(validateOrder)
```

`FlatMap` chains two operations that each return `Result`. If `orderResult` is Err, `validateOrder` is skipped entirely. If validation fails, that error propagates. The "flat" part: since `validateOrder` itself returns `Result[Order]`, a plain `Map` would nest that into `Result[Result[Order]]`. `FlatMap` keeps it one level deep.

In practice: if decoding fails (wrong Content-Type -> 415, malformed JSON -> 400), validation is never called. If validation fails, that error propagates. The chain continues with `.Transform(withNewID).FlatMap(enrich)` -- each step feeds the next, errors propagate automatically.

The validation chain is a list of named functions:

```go
validateOrder := web.Steps(
    hasCustomer, hasItems, itemsHavePositiveQty)
```

Each validator has the signature `func(Order) rslt.Result[Order]` and carries its own HTTP status code:

```go
func itemsHavePositiveQty(o Order) rslt.Result[Order] {
    if !slice.From(o.Items).Every(LineItem.HasPositiveQty) {
        return rslt.Err[Order](
            web.BadRequest(
                "all items must have positive quantity"))
    }
    return rslt.Ok(o)
}
```

Adding a validation means adding a name to `web.Steps(...)`. Each validator is independently testable. In conventional Go, this is a monolithic function where every check returns a bare `error` that loses the HTTP status code.

### Wrapping functions with resilience (call)

The pricing service might be slow or failing. In conventional Go, adding a circuit breaker means writing 40+ lines of mutex-protected state (open/closed/half-open), then threading check/record/branch logic through the handler. The breaker becomes more code than the call it protects.

With fluentfp, it's a decorator:

```go
breaker := call.NewBreaker(call.BreakerConfig{
    ResetTimeout: 10 * time.Second,
    ReadyToTrip:  call.ConsecutiveFailures(3),
})
enrichWithBreaker := call.WithBreaker(breaker, enrichOrder)
```

`enrichWithBreaker` has the same signature as `enrichOrder`. The handler calls it without knowing a breaker exists. After 3 consecutive failures, it rejects with `call.ErrCircuitOpen`.

In the POST handler, `rslt.LiftCtx` partially applies the request context, producing a `func(Order) Result[Order]` that slots directly into the FlatMap chain:

```go
enrich := rslt.LiftCtx(req.Context(), enrichWithBreaker)
```

No closure needed -- LiftCtx does the wrapping.

At the HTTP boundary, one error mapper translates all domain errors to HTTP responses:

```go
func mapDomainError(err error) (*web.Error, bool) {
    if errors.Is(err, call.ErrCircuitOpen) {
        return &web.Error{
            Status: 503, Message: "pricing service unavailable",
        }, true
    }
    if errors.Is(err, errPricingFailure) {
        return &web.Error{
            Status: 502, Message: "pricing service error",
        }, true
    }
    return nil, false
}
```

Defined once, applied to every handler via `web.WithErrorMapper`. Breaker-open -> 503, pricing failure -> 502. Unknown SKUs are caught earlier in validation (returns 400) so they never reach the breaker -- bad input can't trip the circuit.

**Try it:** send 3 orders with SKU `"FAIL-PRICE"`, then one with a normal SKU -- it gets 503.

### Replacing if-not-ok with Option->Result (option)

Looking up a resource and returning 404 on miss is two branches with identical boilerplate in conventional Go. With fluentfp, it's a pipeline:

```go
handleGetOrder := func(
    req *http.Request,
) rslt.Result[web.Response] {
    // Extract path param; absent -> 400
    idResult := web.PathParam(req, "id").
        OkOr(web.BadRequest("missing order id"))
    // Look up order; missing -> 404
    foundResult := rslt.FlatMap(idResult, findOrder)
    // Wrap in 200 response (standalone Map -- type changes)
    return rslt.Map(foundResult, web.OK[Order])
}
```

Three steps, each building on the last:

1. `web.PathParam` wraps `PathValue` into `Option[string]`; `.OkOr(...)` bridges absent -> `Err(400)`
2. `rslt.FlatMap(idResult, findOrder)` looks up the order (standalone FlatMap because string -> Order is a type change)
3. `rslt.Map(foundResult, web.OK[Order])` wraps the Ok value in a 200 response

`findOrder` is a named function that bridges the store lookup:

```go
findOrder := func(id string) rslt.Result[Order] {
    return option.New(s.get(id)).
        OkOr(web.NotFound("order not found"))
}
```

If any step fails, the error propagates through FlatMap and Map untouched.

When a map lookup needs a fallback instead of an error, `option.Lookup(m, k).Or(default)` replaces the entire `if !ok` block:

```go
price := option.Lookup(prices, item.SKU).Or(100)
```

### Parsing and filtering without if-blocks (option, slice)

Query parameters are optional. In conventional Go, each one requires checking emptiness, parsing, handling parse errors, and conditionally filtering -- each adding an `if` block that breaks the flow.

With fluentfp, parsing is a pipeline and filtering is a chain:

```go
// NonEmpty: "" -> not-ok, non-empty -> ok. Get: (value, bool).
status, hasStatus := option.NonEmpty(q.Get("status")).Get()

// Parse min_total: absent -> Ok(not-ok), valid -> Ok(Of(n)),
// invalid -> Err(400). FlatMapResult bridges
// optional + fallible.
parseMinTotal := func(raw string) rslt.Result[int] {
    return option.Atoi(raw).OkOr(web.BadRequest(
        fmt.Sprintf(
            "min_total must be an integer (cents), got %q",
            raw)))
}
rawMinTotalOption := option.NonEmpty(q.Get("min_total"))
minTotalResult := option.FlatMapResult(
    rawMinTotal, parseMinTotal)
mtOption, err := minTotalResult.Unpack()
if err != nil {
    return rslt.Err[web.Response](err)
}
mt, hasMinTotal := mtOption.Get()

// Named predicates for optional filters.
hasMatchingStatus := func(o Order) bool {
    return o.Status == status
}
totalAtLeast := func(o Order) bool {
    return o.TotalCents >= mt
}

// SortBy sorts by key function. KeepIf keeps elements
// where the predicate returns true (like filter()).
orders := slice.SortBy(s.list(), orderNum)
if hasStatus {
    orders = orders.KeepIf(hasMatchingStatus)
}
if hasMinTotal {
    orders = orders.KeepIf(totalAtLeast)
}
```

`option.FlatMapResult` bridges the gap between optional and fallible: absent -> `Ok(NotOk)` (no filter applied), present+valid -> `Ok(Of(n))`, present+invalid -> `Err(400)`. This cleanly distinguishes "not provided" from "provided but wrong" -- a common pattern for optional query parameters.

`slice.SortBy(list, orderNum)` sorts by a key function -- `orderNum` extracts the numeric suffix from `"ord-N"` for correct numeric ordering. The conditional `KeepIf` filters are plain Go `if` statements -- no special API needed when you're not inside a chain that would break.

The conventional equivalent uses `sort.Slice` with an `i, j` closure that accesses the outer slice by index, and `var filtered []Order` with `append` inside a `for` loop for each filter. The fluentfp version uses `SortBy` with a key function and `KeepIf` with named predicates -- same structure, less scaffolding.

### Type-safe context values (ctxval)

Go's `context.WithValue` requires a private key type, a type assertion on retrieval, and a nil check. Three lines of ceremony per value.

```go
// Middleware -- store by type, no key needed:
ctx := ctxval.With(r.Context(), RequestID("req-1"))

// Handler -- retrieve with fallback:
reqID := ctxval.From[RequestID](req.Context()).Or("unknown")
```

`ctxval.From` returns an `Option`, so `.Or("unknown")` is the fallback. No sentinel types, no type assertions.

### Bounded background pipelines (toc)

After storing the order, the handler sends it to a background pipeline for audit logging and inventory tracking. In conventional Go this means creating channels, writing fan-out goroutines, and manually sequencing channel closes. None of it has metrics.

With toc, the pipeline is three lines:

```go
// Bridge plain channel, broadcast to 2 branches
tee := toc.NewTee(ctx, toc.FromChan(postCh), 2)
// Branch 0: audit log
auditPipe := toc.Pipe(
    ctx, tee.Branch(0), logOrder, toc.Options[Order]{})
// Branch 1: inventory count
inventoryPipe := toc.Pipe(
    ctx, tee.Branch(1), countItems, toc.Options[Order]{})
```

`toc.FromChan` bridges the plain `chan Order` into the `chan rslt.Result[Order]` that toc operators expect -- no passthrough stage needed. `Tee` broadcasts each order to both branches. `Pipe` chains a function onto each branch. The pipeline runs for the lifetime of the server.

What toc gives you that bare goroutines don't:

- **Backpressure** -- `Submit` blocks when the buffer is full, so you don't silently drop work
- **Stats** -- `stage.Stats()` tracks submitted, completed, failed, service time, idle time, blocked time -- ready for a `/debug` endpoint
- **Shutdown ordering** -- closing input propagates through Stage -> Tee -> Pipe; context cancellation is a backstop
- **Error propagation** -- errors flow as `Result` values instead of disappearing into goroutine logs

## Architecture

```
HTTP request (synchronous):
  decode (web) -> validate (web.Steps) -> enrich (call.WithBreaker) -> store -> 201

Background (fire-and-forget):
  postCh -> toc.FromChan -> toc.Tee(2) -> [audit Pipe | inventory Pipe]
```

The HTTP path is fully synchronous: the order is validated, enriched, and stored before the response is sent. Validation errors (including unknown SKUs) return 400, pricing failures return 502, and a tripped circuit breaker returns 503.

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
