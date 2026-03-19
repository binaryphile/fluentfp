# Conventional Go vs fluentfp — Complete Side by Side

This document shows every place the orders example uses fluentfp, paired with the conventional Go equivalent. Read the [README](README.md) first for the narrative walkthrough.

## The Complete POST Handler

This is the same handler shown in the README. The conventional version is what you'd write with only the standard library.

**Conventional (52 lines):**

```go
func handleCreateOrder(w http.ResponseWriter, req *http.Request) {
    if req.Header.Get("Content-Type") != "application/json" {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnsupportedMediaType)
        json.NewEncoder(w).Encode(map[string]string{"error": "expected application/json"})
        return
    }

    var order Order
    dec := json.NewDecoder(http.MaxBytesReader(w, req.Body, 1<<20))
    dec.DisallowUnknownFields()
    if err := dec.Decode(&order); err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }

    if order.Customer == "" {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": "customer is required"})
        return
    }
    if len(order.Items) == 0 {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": "order must have at least one item"})
        return
    }

    order.ID = fmt.Sprintf("ord-%d", idCounter.Add(1))
    order.Status = "pending"

    if !breaker.allow() {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(map[string]string{"error": "pricing service unavailable"})
        return
    }
    enriched, err := enrichOrder(req.Context(), order)
    if err != nil {
        breaker.recordFailure()
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": "internal error"})
        return
    }
    breaker.recordSuccess()

    store.put(enriched)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(enriched)
}
```

**fluentfp (20 lines):**

```go
handleCreateOrder := func(req *http.Request) rslt.Result[web.Response] {
    validated := rslt.FlatMap(web.DecodeJSON[Order](req), validateOrder)
    order, err := validated.Unpack()
    if err != nil {
        return rslt.Err[web.Response](err)
    }

    order.ID = fmt.Sprintf("ord-%d", idCounter.Add(1))
    order.Status = "pending"

    enriched, err := enrichWithBreaker(req.Context(), order)
    if err != nil {
        return rslt.Err[web.Response](err)
    }

    s.put(enriched)
    return rslt.Ok(web.Created(enriched))
}
```

52 lines → 20 lines. The reduction comes from:

- **6 response-writing blocks eliminated.** Each `w.Header().Set` / `w.WriteHeader` / `json.NewEncoder` block is 3-4 lines. `web.Adapt` handles all response rendering once, outside the handler.
- **Content-Type / MaxBytesReader / DisallowUnknownFields eliminated.** `web.DecodeJSON` does all three in one call.
- **Validation inlined into the decode pipeline.** `rslt.FlatMap(decode, validate)` replaces separate decode and validate error blocks.
- **Circuit breaker invisible.** `enrichWithBreaker` has the same signature as `enrichOrder`. No `allow()` / `recordSuccess()` / `recordFailure()` in the handler.
- **Error mapping centralized.** `ErrCircuitOpen` → 503 happens at the adapter boundary, not in the handler.

The conventional version also excludes the 40+ lines needed to implement `circuitBreaker` itself. fluentfp provides it as `hof.WithBreaker`.

Now let's look at each piece individually.

---

## Request Decoding

**Conventional:**
```go
if req.Header.Get("Content-Type") != "application/json" {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusUnsupportedMediaType)
    json.NewEncoder(w).Encode(map[string]string{"error": "expected application/json"})
    return
}

var order Order
dec := json.NewDecoder(http.MaxBytesReader(w, req.Body, 1<<20))
dec.DisallowUnknownFields()
if err := dec.Decode(&order); err != nil {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusBadRequest)
    json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
    return
}
```

15 lines: content-type check, body size limit, decoder setup, decode call, error response. Three separate concerns (security, parsing, error rendering) interleaved.

**fluentfp:**
```go
web.DecodeJSON[Order](req)
```

One call. Returns `Result[Order]` — either the decoded order or a structured error with the right HTTP status code (415 for wrong content-type, 413 for too large, 400 for malformed JSON or unknown fields).

