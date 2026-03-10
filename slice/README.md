# slice

Replace loop scaffolding with type-safe collection chains.

- **Interchangeable** — `Mapper[T]` is `[]T`. Index, `range`, `append`, `len` all work. Pass to or return from any function expecting `[]T`.
- **Generics** — 100% type-safe. No `any`, no reflection, no type assertions.
- **Method expressions** — pass `User.IsActive` directly. No wrapper closures.
- **Comma-ok** — `Find`, `IndexWhere` return `option` with `.Get()` → `(value, ok)`.

```go
// Before: 3 lines of scaffolding, 2 closing braces, 1 line of intent
var names []string
for _, u := range users {
    if u.IsActive() {
        names = append(names, u.Name)
    }
}

// After: intent only
names := slice.From(users).KeepIf(User.IsActive).ToString(User.GetName)
```

Six lines become one.

## What It Looks Like

```go
// Ranking
top5 := slice.SortByDesc(players, Player.Score).Take(5)
```

```go
// Tag filtering (allowlist semantics — empty filter matches all)
if !slice.String(m.Tags).Matches(filter.Tags) {
    continue
}
```

```go
// Cross-type mapping (both types inferred)
users := slice.Map(ids, FetchUser)
```

```go
// Multi-field extraction in one pass
prices, quantities, discounts, taxes := slice.Unzip4(orders,
    Order.Price, Order.Quantity, Order.Discount, Order.Tax,
)
```

