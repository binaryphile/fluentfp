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
top5 := slice.SortByDesc(players, Player.Score).TakeFirst(5)
```

```go
// Tag filtering (allowlist semantics — empty filter matches all)
if !slice.String(m.Tags).Matches(filter.Tags) {
    continue
}
```

```go
// Type mapping
users := slice.MapTo[User](ids).Map(FetchUser)
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
- `From()` is a type conversion (copies the slice header, shares the backing array — `append` to either may mutate the other if capacity remains)
- Nil-safe: `From(nil).KeepIf(...).ToString(...)` returns an empty slice — Go's range over nil is zero iterations

Other Go FP libraries can't do this:
- **go-linq**, **fuego**: 6+ lines of `[]any` → `[]string` with type assertions to get results out
- **gofp**: conversion loops on both ends — `[]User` → `[]any` in, `[]any` → `[]string` out

See [comparison](../comparison.md) for the full library comparison.

## Operations

`From` creates `Mapper[T]`. `MapTo[R]` creates `MapperTo[R,T]` — all Mapper methods plus `Map` for arbitrary type mapping. `String` (`[]string`), `Int` (`[]int`), and `Float64` (`[]float64`) are separate defined types with additional methods.

- **Filter**: `KeepIf`, `RemoveIf`, `TakeFirst`
- **Search**: `Find`, `IndexWhere`, `FindAs`, `Any`, `First`, `Single`, `Contains`, `ContainsAny`, `Matches` (String)
- **Transform**: `Convert`, `Map` (MapperTo), `ToString`, `ToInt`, other `To*`, `Clone`, `Unique` (String), `SortBy`, `SortByDesc`
- **Aggregate**: `Fold`, `MapAccum`, `Len`, `Max` (Int, Float64), `Min` (Int, Float64), `Sum` (Int, Float64), `ToSet`, `Each`, `Unzip2`/`3`/`4`
- **Parallel**: `ParallelMap`, `ParallelKeepIf`, `ParallelEach`

`Fold`, not `Reduce`: `Fold` takes an initial value and allows the return type to differ from the element type (`func(R, T) R`). `Reduce` conventionally implies no initial value and same-type accumulation. The name matches the semantics.

## Parallel Operations

Same API, concurrent execution. The pure-function contract (no shared state in predicates or transforms) makes parallelism safe by construction.

```go
// CPU-bound transform — 4x speedup at 10k elements
scores := slice.ParallelMap(users, runtime.GOMAXPROCS(0), ComputeScore)

// Parallel filter chains naturally into sequential operations
active := slice.From(users).ParallelKeepIf(4, User.IsActive).ToString(User.Name)
```

Parallel overhead only pays off when `fn` does meaningful work per element. Pure field access won't benefit — CPU-bound transforms and I/O-bound operations will. Benchmarks:

| Operation | Work | 10k elements | Speedup |
|-----------|------|-------------|---------|
| ParallelMap | trivial (n*2) | ~2x slower | — |
| ParallelMap | CPU-bound (50 sin/cos) | ~5x faster | yes |
| ParallelKeepIf | trivial (n%2) | ~4x slower | — |
| ParallelKeepIf | CPU-bound | ~4x faster | yes |

Run `go test -bench=BenchmarkParallel -benchmem ./slice/` for numbers on your hardware.

Edge cases: `workers <= 0` panics, `workers == 1` runs sequentially (no goroutine overhead), `workers > len` clamps, nil/empty returns nil.

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/slice) for complete API documentation and the [main README](../README.md) for installation and performance characteristics.