## Validation

**Conventional:**
```go
func validateOrder(o Order) error {
    if o.Customer == "" {
        return fmt.Errorf("customer is required")
    }
    if len(o.Items) == 0 {
        return fmt.Errorf("order must have at least one item")
    }
    for _, item := range o.Items {
        if item.Quantity <= 0 {
            return fmt.Errorf("all items must have positive quantity")
        }
    }
    return nil
}
```

One monolithic function. Adding a validation means editing the body. Every error returns a bare `error` — to map different validations to different HTTP status codes (400 vs 422 vs 409), you'd need a custom error type and a switch statement.

**fluentfp:**
```go
validateOrder := web.Steps(hasCustomer, hasItems, itemsHavePositiveQty)
```

A list of named functions. Adding a validation means adding a name. Each function is independently testable and carries its own HTTP status code:

```go
func hasCustomer(o Order) rslt.Result[Order] {
    if o.Customer == "" {
        return rslt.Err[Order](web.BadRequest("customer is required"))
    }
    return rslt.Ok(o)
}

func itemsHavePositiveQty(o Order) rslt.Result[Order] {
    if !slice.From(o.Items).Every(LineItem.HasPositiveQty) {
        return rslt.Err[Order](web.BadRequest("all items must have positive quantity"))
    }
    return rslt.Ok(o)
}
```

The quantity check uses `slice.Every` with a method expression instead of a `for`/`if` loop. "Every item has positive quantity" reads directly as English.

## Chaining Decode into Validation

**Conventional:**
```go
// decode...
if err := dec.Decode(&order); err != nil {
    // write 400 response...
    return
}
if err := validateOrder(order); err != nil {
    // write 400 response...
    return
}
```

Two separate error checks, each with its own response block. If you add a third step (normalize, enrich, transform), you add a third block.

**fluentfp:**
```go
validated := rslt.FlatMap(web.DecodeJSON[Order](req), validateOrder)
order, err := validated.Unpack()
if err != nil {
    return rslt.Err[web.Response](err)
}
```

`rslt.FlatMap` chains two operations that each return `Result`. If the first fails, the second is skipped. Adding a third step is another `FlatMap` call, not another error block.

## Circuit Breaker

**Conventional (the breaker itself — 40+ lines):**
```go
type circuitBreaker struct {
    mu                  sync.Mutex
    failures            int
    state               int // 0=closed, 1=open, 2=half-open
    openedAt            time.Time
    resetTimeout        time.Duration
    consecutiveFailures int
}

func (cb *circuitBreaker) allow() bool {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    switch cb.state {
    case 0:
        return true
    case 1:
        if time.Since(cb.openedAt) > cb.resetTimeout {
            cb.state = 2
            return true
        }
        return false
    case 2:
        return false
    }
    return false
}

func (cb *circuitBreaker) recordSuccess() {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    cb.failures = 0
    cb.state = 0
}

func (cb *circuitBreaker) recordFailure() {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    cb.failures++
    if cb.failures >= cb.consecutiveFailures {
        cb.state = 1
        cb.openedAt = time.Now()
    }
}
```

Note: this simplified version has a bug — half-open rejects all requests instead of allowing one probe. A correct implementation is longer.

**Conventional (in the handler — 12 lines):**
```go
if !breaker.allow() {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusServiceUnavailable)
    json.NewEncoder(w).Encode(map[string]string{"error": "pricing service unavailable"})
    return
}
enriched, err := enrichOrder(req.Context(), order)
if err != nil {
    breaker.recordFailure()
    // write 500 response...
} else {
    breaker.recordSuccess()
}
```

Three concerns tangled: breaker state checking, result recording, and HTTP response writing.

