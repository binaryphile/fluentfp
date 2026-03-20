# Conventional Go vs fluentfp -- Complete Side by Side

This document shows every place the orders example uses fluentfp, paired with the conventional Go equivalent. Read the [README](README.md) first for the narrative walkthrough.

## Request Lifecycle

```mermaid
flowchart TD
    A["Decode JSON body"] --> B["Validate fields"]
    B --> C["Assign ID"]
    C --> D["Enrich via pricing service"]
    D --> E{"Success?"}
    E -->|Yes| F["Store order"]
    E -->|No| G["Log failure"]
    F --> H["Send to post-processing"]
    H --> I["Return 201 Created"]
    G --> J["Return error response"]

    style A fill:#e3f2fd,stroke:#1976d2
    style B fill:#e3f2fd,stroke:#1976d2
    style C fill:#fff3e0,stroke:#f57c00
    style D fill:#e8f5e9,stroke:#388e3c
    style F fill:#fff3e0,stroke:#f57c00
    style H fill:#f3e5f5,stroke:#7b1fa2
    style I fill:#e3f2fd,stroke:#1976d2
    style G fill:#ffebee,stroke:#c62828
    style J fill:#e3f2fd,stroke:#1976d2
```

| Step | Conventional Go | fluentfp | What changes |
|------|----------------|----------|--------------|
| **Decode** | Content-Type check + MaxBytesReader + DisallowUnknownFields + Decode + error response (15 lines) | `web.DecodeJSON[Order](req)` (1 line) | All decoding policy in one call |
| **Validate** | Monolithic `validateOrder()` returning bare `error` + error response block (15 lines) | `.FlatMap(validateOrder)` where `validateOrder = web.Steps(...)` (1 line) | Composable list of named validators, each carrying its HTTP status |
| **Assign ID** | `order.ID = ...; order.Status = ...` mutating in place (2 lines) | `.Transform(withNewID)` -- pure transform on Ok value (1 line) | Mutation wrapped in named function |
| **Enrich** | `breaker.allow()` check + call + `recordSuccess`/`recordFailure` + error response (12 lines, plus 40+ line breaker impl) | `.FlatMap(enrich)` in the chain (1 line) | Breaker is a decorator -- invisible to caller |
| **Log failure** | `log.Printf` inside `if err != nil` branch (1 line, tangled with response writing) | `.TapErr(logFailure)` -- error-side side effect in pipeline (1 line) | Logging separated from response rendering |
| **Store + notify** | `store.put` + `log` + channel send (6 lines) | `.Tap(storeAndNotify)` -- side effects in named function (1 line) | Side effects named and composable |
| **Respond** | `w.Header().Set` + `WriteHeader` + `Encode` (3 lines, repeated 6×) | `rslt.Map(stored, web.Created[Order])` (1 line) | `Adapt` renders once |
| **Error -> HTTP** | `if errors.Is(err, ...)` + response block, repeated per handler | `web.WithErrorMapper(mapDomainError)` defined once at boundary | One mapping function for all handlers |

## The Complete POST Handler

The conventional version on the left is what you'd write with the standard library. The fluentfp version on the right has the same behavior. Each section below shows the same step side by side.

### Signature

<table>
<tr><th>Conventional</th><th>fluentfp</th></tr>
<tr>
<td>

```go
func handleCreateOrder(
  w http.ResponseWriter,
  req *http.Request,
) {
```

Mutates `ResponseWriter`.

</td>
<td>

```go
handleCreateOrder := func(
  req *http.Request,
) rslt.Result[web.Response] {
```

Returns a value. `web.Adapt` renders it.

</td>
</tr>
</table>

### Decode

<table>
<tr><th>Conventional</th><th>fluentfp</th></tr>
<tr>
<td>

```go
ct := req.Header.Get("Content-Type")
if ct != "application/json" {
  w.Header().Set("Content-Type",
    "application/json")
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
  w.Header().Set("Content-Type",
    "application/json")
  w.WriteHeader(400)
  json.NewEncoder(w).Encode(
    map[string]string{
      "error": err.Error(),
    })
  return
}
```

</td>
<td>

```go
// One call: content-type, size limit,
// unknown fields, JSON decode.
order := web.DecodeJSON[Order](req)
```

Returns `Result[Order]` -- Ok or Err with the right HTTP status.

</td>
</tr>
</table>

### Validate

<table>
<tr><th>Conventional</th><th>fluentfp</th></tr>
<tr>
<td>

