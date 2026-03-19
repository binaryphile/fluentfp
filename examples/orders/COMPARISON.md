# fluentfp vs Conventional Go — Side by Side

Every place the orders example uses fluentfp, shown next to the conventional Go equivalent.

## Handler Signature

**fluentfp:**
```go
handleCreateOrder := func(req *http.Request) rslt.Result[web.Response] {
    // ...
    return rslt.Ok(web.Created(enriched))
}

mux.HandleFunc("POST /orders", web.Adapt(handleCreateOrder, errorMapper))
```

**Conventional:**
```go
handleCreateOrder := func(w http.ResponseWriter, req *http.Request) {
    // ...
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(enriched)
}

mux.HandleFunc("POST /orders", handleCreateOrder)
```

The fluentfp handler returns a value. The conventional handler mutates a `ResponseWriter`. Returning a value means the handler is a pure expression — you can test it by calling it and inspecting the result, without constructing a `httptest.ResponseRecorder`.

## Request Decoding + Validation

**fluentfp:**
```go
validated := rslt.FlatMap(web.DecodeJSON[Order](req), validateOrder)
order, err := validated.Unpack()
if err != nil {
    return rslt.Err[web.Response](err)
}
```

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

if err := validateOrder(order); err != nil {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusBadRequest)
    json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
    return
}
```

`web.DecodeJSON` handles content-type checking, body size limits (1 MB default), unknown field rejection, and structured error responses in one call. `rslt.FlatMap` chains it with the validation pipeline — if decoding fails, validation is never called. The conventional version requires `MaxBytesReader` wrapping, a separate `DisallowUnknownFields` call, and manual error branching for each step. Validation is a separate function call that still needs its own error-to-response block (see next section).

## Validation Chain

**fluentfp:**
```go
validateOrder := web.Steps(hasCustomer, hasItems, itemsHavePositiveQty)
```

Each validator is a named function with the signature `func(Order) rslt.Result[Order]`:

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

`web.Steps` composes them into a single function that runs each in order and short-circuits on the first error. `slice.From(o.Items).Every(LineItem.HasPositiveQty)` replaces a manual loop with a method expression — the intent reads directly as "every line item has positive quantity."

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

The conventional version is one monolithic function. Adding a new validation means editing the function body. The fluentfp version is a list — adding a validator means adding a name to `web.Steps(...)`. Each validator is independently testable and reusable. The quantity check replaces a `for`/`if` loop with `Every` and a method expression — no iteration variable, no negated condition.

Note: the conventional validator also loses the HTTP status code. To return 400 vs 422 vs 409 for different validations, you'd need a custom error type and a switch statement in the handler. `web.BadRequest` and `web.Conflict` carry the status code with the error.

## Circuit Breaker Wrapping

**fluentfp:**
```go
breaker := hof.NewBreaker(hof.BreakerConfig{
    ResetTimeout: 10 * time.Second,
    ReadyToTrip:  hof.ConsecutiveFailures(3),
})
enrichWithBreaker := hof.WithBreaker(breaker, enrichOrder)

// In handler — same signature as enrichOrder:
enriched, err := enrichWithBreaker(req.Context(), order)
```

**Conventional:**
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
    case 0: // closed
        return true
    case 1: // open
        if time.Since(cb.openedAt) > cb.resetTimeout {
            cb.state = 2
            return true
        }
        return false
    case 2: // half-open
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

// In handler:
if !breaker.allow() {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusServiceUnavailable)
    json.NewEncoder(w).Encode(map[string]string{"error": "pricing service unavailable"})
    return
}
enriched, err := enrichOrder(req.Context(), order)
if err != nil {
    breaker.recordFailure()
    // handle error...
} else {
    breaker.recordSuccess()
}
```

`hof.WithBreaker` is a decorator — it wraps the function, preserving its signature. The handler doesn't know the breaker exists. The conventional version requires the handler to explicitly check breaker state, record outcomes, and write error responses — tangling three concerns (resilience policy, result recording, HTTP rendering) in one code block. Note: the hand-rolled version above is simplified and still has a bug (half-open rejects all requests instead of allowing a probe). A correct implementation would be longer. `hof.WithBreaker` handles half-open probe gating, panic recovery, context cancellation semantics, and concurrent state transitions.

## Error Mapping at the Boundary

**fluentfp:**
```go
func mapDomainError(err error) (*web.Error, bool) {
    if errors.Is(err, hof.ErrCircuitOpen) {
        return &web.Error{Status: 503, Message: "pricing service unavailable", Code: "SERVICE_UNAVAILABLE"}, true
    }
    return nil, false
}
errorMapper := web.WithErrorMapper(mapDomainError)

mux.HandleFunc("POST /orders", web.Adapt(handleCreateOrder, errorMapper))
```

**Conventional:**
```go
// Scattered across every handler that calls the breaker-wrapped function:
enriched, err := enrichOrder(req.Context(), order)
if err != nil {
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
}
```

The fluentfp version defines the mapping once, at the adapter boundary. Every handler behind that adapter gets the same error-to-HTTP translation. The conventional version duplicates the error checking and response writing in every handler that could encounter that error.

## Query Parameter Parsing