**fluentfp (setup — 5 lines):**
```go
breaker := hof.NewBreaker(hof.BreakerConfig{
    ResetTimeout: 10 * time.Second,
    ReadyToTrip:  hof.ConsecutiveFailures(3),
})
enrichWithBreaker := hof.WithBreaker(breaker, enrichOrder)
```

**fluentfp (in the handler — 1 line):**
```go
enriched, err := enrichWithBreaker(req.Context(), order)
```

Same signature as `enrichOrder`. The handler doesn't know a breaker exists. `hof.WithBreaker` handles probe gating, panic recovery, context cancellation, and concurrent state transitions.

## Error Mapping

**Conventional (repeated in every handler):**
```go
if errors.Is(err, errCircuitOpen) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusServiceUnavailable)
    json.NewEncoder(w).Encode(map[string]string{"error": "pricing service unavailable"})
    return
}
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusInternalServerError)
json.NewEncoder(w).Encode(map[string]string{"error": "internal error"})
return
```

Every handler that calls the enrichment function needs this block.

**fluentfp (defined once):**
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

One mapping function, applied at the adapter boundary. Every handler behind that adapter gets the same translation. The handler just returns the error; the boundary decides how to render it.

## Resource Lookup (GET Handler)

**Conventional:**
```go
order, ok := s.get(id)
if !ok {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusNotFound)
    json.NewEncoder(w).Encode(map[string]string{"error": "order not found", "code": "NOT_FOUND"})
    return
}

w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(order)
```

Two branches with identical response-writing boilerplate. The lookup, the absence handling, and the response rendering are all mixed together.

**fluentfp:**
```go
return rslt.Map(
    option.New(s.get(id)).OkOr(web.NotFound("order not found")),
    web.OK[Order],
)
```

Three operations compose into one return:
1. `option.New(s.get(id))` — wraps `(Order, bool)` into `Option[Order]`
2. `.OkOr(web.NotFound(...))` — present becomes `Ok(order)`, absent becomes `Err(404)`
3. `rslt.Map(..., web.OK[Order])` — wraps the Ok value in a 200 response

The 404 propagates through `Map` untouched. No intermediate variables, no branching.

## Map Lookup with Fallback

**Conventional:**
```go
price, ok := prices[item.SKU]
if !ok {
    price = 100
}
```

Three lines: declare, check, reassign.

**fluentfp:**
```go
price := option.Lookup(prices, item.SKU).Or(100)
```

One line. `Lookup` wraps the comma-ok map access into an `Option`. `.Or(100)` provides the fallback.

## Query Parameter Parsing

**Conventional:**
```go
status := q.Get("status")
hasStatus := status != ""

var mt int
var hasMinTotal bool
if raw := q.Get("min_total"); raw != "" {
    parsed, err := strconv.Atoi(raw)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{
            "error": fmt.Sprintf("min_total must be an integer (cents), got %q", raw),
        })
        return
    }
    mt = parsed
    hasMinTotal = true
}
```

Mutable variables declared before the conditional block, then assigned inside it. The parse error check is nested inside the emptiness check.

**fluentfp:**
```go
status, hasStatus := option.NonEmpty(q.Get("status")).Get()
minTotalOption := option.FlatMap(option.NonEmpty(q.Get("min_total")), option.Atoi)
mt, hasMinTotal := minTotalOption.Get()
```

`option.NonEmpty` → `option.FlatMap` → `option.Atoi` is a parse pipeline: empty → skip, non-empty → parse integer. Each step handles the absent case automatically. No mutable variables, no declaration-before-assignment.

## Conditional List Filtering

**Conventional:**
```go
orders := s.list()
sort.Slice(orders, func(i, j int) bool {
    return orders[i].ID < orders[j].ID
})

if hasStatus {
    var filtered []Order
    for _, o := range orders {
        if o.Status == status {
            filtered = append(filtered, o)
        }
    }
    orders = filtered
}

if hasMinTotal {
    var filtered []Order
    for _, o := range orders {
        if o.TotalCents >= mt {
            filtered = append(filtered, o)
        }
    }
    orders = filtered
}
```