```go
if order.Customer == "" {
  w.Header().Set("Content-Type",
    "application/json")
  w.WriteHeader(400)
  json.NewEncoder(w).Encode(
    map[string]string{
      "error": "customer is required",
    })
  return
}
if len(order.Items) == 0 {
  // ... same 7-line block ...
}
```

Each check repeats the response block.

</td>
<td>

```go
// FlatMap: if decode failed, skip.
// If validation fails, stop chain.
stored := order.
  FlatMap(validateOrder). ...
```

`validateOrder = web.Steps(...)` -- a list of named validators, each returning `Result[Order]`.

</td>
</tr>
</table>

### Assign ID

<table>
<tr><th>Conventional</th><th>fluentfp</th></tr>
<tr>
<td>

```go
order.ID = fmt.Sprintf(
  "ord-%d", idCounter.Add(1))
order.Status = "pending"
```

Mutates `order` in place.

</td>
<td>

```go
  ... .Transform(withNewID). ...
```

`withNewID` returns a new Order with ID set. `Transform` applies `func(T) T` to the Ok value.

</td>
</tr>
</table>

### Enrich (circuit breaker)

<table>
<tr><th>Conventional</th><th>fluentfp</th></tr>
<tr>
<td>

```go
if !breaker.allow() {
  w.Header().Set("Content-Type",
    "application/json")
  w.WriteHeader(503)
  json.NewEncoder(w).Encode(
    map[string]string{
      "error": "pricing unavailable",
    })
  return
}
enriched, err := enrichOrder(
  req.Context(), order)
if err != nil {
  breaker.recordFailure()
  log.Printf(
    "enrichment failed: %v", err)
  // ... 7-line error response ...
  return
}
breaker.recordSuccess()
```

Breaker check + call + record + error response tangled.

</td>
<td>

```go
  ... .FlatMap(enrich). ...
```

`enrich = rslt.LiftCtx(ctx, enrichWithBreaker)` -- binds context, wraps `(T, error)` -> `Result[T]`. Breaker is invisible to the pipeline.

</td>
</tr>
</table>

### Error logging + Store + Respond

<table>
<tr><th>Conventional</th><th>fluentfp</th></tr>
<tr>
<td>

```go
store.put(enriched)
log.Printf(
  "created order %s", enriched.ID)
select {
case postCh <- enriched:
default:
  log.Printf("channel full, skipping")
}
w.Header().Set("Content-Type",
  "application/json")
w.WriteHeader(201)
json.NewEncoder(w).Encode(enriched)
```

</td>
<td>

```go
  ... .TapErr(logFailure).  // log err
      .Tap(storeAndNotify)  // persist
return rslt.Map(
  stored, web.Created[Order]) // 201
```

`TapErr`: side effect on error. `Tap`: side effect on success. `rslt.Map`: wrap Ok in 201 response.

</td>
</tr>
</table>

### What the pipeline eliminates

| Conventional | fluentfp | Why |
|---|---|---|
| 6× `Set`/`WriteHeader`/`Encode` | Zero | `Adapt` renders once |
| Content-Type + MaxBytesReader + DisallowUnknownFields | `DecodeJSON` | One call |
| Separate decode/validate error blocks | `FlatMap` | Chains; errors propagate |
| `allow`/`recordSuccess`/`recordFailure` | `WithBreaker` | Decorator |
| `errors.Is` in handler | `WithErrorMapper` | At boundary |
| `log.Printf` in error branches | `TapErr` | Pipeline side effect |

The conventional version also excludes the 40+ lines to implement `circuitBreaker` itself.

---

## Each Piece, Individually

### Request Decoding

<table>
<tr><th>Conventional</th><th>fluentfp</th></tr>
<tr>
<td>

```go
if req.Header.Get("Content-Type") !=
    "application/json" {
  w.Header().Set("Content-Type",
    "application/json")
  w.WriteHeader(415)
  json.NewEncoder(w).Encode(
    map[string]string{
      "error": "expected application/json"})
  return
}
var order Order
dec := json.NewDecoder(
  http.MaxBytesReader(w, req.Body, 1<<20))
dec.DisallowUnknownFields()
if err := dec.Decode(&order); err != nil {
  w.Header().Set("Content-Type",
    "application/json")
  w.WriteHeader(400)
  json.NewEncoder(w).Encode(
    map[string]string{"error": err.Error()})
  return
}
```

</td>
<td>

