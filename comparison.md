# Library Comparison

Compare fluentfp to popular Go FP libraries. Task: filter active users, extract names.

## Quick Comparison

| Library | Stars | Type-Safe | Concise | Method Exprs | Fluent |
|---------|-------|-----------|---------|--------------|--------|
| fluentfp | — | ✅ | ✅ | ✅ | ✅ |
| samber/lo | 17k | ✅ | ❌ | ❌ | ❌ |
| thoas/go-funk | 4k | ❌ | ✅ | ✅ | ❌ |
| ahmetb/go-linq | 3k | ❌ | ❌ | ❌ | ✅ |
| rjNemo/underscore | — | ✅ | ✅ | ✅ | ❌ |

## Criteria Explained

**Type-Safe:** Uses Go generics. No `any` or type assertions.

**Concise:** Callbacks don't require unused parameters (like index).

**Method Expressions:** Can pass `User.IsActive` directly without wrapper.

**Fluent:** Supports method chaining: `slice.KeepIf(...).ToString(...)`

## Code Comparison

Each example shows idiomatic usage for that library. fluentfp uses method expressions (`User.IsActive`) directly. Other libraries require wrapper functions.

### fluentfp (4 lines)

```go
names := slice.From(users).
    KeepIf(User.IsActive).
    ToString(User.Name)
names.Each(lof.Println)
```

Method expressions, fluent chaining, `Each` for iteration.

### samber/lo (7 lines)

```go
userIsActive := func(u User, _ int) bool { return u.IsActive() }
getName := func(u User, _ int) string { return u.Name() }
actives := lo.Filter(users, userIsActive)
names := lo.Map(actives, getName)
for _, name := range names {
    fmt.Println(name)
}
```

Requires index wrappers (`_ int`), no fluent chaining, manual iteration.

### thoas/go-funk (5 lines)

```go
actives := funk.Filter(users, User.IsActive).([]User)
names := funk.Map(actives, User.Name).([]string)
for _, name := range names {
    fmt.Println(name)
}
```

Supports method expressions, but requires type assertions and manual iteration.

### ahmetb/go-linq (7 lines)

```go
userIsActive := func(user any) bool { return user.(User).IsActive() }
name := func(user any) any { return user.(User).Name() }
var names []any
linq.From(users).Where(userIsActive).Select(name).ToSlice(&names)
for _, n := range names {
    fmt.Println(n)
}
```

Fluent chaining, but requires `any` wrappers and type assertions.

### rjNemo/underscore (5 lines)

```go
actives := u.Filter(users, User.IsActive)
names := u.Map(actives, User.Name)
for _, name := range names {
    fmt.Println(name)
}
```

Type-safe, supports method expressions, but no fluent chaining or `Each`.

## Beyond Filter+Map

The filter+map comparison above shows ergonomic differences. The operations below show **structural** differences — how each library handles search results, accumulation, and multi-step chains.

### Find and Reduce — fluentfp vs lo

**Find — fluentfp:**
```go
user := slice.From(users).Find(User.IsActive).Or(defaultUser)
```

**Find — lo:**
```go
user, ok := lo.Find(users, func(u User, _ int) bool { return u.IsActive() })
if !ok {
    user = defaultUser
}
```

fluentfp's `Find` returns `option.Basic[T]`, so `.Or()` handles the missing case inline. lo returns `(T, bool)` — idiomatic Go, but requires a separate `if` block for the default.

**Reduce — fluentfp:**
```go
total := slice.Fold(orders, 0, func(sum int, o Order) int { return sum + o.Amount() })
```

**Reduce — lo:**
```go
total := lo.Reduce(orders, func(sum int, o Order, _ int) int { return sum + o.Amount() }, 0)
```

Both work. fluentfp's accumulator takes `func(R, T) R` — no index parameter. lo's takes `func(R, T, int) R` — the index is available when you need it, but most reductions don't.

### Multi-Step Chains — fluentfp vs go-linq

Only two libraries support chaining. Both handle a filter → transform → deduplicate pipeline differently.

**fluentfp:**
```go
names := slice.From(users).
    KeepIf(User.IsActive).
    ToString(User.Name).
    Unique()
```

**go-linq:**
```go
userIsActive := func(user any) bool { return user.(User).IsActive() }
getName := func(user any) any { return user.(User).Name() }
var results []any
linq.From(users).
    Where(userIsActive).
    Select(getName).
    Distinct().
    ToSlice(&results)
names := make([]string, len(results))
for i, r := range results {
    names[i] = r.(string)
}
```

