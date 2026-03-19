# Conventional Go vs fluentfp — Complete Side by Side

This document shows every place the orders example uses fluentfp, paired with the conventional Go equivalent. Read the [README](README.md) first for the narrative walkthrough.

## The Complete POST Handler

<table>
<tr>
<th>Conventional Go</th>
<th>fluentfp</th>
</tr>
<tr>
<td>

```go
func handleCreateOrder(
  w http.ResponseWriter,
  req *http.Request,
) {
  // --- decode ---
  if req.Header.Get("Content-Type") !=
      "application/json" {
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
        "error": err.Error()})
    return
  }

  // --- validate ---
  if order.Customer == "" {
    w.Header().Set("Content-Type",
      "application/json")
    w.WriteHeader(400)
    json.NewEncoder(w).Encode(
      map[string]string{
        "error": "customer is required"})
    return
  }
  if len(order.Items) == 0 {
    w.Header().Set("Content-Type",
      "application/json")
    w.WriteHeader(400)
    json.NewEncoder(w).Encode(
      map[string]string{
        "error": "must have items"})
    return
  }

  // --- assign ID ---
  order.ID = fmt.Sprintf("ord-%d",
    idCounter.Add(1))
  order.Status = "pending"

  // --- enrich (with breaker) ---
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
    w.Header().Set("Content-Type",
      "application/json")
    w.WriteHeader(500)
    json.NewEncoder(w).Encode(
      map[string]string{
        "error": "internal error"})
    return
  }
  breaker.recordSuccess()

  // --- store + respond ---
  store.put(enriched)
  w.Header().Set("Content-Type",
    "application/json")
  w.WriteHeader(201)
  json.NewEncoder(w).Encode(enriched)
}
```

</td>
<td>

```go
handleCreateOrder := func(
  req *http.Request,
) rslt.Result[web.Response] {

  // --- decode + validate ---
  validated := rslt.FlatMap(
    web.DecodeJSON[Order](req),
    validateOrder,
  )
  order, err := validated.Unpack()
  if err != nil {
    return rslt.Err[web.Response](err)
  }










  // --- assign ID ---
  order.ID = fmt.Sprintf("ord-%d",
    idCounter.Add(1))
  order.Status = "pending"

  // --- enrich (breaker is invisible) ---
  enriched, err := enrichWithBreaker(
    req.Context(), order)
  if err != nil {
    return rslt.Err[web.Response](err)
  }








  // --- store + respond ---
  s.put(enriched)
  return rslt.Ok(web.Created(enriched))
}
```

</td>
</tr>
</table>

The blank lines on the right aren't padding — they show where the code *isn't*. Six response-writing blocks on the left, zero on the right. The `web.Adapt` wrapper (not shown) handles all response rendering once, outside the handler.

What disappeared:

- **Content-Type / MaxBytesReader / DisallowUnknownFields** — `web.DecodeJSON` does it all
- **Validation error blocks** — `rslt.FlatMap` chains decode into validation; errors propagate
- **Breaker state management** — `enrichWithBreaker` has the same signature as `enrichOrder`
- **Error-to-503 mapping** — defined once at the adapter boundary, not in the handler
- **Response mutation** — `web.Created(enriched)` replaces 3 lines of `Set`/`WriteHeader`/`Encode`

The conventional version also excludes the 40+ lines needed to implement `circuitBreaker` itself.

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
web.DecodeJSON[Order](req)
```

Returns `Result[Order]` — the decoded order or a structured error (415, 413, 400).

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

One monolithic function. Bare `error` return loses the HTTP status code.

</td>
<td>

```go
validateOrder := web.Steps(
  hasCustomer,
  hasItems,
  itemsHavePositiveQty,
)
```

A list of named functions. Each carries its own HTTP status code:

```go
func itemsHavePositiveQty(
  o Order,
) rslt.Result[Order] {
  if !slice.From(o.Items).
      Every(LineItem.HasPositiveQty) {
    return rslt.Err[Order](
      web.BadRequest("positive qty required"))
  }
  return rslt.Ok(o)
}
```

</td>
</tr>
</table>

### Chaining Decode into Validation

<table>
<tr><th>Conventional</th><th>fluentfp</th></tr>
<tr>
<td>

```go
if err := dec.Decode(&order); err != nil {
  // write 400 response...
  return
}
if err := validateOrder(order); err != nil {
  // write 400 response...
  return
}
```

Two error blocks. A third step means a third block.

</td>
<td>

```go
validated := rslt.FlatMap(
  web.DecodeJSON[Order](req),
  validateOrder,
)
order, err := validated.Unpack()
if err != nil {
  return rslt.Err[web.Response](err)
}
```

One chain. A third step is another `FlatMap`.

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
// Setup (5 lines):
breaker := hof.NewBreaker(hof.BreakerConfig{
  ResetTimeout: 10 * time.Second,
  ReadyToTrip:  hof.ConsecutiveFailures(3),
})
enrichWithBreaker := hof.WithBreaker(
  breaker, enrichOrder)

// In handler (1 line):
enriched, err := enrichWithBreaker(
  req.Context(), order)
```