```go
order := web.DecodeJSON[Order](req)
```

Returns `Result[Order]` -- the decoded order or a structured error (415, 413, 400). Content-type, body size limit, unknown fields -- all handled.

</td>
</tr>
</table>

### Validation

<table>
<tr><th>Conventional</th><th>fluentfp</th></tr>
<tr>
<td>

```go
func validateOrder(o Order) error {
  if o.Customer == "" {
    return fmt.Errorf("customer is required")
  }
  if len(o.Items) == 0 {
    return fmt.Errorf("must have items")
  }
  for _, item := range o.Items {
    if item.Quantity <= 0 {
      return fmt.Errorf("positive qty required")
    }
  }
  return nil
}
```

One monolithic function. Bare `error` loses the HTTP status code.

</td>
<td>

```go
validateOrder := web.Steps(
  hasCustomer, hasItems, itemsHavePositiveQty)
```

A list of named functions. Each carries its own status code:

```go
func itemsHavePositiveQty(o Order) rslt.Result[Order] {
  if !slice.From(o.Items).Every(LineItem.HasPositiveQty) {
    return rslt.Err[Order](
      web.BadRequest("positive qty required"))
  }
  return rslt.Ok(o)
}
```

Adding a validation = adding a name to the list.

</td>
</tr>
</table>

### Chaining Decode -> Validate -> Transform

This is where `FlatMap` and `Transform` replace `if err != nil` blocks.

<table>
<tr><th>Conventional</th><th>fluentfp</th></tr>
<tr>
<td>

```go
// decode
if err := dec.Decode(&order); err != nil {
  // write 400...
  return
}
// validate
if err := validateOrder(order); err != nil {
  // write 400...
  return
}
// assign
order.ID = fmt.Sprintf("ord-%d",
  idCounter.Add(1))
order.Status = "pending"
```

Three blocks. Each fallible step needs its own error check and response. The assignment mutates `order` in place.

</td>
<td>

```go
order := web.DecodeJSON[Order](req)  // -> Result
assigned := order.
  FlatMap(validateOrder).            // can fail -> stops chain
  Transform(withNewID)               // pure transform on Ok
```

A chain. `FlatMap` passes the Ok value to `validateOrder`; if it returns Err, the chain stops. `Transform` applies `withNewID` (no error possible). Each method returns a `Result`, so the chain continues.

`FlatMap` is called "flat" because `validateOrder` returns `Result[Order]`: a plain `Map` would nest that into `Result[Result[Order]]`. FlatMap keeps it one level deep.

</td>
</tr>
</table>

### Circuit Breaker

<table>
<tr><th>Conventional</th><th>fluentfp</th></tr>
<tr>
<td>

```go
// 40+ lines of breaker implementation:
type circuitBreaker struct {
  mu       sync.Mutex
  failures int
  state    int // 0=closed, 1=open, 2=half-open
  openedAt time.Time
  // ...
}
func (cb *circuitBreaker) allow() bool { ... }
func (cb *circuitBreaker) recordSuccess() { ... }
func (cb *circuitBreaker) recordFailure() { ... }

// In handler (12 lines):
if !breaker.allow() {
  w.Header().Set("Content-Type",
    "application/json")
  w.WriteHeader(503)
  json.NewEncoder(w).Encode(
    map[string]string{
      "error": "pricing unavailable"})
  return
}
enriched, err := enrichOrder(
  req.Context(), order)
if err != nil {
  breaker.recordFailure()
  // write 500 response...
} else {
  breaker.recordSuccess()
}
```

Three concerns tangled: state checking, result recording, response writing. The hand-rolled breaker also has a bug (half-open rejects all requests).

</td>
<td>

```go
// Setup (once): create breaker, wrap enrichOrder.
// WithBreaker preserves the function signature.
breaker := call.NewBreaker(call.BreakerConfig{
  ResetTimeout: 10 * time.Second,
  ReadyToTrip:  call.ConsecutiveFailures(3),
})
enrichWithBreaker := call.WithBreaker(breaker, enrichOrder)

// In handler: LiftCtx binds context and wraps
// (Order, error) -> Result[Order] for the pipeline.
enrich := rslt.LiftCtx(req.Context(), enrichWithBreaker)

// In the pipeline chain:
.FlatMap(enrich)  // can fail -> stops the chain
```

Same signature as `enrichOrder`. The handler doesn't know a breaker exists. Probe gating, panic recovery, cancellation handled.