fluentfp chains are type-safe throughout — no `any` wrappers, no type assertions, and the result assigns directly to `[]string`. go-linq's strength is lazy evaluation: intermediate slices aren't materialized, which matters when collections are large. fluentfp evaluates eagerly — each step allocates.

### Multi-Field Extraction — fluentfp (unique)

No other library has `Unzip`. When you need multiple fields from the same slice, fluentfp extracts them in one pass.

**fluentfp:**
```go
names, ages, scores := slice.Unzip3(students, Student.Name, Student.Age, Student.Score)
```

**lo (3 separate passes):**
```go
names := lo.Map(students, func(s Student, _ int) string { return s.Name() })
ages := lo.Map(students, func(s Student, _ int) int { return s.Age() })
scores := lo.Map(students, func(s Student, _ int) float64 { return s.Score() })
```

Niche operation — most codebases won't need it. But when you do, one call with method expressions replaces three calls with wrapper closures. `Unzip3` also traverses the input once instead of three times.

## Performance

### Benchmark Results

Library comparison benchmarks (1000 users, 50% active, Intel i5-1135G7). Loop baselines use pre-allocated slices per [methodology.md](methodology.md#i-performance-analysis).

```
Filter + Map (pre-allocated loop baseline):
BenchmarkLoop-8           8795 ns/op   32768 B/op    2 allocs/op
BenchmarkFluentfp-8       8095 ns/op   32768 B/op    2 allocs/op
BenchmarkLo-8             8125 ns/op   32768 B/op    2 allocs/op
BenchmarkUnderscore-8    11784 ns/op   40296 B/op   11 allocs/op
BenchmarkGoLinq-8        86005 ns/op   51288 B/op 1529 allocs/op
BenchmarkGoFunk-8       495920 ns/op  120000 B/op 4024 allocs/op

Filter only (pre-allocated):
BenchmarkLoop_FilterOnly-8         6743 ns/op   24576 B/op    1 allocs/op
BenchmarkFluentfp_FilterOnly-8     6060 ns/op   24576 B/op    1 allocs/op
BenchmarkLo_FilterOnly-8           6243 ns/op   24576 B/op    1 allocs/op
BenchmarkUnderscore_FilterOnly-8   9737 ns/op   32104 B/op   10 allocs/op
BenchmarkGoFunk_FilterOnly-8     334949 ns/op   69152 B/op 2512 allocs/op

Map only (pre-allocated):
BenchmarkLoop_MapOnly-8         4581 ns/op   16384 B/op    1 allocs/op
BenchmarkFluentfp_MapOnly-8     5172 ns/op   16384 B/op    1 allocs/op
BenchmarkLo_MapOnly-8           4416 ns/op   16384 B/op    1 allocs/op
BenchmarkUnderscore_MapOnly-8   5292 ns/op   16384 B/op    1 allocs/op
BenchmarkGoFunk_MapOnly-8     384516 ns/op   99232 B/op 3013 allocs/op
```

### Key Findings

1. **Generic libraries equal pre-allocated loops:** fluentfp, lo, and loops all show ~8μs for filter+map with identical allocations (2). Differences are measurement noise.

2. **Reflection libraries are orders of magnitude slower:** go-funk (56×) and go-linq (10×) use reflection for type handling.

3. **underscore has extra allocations:** 11 allocs vs 2 for other generic libraries, causing ~1.3× overhead.

### Guidance

- **Generic libraries match loops:** fluentfp and lo perform equivalently to properly-written pre-allocated loops
- **Naive loops are slower:** Loops using `var s []T` + `append` (without pre-allocation) perform worse than generic libraries due to repeated slice growth
- **Avoid reflection libraries in hot paths:** go-funk and go-linq have measurable overhead
- **Profile first:** Differences between generic libraries are noise

Source: `examples/comparison/benchmark_test.go`. Run with `cd examples/comparison && go test -bench=. -benchmem`

See [methodology.md § Performance Analysis](methodology.md#i-performance-analysis) for detailed fluentfp vs loop analysis on different operation types.

## When to Use a Different Library

| Use Case | Library | Why |
|----------|---------|-----|
| Broad adoption | samber/lo | 17k+ stars, wide community support |
| Lazy evaluation | go-linq | Deferred execution avoids materializing large intermediates |
| Pre-generics codebase | go-funk | Works without Go 1.18+ |
| Already using one | — | Consistency matters more than marginal gains |

## Recommendation

Use fluentfp when you need all four criteria. Use lo if you need the most popular/maintained option and don't mind wrapper functions.

See [examples/comparison/main.go](examples/comparison/main.go) for full executable comparison with additional libraries.

*Star counts approximate as of 2025.*
