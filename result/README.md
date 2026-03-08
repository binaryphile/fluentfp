# result

Typed results for operations that may fail. Pairs with `FanOut` for per-item error and panic handling in concurrent workloads.

```go
// Concurrent fetch with per-item error handling
results := slice.FanOut(ctx, 8, urls, fetchURL)

// Fail-fast: all values if every item succeeded, or first error
values, err := result.CollectAll(results)

// Lenient: gather successes, skip failures
values := result.CollectOk(results)
```

## What It Looks Like

```go
// Construct
ok := result.Ok(42)
fail := result.Err[int](errors.New("not found"))

// From (T, error) — wraps any Go function's return
res := result.Of(strconv.Atoi("42"))
```

```go
// Comma-ok extraction
if v, ok := res.Get(); ok {
    fmt.Println(v)
}
```

```go
// Default value
port := res.GetOr(8080)
```

```go
// Cross-type transform (standalone — Go methods can't introduce type params)
name := result.Map(userResult, User.Name)
```

```go
// Dispatch by state
msg := result.Fold(res,
    func(err error) string { return "failed: " + err.Error() },
    func(v int) string { return fmt.Sprintf("got %d", v) },
)
```

```go
// Detect panics from FanOut
if err, ok := res.GetErr(); ok {
    var pe *result.PanicError
    if errors.As(err, &pe) {
        log.Printf("panic: %v\nstack:\n%s", pe.Value, pe.Stack)
    }
}
```

## PanicError

`FanOut` recovers panics per item and wraps them as `*PanicError`. This type carries the original panic value and a stack trace. Detect it with `errors.As`:

```go
var pe *result.PanicError
if errors.As(err, &pe) {
    log.Printf("panic: %v\nstack:\n%s", pe.Value, pe.Stack)
}
```

If the panic value was an `error`, `PanicError.Unwrap()` returns it — enabling `errors.Is` chains through the original error.

## Operations

**Create**: `Ok`, `Err`, `Of` (from `(T, error)` pair)

**Extract**: `Get`, `GetOr`, `GetErr`, `IsOk`, `IsErr`, `MustGet`

**Transform**: `Convert` (same type), `Map` (cross-type, standalone), `Fold` (standalone)

**Side effects**: `IfOk`, `IfErr`

**Collect**: `CollectAll` (fail-fast), `CollectOk` (successes only)

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/result) for complete API documentation, the [main README](../README.md) for installation, and the [showcase](../docs/showcase.md) for real-world comparisons.