</td>
</tr>
</table>

### Error Logging + Error Mapping

<table>
<tr><th>Conventional</th><th>fluentfp</th></tr>
<tr>
<td>

```go
// In handler -- logging + error mapping tangled:
enriched, err := enrichOrder(
  req.Context(), order)
if err != nil {
  log.Printf("enrichment failed: %v", err)
  if errors.Is(err, errCircuitOpen) {
    w.Header().Set("Content-Type",
      "application/json")
    w.WriteHeader(503)
    json.NewEncoder(w).Encode(
      map[string]string{
        "error": "pricing unavailable"})
    return
  }
  w.Header().Set("Content-Type",
    "application/json")
  w.WriteHeader(500)
  json.NewEncoder(w).Encode(
    map[string]string{
      "error": "internal error"})
  return
}
```

Logging, error classification, and response rendering -- all in one `if err` block. Repeated per handler.

</td>
<td>

```go
// Side effect on error -- doesn't change the error.
logFailure := func(err error) {
  log.Printf("[%s] enrichment failed: %v", reqID, err)
}
// In the chain: TapErr runs on Err, Tap runs on Ok.
// ... .TapErr(logFailure).Tap(storeAndNotify)

// Error mapping (defined once, at Adapt boundary).
// Adapt calls this for errors that aren't already *web.Error.
func mapDomainError(err error) (*web.Error, bool) {
  if errors.Is(err, call.ErrCircuitOpen) {
    return &web.Error{Status: 503, Message: "unavailable"}, true
  }
  if errors.Is(err, errPricingFailure) {
    return &web.Error{
      Status: 502, Message: "pricing error",
    }, true
  }
  return nil, false
  // Unknown SKUs are caught in validation (400),
  // so they never reach the breaker.
}
web.Adapt(handler, web.WithErrorMapper(
  mapDomainError))
```

Three concerns, three locations: `TapErr` logs in the pipeline (side effect, error unchanged), `WithErrorMapper` classifies at the boundary, `Adapt` renders.

</td>
</tr>
</table>

### Resource Lookup (GET Handler)

<table>
<tr><th>Conventional</th><th>fluentfp</th></tr>
<tr>
<td>

```go
order, ok := s.get(id)
if !ok {
  w.Header().Set("Content-Type",
    "application/json")
  w.WriteHeader(404)
  json.NewEncoder(w).Encode(
    map[string]string{
      "error": "order not found",
      "code":  "NOT_FOUND"})
  return
}
w.Header().Set("Content-Type",
  "application/json")
w.WriteHeader(200)
json.NewEncoder(w).Encode(order)
```

Two branches, identical boilerplate.

</td>
<td>

```go
// PathParam wraps PathValue -> Option[string].
// OkOr bridges absent -> Err(400).
id := web.PathParam(req, "id").
  OkOr(web.BadRequest("missing order id"))
// findOrder: Option.New(store lookup).OkOr(404).
// Standalone FlatMap because string -> Order is cross-type.
found := rslt.FlatMap(id, findOrder)
// Standalone Map because Order -> Response is cross-type.
return rslt.Map(found, web.OK[Order])
```

`web.PathParam` wraps `PathValue` + `NonEmpty` into `Option[string]`. `.OkOr` bridges to `Result[string]`. `rslt.FlatMap` calls `findOrder` (which uses `option.New` + `.OkOr` to bridge the store lookup). `rslt.Map` wraps Ok in a 200 response. Errors propagate untouched.

</td>
</tr>
</table>

### Map Lookup

The pricing function uses `prices[item.SKU]` directly -- SKU validation already ran in the validation chain, so every SKU here is known-good. No error check needed.

`option.Lookup` earns its keep when you need a fallback: `option.Lookup(m, k).Or(default)` replaces the entire `if !ok` block in one expression.

### Query Parameter Parsing

<table>
<tr><th>Conventional</th><th>fluentfp</th></tr>
<tr>
<td>

```go
status := q.Get("status")
hasStatus := status != ""

var mt int
var hasMinTotal bool
if raw := q.Get("min_total"); raw != "" {
  parsed, err := strconv.Atoi(raw)
  if err != nil {
    w.Header().Set("Content-Type",
      "application/json")
    w.WriteHeader(400)
    json.NewEncoder(w).Encode(
      map[string]string{
        "error": fmt.Sprintf(
          "min_total must be int, got %q",
          raw)})
    return
  }
  mt = parsed
  hasMinTotal = true
}
```