`Order.Price` is a [method expression](https://go.dev/ref/spec#Method_expressions) — Go turns the method into a `func(Order) float64`, which is exactly what Unzip expects. This works when your types have accessor methods. Without them, named functions work in the same position:

```go
// With accessor methods
prices, qtys := slice.Unzip2(orders, Order.Price, Order.Quantity)

// Without — named functions instead
getPrice := func(o Order) float64 { return o.Price() }
getQty := func(o Order) int { return o.Qty() }
prices, qtys := slice.Unzip2(orders, getPrice, getQty)
```

Method expressions read as intent at the call site — `Order.Price` vs reading a function body. They pay off when multiple sites extract the same field. For one-off use, named functions avoid the extra method.

```go
// Set construction for O(1) lookup — works with any comparable type
allowed := slice.ToSet(cfg.AllowedRoles)
if allowed[user.Role] {
    grant(user)
}
```

```go
// Expand each department into its members — one-to-many, then flatten
allEmployees := slice.From(departments).FlatMap(Department.Employees)
```

```go
// Reduce to map
byMAC := slice.Fold(devices, make(map[string]Device), addDevice)
```

```go
// Number items starting from 1 — state is the counter, output is the labeled string
// number returns the next counter and a formatted label.
number := func(n int, item Item) (int, string) {
    return n + 1, fmt.Sprintf("%d. %s", n, item.Name)
}
_, numbered := slice.MapAccum(items, 1, number)
// ["1. Apples", "2. Bread", "3. Milk"]
```

## It's Just a Slice

`Mapper[T]` is `[]T`. Use it anywhere you'd use a slice:

```go
func activeNames(users []User) []string {
    names := slice.From(users).KeepIf(User.IsActive).ToString(User.Name)
    names.Each(lof.Println)
    return names  // return as []string — no conversion needed
}
```

```go
result := slice.From(users).KeepIf(User.IsActive)
fmt.Println(result[0])         // index
fmt.Println(len(result))       // len
result = append(result, extra) // append
for _, u := range result {     // range
    process(u)
}
```

- Keep `[]T` in your function signatures, not `Mapper[T]` — use `From()` at the point of use. This keeps fluentfp as an implementation detail; callers don't need to import it.
- `From()` is a zero-cost type conversion — no array copy. The Go spec [guarantees](https://go.dev/ref/spec#Conversions) that converting between types with identical underlying types only changes the type, not the representation. The 24-byte slice header (pointer, length, capacity) is shared; the backing array is the same. (`append` to either may mutate the other if capacity remains.)
- **Mutation boundaries:** Most operations (`KeepIf`, `Convert`, `ToString`, etc.) allocate fresh slices — the result is independent of the input. View operations (`From` alone, `Take`, `TakeLast`, `Chunk`) share the backing array. Use `.Clone()` at boundaries where shared state could be mutated. See [design.md § Boundaries and Defensive Copying](../docs/design.md#boundaries-and-defensive-copying) for practical guidance.
- Nil-safe: `From(nil).KeepIf(...).ToString(...)` returns an empty slice — Go's range over nil is zero iterations

Other Go FP libraries can't do this:
- **go-linq**, **fuego**: 6+ lines of `[]any` → `[]string` with type assertions to get results out
- **gofp**: conversion loops on both ends — `[]User` → `[]any` in, `[]any` → `[]string` out

See [comparison](../comparison.md) for the full library comparison.

## Operations

`From` creates `Mapper[T]`. For cross-type mapping, prefer the standalone `Map(ts, fn)` which infers all types and returns `Mapper[R]` for chaining. `MapTo[R]` creates `MapperTo[R,T]` for the narrow case where you filter before cross-type mapping: `MapTo[R](ts).KeepIf(pred).Map(fn)`. `String` (`[]string`), `Int` (`[]int`), and `Float64` (`[]float64`) are separate defined types with additional methods.

- **Filter**: `KeepIf`, `RemoveIf`, `Take`, `TakeLast`, `NonZero`
- **Search**: `Find`, `IndexWhere`, `FindAs`, `Any`, `Every`, `None`, `First`, `Single`, `Contains`, `ContainsAny`, `Matches` (String)
- **Transform**: `Convert`, `FlatMap`, `Map` (MapperTo), `Reverse`, `ToString`, `ToInt`, other `To*`, `Clone`, `Unique` (String), `UniqueBy`, `SortBy`, `SortByDesc`
- **Aggregate**: `Fold`, `MapAccum`, `Len`, `Max` (Int, Float64), `Min` (Int, Float64), `Sum` (Int, Float64), `ToSet`, `ToSetBy`, `Each`, `Unzip2`/`3`/`4`
- **Parallel**: `ParallelMap`, `ParallelKeepIf`, `ParallelEach` — concurrent versions for CPU/IO-bound transforms. Goroutine overhead makes these slower for trivial operations; only beneficial when `fn` does meaningful work per element. Run `go test -bench=BenchmarkParallel ./slice/` for numbers on your hardware.
- **Concurrent I/O**: `FanOut`, `FanOutEach` — bounded concurrent traversal with per-item scheduling, context-aware cancellation, and panic recovery. Returns `Mapper[result.Result[R]]` for chainability. See [result](../result/) for `CollectAll`/`CollectOk`.

`Fold`, not `Reduce`: `Fold` takes an initial value and allows the return type to differ from the element type (`func(R, T) R`). `Reduce` conventionally implies no initial value and same-type accumulation. The name matches the semantics.

## FanOut

`FanOut` runs a function on every element of a slice concurrently, limited to `n` at a time. Each element gets its own goroutine. Results come back in input order — `output[i]` corresponds to `input[i]`.

```go
results := slice.FanOut(ctx, 10, cities, City)
infos, err := result.CollectAll(results)
```

`FanOutEach` is the side-effect variant for operations that don't produce values — it returns `[]error` instead of `Mapper[result.Result[R]]`.

### How it works

1. A dispatch loop iterates the input slice, acquiring a semaphore slot before launching each goroutine.
2. Each goroutine calls `fn(ctx, item)`, stores the result, releases its semaphore slot, and exits.
3. After the loop, all goroutines are joined. No goroutine outlives `FanOut`.

The semaphore is a buffered channel of size `n`. When all slots are taken, the dispatch loop blocks until a running goroutine finishes and releases one.

### Cancellation

FanOut passes `ctx` to every callback. When `ctx` is cancelled:

- **Dispatch stops.** The loop checks `ctx` before acquiring each semaphore slot. Unscheduled items get `Err(ctx.Err())` without launching a goroutine.
- **In-flight callbacks continue** until `fn` returns. FanOut does not kill goroutines — it waits for them. If `fn` checks `ctx` (e.g., via `http.NewRequestWithContext`), it can exit early. If `fn` ignores `ctx`, it runs to completion.

FanOut returns only after every started goroutine has finished — no goroutine leaks.

### Fail-fast

By default, one item's error does not cancel siblings. If you want fail-fast — stop everything on the first error — cancel the context from within `fn`:

```go
ctx, cancel := context.WithCancel(parentCtx)
defer cancel()

// failFast wraps City to cancel siblings on first error.
failFast := func(callCtx context.Context, name string) (*Info, error) {
    info, err := City(callCtx, name)
    if err != nil {
        cancel()
        return nil, err
    }
    return info, nil
}

results := slice.FanOut(ctx, 10, cities, failFast)
```

This is opt-in because many I/O workloads want partial results — fetch what you can, report what failed. Use `result.CollectOk(results)` to gather successes and discard failures.

### Panic recovery

If `fn` panics, that item's result becomes `Err` wrapping a `*result.PanicError` (with the panic value and stack trace). Other items are unaffected. Detect panics with `errors.As`:

```go
var pe *result.PanicError
if errors.As(err, &pe) {
    log.Printf("panic: %v\n%s", pe.Value, pe.Stack)
}
```

### FanOut vs ParallelMap

| | FanOut | ParallelMap |
|---|---|---|
| **Scheduling** | Per-item (semaphore) | Batch chunking |
| **Context** | Passed to `fn` | None |
| **Errors** | Per-item `Result[R]` | None (panic crashes) |
| **Use case** | I/O-bound (HTTP, DB) | CPU-bound (parsing, hashing) |

The scheduling difference matters for I/O. Per-item scheduling means as soon as one HTTP call finishes, the next item starts — the semaphore slot is immediately reused. Batch chunking divides items into equal groups; if one item in a batch is slow, its worker sits occupied while faster workers in other batches finish and go idle. For CPU-bound work with uniform cost per item, batches avoid the per-item goroutine overhead.

Use `FanOut` when items have variable latency and you need error handling. Use `ParallelMap` when items have uniform cost and `fn` can't fail.

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/slice) for complete API documentation, the [main README](../README.md) for installation and performance characteristics, and the [showcase](../docs/showcase.md) for real-world rewrites.
