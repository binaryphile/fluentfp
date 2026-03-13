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

- Keep `[]T` for your slice arguments in function signatures, not `Mapper[T]` — use `From()` at the point of use. This keeps fluentfp as an implementation detail; callers don't need to import it.  It can be useful to return `Mapper[T]` from many functions, however, knowing that it's still usable by the consumer as a regular slice.
- `From()` is a zero-cost type conversion — no array copy. The Go spec [guarantees](https://go.dev/ref/spec#Conversions) that converting between types with identical underlying types only changes the type, not the representation. The 24-byte slice header (pointer, length, capacity) is shared; the backing array is the same. (`append` to either may mutate the other if capacity remains.)
- **Mutation boundaries:** Most operations (`KeepIf`, `Convert`, `ToString`, etc.) allocate fresh slices — the result is independent of the input. View operations (`From` alone, `Take`, `TakeLast`, `TakeWhile`, `Drop`, `DropLast`, `DropWhile`, `Chunk`) share the backing array. Use `.Clone()` at boundaries where shared state could be mutated. See [design.md § Boundaries and Defensive Copying](../docs/design.md#boundaries-and-defensive-copying) for practical guidance.
- Nil-safe: `From(nil).KeepIf(...).ToString(...)` returns an empty slice — Go's range over nil is zero iterations

Other Go FP libraries can't do this:
- **go-linq**, **fuego**: 6+ lines of `[]any` → `[]string` with type assertions to get results out
- **gofp**: conversion loops on both ends — `[]User` → `[]any` in, `[]any` → `[]string` out

See [comparison](../comparison.md) for the full library comparison.

## Operations

`From` creates `Mapper[T]`. For cross-type mapping, prefer the standalone `Map(ts, fn)` which infers all types and returns `Mapper[R]` for chaining. `MapTo[R]` creates `MapperTo[R,T]` for the narrow case where you filter before cross-type mapping: `MapTo[R](ts).KeepIf(pred).Map(fn)`. `String` (`[]string`), `Int` (`[]int`), and `Float64` (`[]float64`) are separate defined types with additional methods.

- **Filter**: `KeepIf`, `RemoveIf`, `Take`, `TakeLast`, `TakeWhile`, `Drop`, `DropLast`, `DropWhile`, `NonZero`
- **Search**: `Find`, `IndexWhere`, `FindAs`, `Any`, `Every`, `None`, `First`, `Single`, `Contains`, `ContainsAny`, `Matches` (`String`)
- **Transform**: `Convert`, `FlatMap`, `Map` (`MapperTo`), `Reverse`, `Intersperse`, `ToString`, `ToInt`, other `To*`, `Clone`, `Unique` (`String`), `UniqueBy`, `SortBy`, `SortByDesc`
- **Combine**: `Zip`, `ZipWith`
- **Aggregate**: `Fold`, `Scan`, `MapAccum`, `Len`, `Max` (`int`, `float64`), `Min` (`int`, `float64`), `Sum` (`int`, `float64`), `ToSet`, `ToSetBy`, `Each`, `Unzip2`/`3`/`4`, `GroupBy`, `Tally`
- **Generate**: `Range`, `RangeFrom`, `RangeStep` (return `Int` for numeric chaining), `RepeatN`
- **View**: `Chunk`, `Window` (sliding windows sharing backing array — overlapping windows alias the same memory; mutating one affects adjacent windows; use `.Clone()` for independent copies)
- **Parallel (no error return)**: `PMap`, `PKeepIf`, `PEach` — bounded concurrent operations for callbacks that do not return errors. Panics in fn are recovered, converted to `*result.PanicError` with a stack captured during recovery, and re-panicked on the calling goroutine after all workers exit. If multiple workers panic, one arbitrary panic is re-thrown; others are suppressed. Usually only worth using when per-item workload is large enough to amortize the overhead caused by creation and scheduling of goroutines.
- **Parallel (error-aware)**: `FanOut`, `FanOutAll`, `FanOutEach` — bounded concurrency for callbacks that take `context.Context` and return errors. Use `FanOut` for value-producing operations where partial success is acceptable, `FanOutAll` for all-or-nothing operations with early cancellation, and `FanOutEach` for side-effecting callbacks that return only `error`. If item costs vary widely, use the corresponding weighted variant (`FanOutWeighted`, `FanOutWeightedAll`, `FanOutEachWeighted`). See [result](../result/) for `CollectAll`, `CollectOk`, and `CollectOkAndErr`.

`Fold`, not `Reduce`: `Fold` takes an initial value and allows the return type to differ from the element type (`func(R, T) R`). `Reduce` conventionally implies no initial value and same-type accumulation. The name matches the semantics.

`Scan` is `Fold` that collects all intermediate accumulator values. It includes the initial value as the first element (Haskell `scanl` semantics), so `Scan(ts, z, f)` returns `len(ts)+1` elements. Law: `last(Scan(ts, z, f)) == Fold(ts, z, f)`.

`Zip` and `ZipWith` truncate to the shorter input — safe for pipelines where lengths are data-dependent. This differs from `pair.Zip`, which panics on length mismatch. Use `pair.Zip` when equal lengths are a structural invariant you want enforced.

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

### All-or-nothing: FanOutAll

When every item must succeed or the whole operation fails, use `FanOutAll`. On the first error or panic, it cancels a derived context — this skips unstarted work and lets in-flight callbacks stop early if they honor `ctx`, but it still waits for started callbacks to return. Returns `([]R, error)` directly:

```go
infos, err := slice.FanOutAll(ctx, 10, cities, City)
```

`FanOutAll` derives a child context internally — the caller's context is never cancelled. `FanOutWeightedAll` provides the same semantics with a cost budget instead of a fixed concurrency limit.

### Fail-fast vs partial results

`FanOut` does not cancel siblings on failure — every item runs to completion. This is intentional: many workloads want partial results. Use `result.CollectOk(results)` to gather successes and discard failures, or `result.CollectOkAndErr(results)` to get both halves.

If any failure should fail the whole batch, use `FanOutAll` instead — on the first error or panic it cancels a derived context, skips unstarted work, lets in-flight callbacks stop early if they honor `ctx`, and still waits for started callbacks to return.

### Panic recovery

If `fn` panics, that item's result becomes `Err` wrapping a `*result.PanicError` (with the panic value and stack trace). Other items are unaffected. Detect panics with `errors.As`:

```go
var pe *result.PanicError
if errors.As(err, &pe) {
    log.Printf("panic: %v\n%s", pe.Value, pe.Stack)
}
```

### Choosing a parallel operation

Choose in two steps:

1. **What callback signature do you have?**
   - `func(T) R`, `func(T) bool`, or `func(T)`: `PMap` / `PKeepIf` / `PEach`
   - `func(context.Context, T) (R, error)` or `func(context.Context, T) error`: the `FanOut*` family

2. **If you're using `FanOut*`, what result/error contract do you want?**
   - Return values and keep per-item failures: `FanOut` / `FanOutWeighted` + collector
   - Return values and fail the whole batch on first error: `FanOutAll` / `FanOutWeightedAll`
   - Side effects only: `FanOutEach` / `FanOutEachWeighted`

| Callback signature | Result/error contract | Operation |
|---|---|---|
| `func(T) R` / `func(T) bool` / `func(T)` | — | `PMap` / `PKeepIf` / `PEach` |
| `func(context.Context, T) (R, error)` | All must succeed | `FanOutAll` / `FanOutWeightedAll` |
| `func(context.Context, T) (R, error)` | Partial success OK | `FanOut` / `FanOutWeighted` + collector |
| `func(context.Context, T) error` | Side effects only | `FanOutEach` / `FanOutEachWeighted` |

**No error return** — use `PMap`, `PKeepIf`, or `PEach` when your callback does not return an error. Panics in fn are recovered, converted to `*result.PanicError` with a stack captured during recovery, and re-panicked on the calling goroutine after all workers exit. If multiple workers panic, one arbitrary panic is re-thrown; others are suppressed. Remaining workers continue until fn returns. If fn may block indefinitely, use `FanOut` or `FanOutAll` instead — they accept `context.Context` for timeout and cancellation. This is a good fit for transforms and filters where failure is not part of the callback's contract:

```go
// Normalize 10K strings
normalized := slice.PMap(inputs, 8, strings.ToLower)

// Compute SHA-256 digests concurrently
digests := slice.PMap(blobs, 4, sha256.Sum256)
```

**All-or-nothing** — every item must succeed or the whole batch fails. On the first error or panic, `FanOutAll` cancels a derived context — this skips unstarted work and lets in-flight callbacks stop early if they honor `ctx`, but it still waits for started callbacks to return:

```go
// Fetch all prices before computing portfolio (partial is useless)
prices, err := slice.FanOutAll(ctx, 5, tickers, fetchPrice)

// Parse and validate all manifests before deployment
manifests, err := slice.FanOutAll(ctx, 4, files, parseAndValidateManifest)
```

**Partial success** — gather what you can, log failures separately. Use `FanOut` with a collector:

```go
// Download avatars (missing ones OK)
avatarResults := slice.FanOut(ctx, 10, users, downloadAvatar)
avatars, errs := result.CollectOkAndErr(avatarResults)
```

```go
// Decode untrusted images (some may be corrupt)
imageResults := slice.FanOut(ctx, 8, blobs, decodeImage)
images := result.CollectOk(imageResults)
```

**Side effects only** — `FanOutEach` returns `[]error` instead of values:

```go
// Send notifications (log failures, keep going)
errs := slice.FanOutEach(ctx, 20, users, sendNotification)
```

**Scheduling model** is the next question after callback signature and error contract. `PMap` uses a fixed set of workers processing chunks of the input, which has lower scheduling overhead for many small, uniform tasks. `FanOut` schedules work per item under a concurrency limit, which handles variable-latency work better and enables cancellation and per-item error handling. If you've chosen the `FanOut` family and item costs vary widely, use the corresponding weighted variant to bound total in-flight cost instead of item count:

```go
// Upload files bounded by total in-flight MB
sizeMB := func(f File) int { return f.SizeMB }
uploads, err := slice.FanOutWeightedAll(ctx, 100, files, sizeMB, uploadFile)
```

For heavily skewed no-error workloads, benchmark `PMap`; if chunking becomes a bottleneck, you may need a custom worker pool.

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/slice) for complete API documentation, the [main README](../README.md) for installation and performance characteristics, and the [showcase](../docs/showcase.md) for real-world rewrites.
