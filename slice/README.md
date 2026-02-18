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
leadTimes, deployFreqs, mttrs, cfrs := slice.Unzip4(history,
    Record.LeadTime, Record.DeployFreq, Record.MTTR, Record.ChangeFailRate,
)
```

```go
// Reduce to map
byMAC := slice.Fold(devices, make(map[string]Device), addDevice)
```

## It's Just a Slice

`Mapper[T]` is `[]T`. Use it anywhere you'd use a slice:

```go
func activeNames(users []User) []string {
    names := slice.From(users).
        KeepIf(User.IsActive).
        ToString(User.Name)     // returns String ([]string)
    names.Each(lof.Println)
    return names                // return as []string
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

- `From()` is a type-cast, not a copy
- Nil-safe: `From(nil).KeepIf(...).ToString(...)` returns an empty slice — Go's range over nil is zero iterations

Other Go FP libraries can't do this:
- **go-linq**, **fuego**: 6+ lines of `[]any` → `[]string` with type assertions to get results out
- **gofp**: conversion loops on both ends — `[]User` → `[]any` in, `[]any` → `[]string` out

See [comparison](../comparison.md) for the full library comparison.

## Operations

`From` creates `Mapper[T]`. `MapTo[R]` creates `MapperTo[R,T]` — all Mapper methods plus `Map` for arbitrary type mapping. `String` (`[]string`) and `Float64` (`[]float64`) are separate defined types with additional methods.

- **Filter**: `KeepIf`, `RemoveIf`, `TakeFirst`
- **Search**: `Find`, `IndexWhere`, `FindAs`, `Any`, `First`, `Single`, `Contains`, `ContainsAny`, `Matches` (String)
- **Transform**: `Convert`, `Map` (MapperTo), `ToString`, `ToInt`, other `To*`, `Clone`, `Unique` (String), `SortBy`, `SortByDesc`
- **Aggregate**: `Fold`, `Len`, `Sum` (Float64), `ToSet` (String), `Each`, `Unzip2`/`3`/`4`
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
