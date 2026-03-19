# result

Typed results for operations that may fail.

`Result[R]` stores a value or an error. The zero value is `Ok` with the zero value of `R` — an uninitialized `Result` (e.g. a struct field never assigned) silently reports success. Always construct explicitly via `Ok`, `Err`, or `Of`.

```go
// Wrap a (T, error) return pair
res := rslt.Of(strconv.Atoi(input))
port := res.Or(8080)
```

Pairs with `FanOut` for per-item error and panic handling in concurrent workloads, and with `FanOutAll` for all-or-nothing behavior:

```go
// Per-item outcomes — collect what you need
results := slice.FanOut(ctx, 8, urls, fetchURL)
pages, errs := rslt.CollectOkAndErr(results)
```

```go
// All-or-nothing — first error cancels remaining work
pages, err := slice.FanOutAll(ctx, 8, urls, fetchURL)
```

## What It Looks Like

```go
// Construct
ok := rslt.Ok(42)
fail := rslt.Err[int](errors.New("not found"))

// From (T, error) pair
res := rslt.Of(strconv.Atoi("42"))
```

```go
// Comma-ok extraction
if v, ok := res.Get(); ok {
    fmt.Println(v)
}
```

```go
// Default value
port := res.Or(8080)
```

```go
// Cross-type transform (standalone — Go methods can't introduce type params)
name := rslt.Map(userResult, User.Name)
```

```go
// Dispatch by state
msg := rslt.Fold(res,
    func(err error) string { return "failed: " + err.Error() },
    func(v int) string { return fmt.Sprintf("got %d", v) },
)
```

```go
// Detect panics from FanOut — panics are recovered per item
results := slice.FanOut(ctx, 8, urls, fetchURL)
for _, res := range results {
    if err, ok := res.GetErr(); ok {
        var pe *rslt.PanicError
        if errors.As(err, &pe) {
            log.Printf("panic: %v\nstack:\n%s", pe.Value, pe.Stack)
        }
    }
}
```

## PanicError

`FanOut` recovers panics per item and wraps them as `*PanicError`. This type carries the original panic value and a stack trace. Detect it with `errors.As`:

```go
var pe *rslt.PanicError
if errors.As(err, &pe) {
    log.Printf("panic: %v\nstack:\n%s", pe.Value, pe.Stack)
}
```

If the panic value was an `error`, `PanicError.Unwrap()` returns it — enabling `errors.Is` chains through the original error.

## Operations

**Create**: `Ok`, `Err`, `Of` (from `(T, error)` pair)

**Extract**: `Get`, `Err`, `Unpack` (to `(R, error)` pair), `Or`, `OrCall`, `GetErr`, `IsOk`, `IsErr`, `MustGet`

**Transform**: `Convert` (same type), `FlatMap` (method + standalone), `MapErr` (transform error), `Map` (cross-type, standalone), `Fold` (standalone), `Lift` (wrap `func(A)(R, error)` → `func(A) Result[R]`)

**Side effects**: `IfOk`, `IfErr`, `Tap` (chainable Ok side effect), `TapErr` (chainable Err side effect)

**Collect**: `CollectAll` (all values or first error by index), `CollectOk` (successes only), `CollectErr` (errors only), `CollectOkAndErr` (both in one pass)

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/rslt) for complete API documentation, the [main README](../README.md) for installation, and the [showcase](../docs/showcase.md) for real-world comparisons.