Mutable variables declared before the conditional, assigned inside it.

</td>
<td>

```go
// NonEmpty: "" -> not-ok, non-empty -> ok.
// Get: unpack to (value, bool).
status, hasStatus := option.NonEmpty(q.Get("status")).Get()
rawMinTotal := option.NonEmpty(q.Get("min_total"))
// FlatMapResult: absent -> Ok(not-ok), valid -> Ok(Of(n)),
// invalid -> Err(400). Three-way optional+fallible.
minTotalResult := option.FlatMapResult(
  rawMinTotal, parseMinTotal)
// Unpack: convert Result back to Go's (value, error).
mtOption, err := minTotalResult.Unpack()
mt, hasMinTotal := mtOption.Get()
```

`FlatMapResult` bridges optional and fallible: absent -> `Ok(NotOk)`, present+valid -> `Ok(Of(n))`, present+invalid -> `Err(400)`. No mutable `var` declarations. Absent vs invalid is cleanly distinguished.

</td>
</tr>
</table>

### Conditional List Filtering

<table>
<tr><th>Conventional</th><th>fluentfp</th></tr>
<tr>
<td>

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

25 lines. Sort key buried in closure. Filter pattern repeated identically.

</td>
<td>

```go
orders := slice.SortBy(s.list(), orderNum)
if hasStatus {
  orders = orders.KeepIf(hasMatchingStatus)
}
if hasMinTotal {
  orders = orders.KeepIf(totalAtLeast)
}
```

`SortBy` takes a key function -- `orderNum` extracts the numeric suffix for correct ordering. `KeepIf` with named predicates replaces the `for`/`append` loop. The `if` guards are plain Go -- no special API needed for conditional filtering outside a chain.

</td>
</tr>
</table>

### Context Values

<table>
<tr><th>Conventional</th><th>fluentfp</th></tr>
<tr>
<td>

```go
type contextKey struct{}
var requestIDKey = contextKey{}

ctx := context.WithValue(
  r.Context(), requestIDKey, "req-1")

reqID, ok := req.Context().
  Value(requestIDKey).(string)
if !ok {
  reqID = "unknown"
}
```

Sentinel key type + type assertion + nil check.

</td>
<td>

```go
// With stores a value keyed by its Go type.
ctx := ctxval.With(r.Context(), RequestID("req-1"))

// From retrieves by type -> Option. Or provides fallback.
reqID := ctxval.From[RequestID](req.Context()).Or("unknown")
```

The Go type itself is the key -- no sentinel type to define. `From` returns an `Option`, so `.Or("unknown")` handles the absent case.

</td>
</tr>
</table>

### Response Construction

<table>
<tr><th>Conventional (each response)</th><th>fluentfp (each response)</th></tr>
<tr>
<td>

```go
w.Header().Set("Content-Type",
  "application/json")
w.WriteHeader(http.StatusCreated)
json.NewEncoder(w).Encode(enriched)
```

3 lines of mutation, repeated per code path. Miss `Content-Type` -> `text/plain`. `WriteHeader` after `Write` -> silently ignored.

</td>
<td>

```go
return rslt.Map(stored, web.Created[Order])
```

1 expression. Status + body travel as data. `Adapt` renders once.

</td>
</tr>
</table>

### Background Pipeline

<table>
<tr><th>Conventional</th><th>fluentfp</th></tr>
<tr>
<td>

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
    log.Printf("audit: %s", o.ID)
  }
}()

go func() {
  for o := range inventoryCh {
    log.Printf("inv: %d", len(o.Items))
  }
}()
```

3 channels, 3 goroutines, manual close ordering. No backpressure, no metrics, no error propagation.

</td>
<td>

```go
// FromChan: plain chan Order -> chan Result[Order] for toc.
// NewTee: broadcast each item to 2 branches.
tee := toc.NewTee(ctx, toc.FromChan(postCh), 2)
// Pipe: chain a processing function onto a branch.
auditPipe := toc.Pipe(
  ctx, tee.Branch(0), logOrder, toc.Options[Order]{})
inventoryPipe := toc.Pipe(
  ctx, tee.Branch(1), countItems, toc.Options[Order]{})
```

`toc.FromChan` bridges `chan Order` -> `chan rslt.Result[Order]` -- no passthrough stage needed. Backpressure, cancellation, shutdown ordering, and `Stats()` built in.

</td>
</tr>
</table>