25 lines. `sort.Slice` buries the sort key inside an `i, j` closure. Each optional filter declares `filtered`, loops, appends, and reassigns. The pattern repeats identically for each filter.

**fluentfp:**
```go
orders := slice.SortBy(s.list(), Order.GetID).
    KeepIfWhen(hasStatus, hasMatchingStatus).
    KeepIfWhen(hasMinTotal, totalAtLeast)
```

3 lines. `SortBy` takes a method expression as the key — "sort by ID" is the entire expression. `KeepIfWhen` applies the filter only when the condition is true, so optional filters chain without `if` blocks.

## Context Values

**Conventional:**
```go
type contextKey struct{}
var requestIDKey = contextKey{}

// Store:
ctx := context.WithValue(r.Context(), requestIDKey, "req-1")

// Retrieve:
reqID, ok := req.Context().Value(requestIDKey).(string)
if !ok {
    reqID = "unknown"
}
```

Requires a private key type (sentinel), a type assertion, and a nil/missing check.

**fluentfp:**
```go
// Store:
ctx := ctxval.With(r.Context(), RequestID("req-1"))

// Retrieve:
reqID := ctxval.From[RequestID](req.Context()).Or("unknown")
```

The Go type itself is the key. `From` returns an `Option`, so `.Or("unknown")` is the fallback. No sentinel types, no type assertions.

## Response Construction

**Conventional (each response is 3 lines):**
```go
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusCreated)
json.NewEncoder(w).Encode(enriched)
```

Repeated for every code path. Miss `Content-Type` and you get `text/plain`. Call `WriteHeader` after `Write` and it's silently ignored.

**fluentfp (each response is 1 expression):**
```go
return rslt.Ok(web.Created(enriched))                                   // 201
return rslt.Ok(web.OK(order))                                          // 200
return rslt.Err[web.Response](web.BadRequest("customer is required"))   // 400
return rslt.Err[web.Response](web.NotFound("order not found"))          // 404
```

Status code, body, and headers travel together as data. `web.Adapt` renders them once. You can't forget `Content-Type` or misordering `WriteHeader` because you never call either.

## Background Pipeline

**Conventional:**
```go
orderCh := make(chan Order, 10)
auditCh := make(chan Order, 10)
inventoryCh := make(chan Order, 10)

go func() {
    for o := range orderCh {
        auditCh <- o
        inventoryCh <- o
    }
    close(auditCh)
    close(inventoryCh)
}()

go func() {
    for o := range auditCh {
        log.Printf("audit: order %s (%d cents)", o.ID, o.TotalCents)
    }
}()

go func() {
    for o := range inventoryCh {
        log.Printf("inventory: %d items", len(o.Items))
    }
}()
```

Three channels, three goroutines, manual close sequencing. No backpressure — if the audit goroutine is slow, the fan-out goroutine blocks, stalling inventory too. No metrics. No error propagation. Adding cancellation means adding `select` + `ctx.Done()` to every goroutine.

**fluentfp:**
```go
stage := toc.Start(ctx, passthrough, toc.Options[Order]{Capacity: 10, Workers: 1})
tee := toc.NewTee(ctx, stage.Out(), 2)
auditPipe := toc.Pipe(ctx, tee.Branch(0), logOrder, toc.Options[Order]{})
inventoryPipe := toc.Pipe(ctx, tee.Branch(1), countItems, toc.Options[Order]{})
```

Four lines. Backpressure, cancellation, and shutdown ordering are built in. Each stage exposes `Stats()` — submitted, completed, failed, service time, idle time, blocked time — ready for a `/debug` endpoint. The conventional version would need custom atomic counters in each goroutine.

For trivial fan-out, the conventional version works. For anything you need to operate, monitor, or shut down cleanly, the channel plumbing becomes the bulk of the code.