**fluentfp:**
```go
status, hasStatus := option.NonEmpty(q.Get("status")).Get()
minTotalOpt := option.FlatMap(option.NonEmpty(q.Get("min_total")), option.Atoi)
if raw, ok := option.NonEmpty(q.Get("min_total")).Get(); ok {
    if _, ok := minTotalOpt.Get(); !ok {
        return rslt.Err[web.Response](web.BadRequest(
            fmt.Sprintf("min_total must be an integer (cents), got %q", raw)))
    }
}
mt, hasMinTotal := minTotalOpt.Get()
```

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

The line counts are similar, but the shape differs. The conventional version requires pre-declaring `var mt int; var hasMinTotal bool` as mutable variables because the assignment happens inside a conditional block. The fluentfp version computes `minTotalOpt` as a single immutable value — `option.NonEmpty` → `option.FlatMap` → `option.Atoi` is a pipeline where each step handles the absent case automatically. No mutable state, no declaration-before-assignment.

## Conditional List Filtering

**fluentfp:**
```go
orders := slice.SortBy(s.list(), Order.GetID).
    KeepIfWhen(hasStatus, hasMatchingStatus).
    KeepIfWhen(hasMinTotal, totalAtLeast)
```

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

The fluentfp version reads as a single pipeline: sort, then conditionally filter by status, then conditionally filter by total. `KeepIfWhen(cond, fn)` applies the filter only when `cond` is true, so optional filters don't break the chain.

The conventional version declares `filtered` twice, iterates twice, and appends element by element. The `sort.Slice` closure takes `i, j int` indices and accesses the outer `orders` slice — the sort key (`ID`) is buried inside the comparison expression.

`slice.SortBy(list, Order.GetID)` uses a method expression as the sort key. The intent — "sort by ID" — is the entire expression.

## Typed Context Values

**fluentfp:**
```go
// Store:
ctx := ctxval.With(r.Context(), RequestID("req-1"))

// Retrieve:
reqID := ctxval.From[RequestID](req.Context()).Or("unknown")
```

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

`ctxval.With` and `ctxval.From` use the Go type itself as the key — no sentinel values to define or manage. `From` returns an `Option`, so the fallback is `.Or("unknown")` instead of a type assertion and `if !ok` branch.

## Response Construction

**fluentfp:**
```go
return rslt.Ok(web.Created(enriched))     // 201
return rslt.Ok(web.OK(order))             // 200
return rslt.Err[web.Response](web.BadRequest("customer is required"))  // 400
return rslt.Err[web.Response](web.NotFound("order not found"))         // 404
```

**Conventional:**
```go
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusCreated)
json.NewEncoder(w).Encode(enriched)

w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(order)

w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusBadRequest)
json.NewEncoder(w).Encode(map[string]string{"error": "customer is required", "code": "BAD_REQUEST"})

w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusNotFound)
json.NewEncoder(w).Encode(map[string]string{"error": "order not found", "code": "NOT_FOUND"})
```

Each conventional response is 3 lines of mutation: set header, write status, encode body. The fluentfp versions are expressions that carry type, status, and body as data. The `Adapt` function handles the actual `ResponseWriter` interaction once.

## Pipeline Construction (toc)

**fluentfp:**
```go
stage := toc.Start(ctx, passthrough, toc.Options[Order]{Capacity: 10, Workers: 1})
tee := toc.NewTee(ctx, stage.Out(), 2)
auditPipe := toc.Pipe(ctx, tee.Branch(0), logOrder, toc.Options[Order]{})
inventoryPipe := toc.Pipe(ctx, tee.Branch(1), countItems, toc.Options[Order]{})
```

**Conventional:**
```go
orderCh := make(chan Order, 10)
auditCh := make(chan Order, 10)
inventoryCh := make(chan Order, 10)

// Worker
go func() {
    for o := range orderCh {
        auditCh <- o
        inventoryCh <- o
    }
    close(auditCh)
    close(inventoryCh)
}()

// Audit branch
go func() {
    for o := range auditCh {
        log.Printf("audit: order %s (%d cents)", o.ID, o.TotalCents)
    }
}()

// Inventory branch
go func() {
    for o := range inventoryCh {
        log.Printf("inventory: %d items", len(o.Items))
    }
}()
```

The conventional version is shorter for this simple case. The toc version pays for itself when you need:

- **Backpressure**: `toc.Start` blocks `Submit` when the capacity buffer is full. The conventional version blocks the broadcaster goroutine if either branch channel is full, stalling both branches.
- **Observability**: Every toc stage tracks `Stats()` — submitted, completed, failed, service time, idle time, output blocked time. The conventional version has no metrics.
- **Error handling**: toc stages propagate errors as `rslt.Result` values through the pipeline. The conventional version has no error channel.
- **Cancellation**: toc stages respond to context cancellation and drain cleanly. The conventional version requires manual `select` on `ctx.Done()` in every goroutine.
- **Shutdown ordering**: Closing the toc pipeline input propagates through `Stage → Tee → Pipe` in order. The conventional version requires manual channel closing in the right sequence.

For two goroutines doing trivial work, the conventional version is fine. For a pipeline with bounded concurrency, error propagation, and operational visibility, the manual channel management becomes the bulk of the code.

In this example, the toc stages expose `Stats()` (e.g., `stage.Stats().Submitted`, `tee.Stats().FullyDelivered`) that could be surfaced via a `/debug/pipeline` endpoint. The conventional version would need custom atomic counters added to each goroutine.