Same signature as `enrichOrder`. The handler doesn't know a breaker exists. Probe gating, panic recovery, cancellation, and concurrent transitions are handled.

</td>
</tr>
</table>

### Error Mapping

<table>
<tr><th>Conventional</th><th>fluentfp</th></tr>
<tr>
<td>

```go
// Repeated in every handler:
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
```

</td>
<td>

```go
// Defined once:
func mapDomainError(
  err error,
) (*web.Error, bool) {
  if errors.Is(err, hof.ErrCircuitOpen) {
    return &web.Error{
      Status:  503,
      Message: "pricing unavailable",
    }, true
  }
  return nil, false
}

errorMapper := web.WithErrorMapper(
  mapDomainError)
mux.HandleFunc("POST /orders",
  web.Adapt(handleCreateOrder, errorMapper))
```

One mapping, applied at the boundary.

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
return rslt.Map(
  option.New(s.get(id)).
    OkOr(web.NotFound("order not found")),
  web.OK[Order],
)
```

`option.New` wraps comma-ok. `.OkOr` bridges to Result. `rslt.Map` wraps in a response. The 404 propagates through `Map` untouched.

</td>
</tr>
</table>

### Map Lookup with Fallback

<table>
<tr><th>Conventional</th><th>fluentfp</th></tr>
<tr>
<td>

```go
price, ok := prices[item.SKU]
if !ok {
  price = 100
}
```

</td>
<td>

```go
price := option.Lookup(prices, item.SKU).Or(100)
```

</td>
</tr>
</table>

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
status, hasStatus :=
  option.NonEmpty(q.Get("status")).Get()
minTotalOption := option.FlatMap(
  option.NonEmpty(q.Get("min_total")),
  option.Atoi,
)
mt, hasMinTotal := minTotalOption.Get()
```

Parse pipeline: empty → skip, non-empty → parse. No mutable variables.

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
orders := slice.SortBy(s.list(), Order.GetID).
  KeepIfWhen(hasStatus, hasMatchingStatus).
  KeepIfWhen(hasMinTotal, totalAtLeast)
```

3 lines. Method expression for sort key. `KeepIfWhen` skips the filter when condition is false.

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
ctx := ctxval.With(
  r.Context(), RequestID("req-1"))

reqID := ctxval.From[RequestID](
  req.Context()).Or("unknown")
```

Go type is the key. `Option` fallback.

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

3 lines of mutation, repeated per code path. Miss `Content-Type` → `text/plain`. `WriteHeader` after `Write` → silently ignored.

</td>
<td>

```go
return rslt.Ok(web.Created(enriched))
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
    log.Printf("inventory: %d", len(o.Items))
  }
}()
```

3 channels, 3 goroutines, manual close ordering. No backpressure, no metrics, no error propagation.

</td>
<td>

```go
stage := toc.Start(ctx, passthrough,
  toc.Options[Order]{
    Capacity: 10, Workers: 1})
tee := toc.NewTee(ctx, stage.Out(), 2)
auditPipe := toc.Pipe(ctx,
  tee.Branch(0), logOrder,
  toc.Options[Order]{})
inventoryPipe := toc.Pipe(ctx,
  tee.Branch(1), countItems,
  toc.Options[Order]{})
```

Backpressure, cancellation, shutdown ordering, and `Stats()` built in.

</td>
</tr>
</table>
