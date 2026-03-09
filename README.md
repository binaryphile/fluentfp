# fluentfp

**Fluent functional programming for Go.**

The thinnest abstraction that eliminates mechanical bugs from Go. Chain type-safe operations on slices, options, and sum types — no loop scaffolding, no intermediate variables, no reflection.

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp) for API docs and the **[showcase](docs/showcase.md)** for 16 before/after rewrites from real GitHub projects.

## Quick Start

Requires Go 1.21+.

```bash
go get github.com/binaryphile/fluentfp
```

```go
// Before: 3 lines of scaffolding, 2 closing braces, 1 line of intent
var names []string                         // state
for _, u := range users {                  // iteration
    if u.IsActive() {                      // predicate
        names = append(names, u.Name)      // accumulation
    }
}

// After: intent only
names := slice.From(users).KeepIf(User.IsActive).ToString(User.Name)
```

That's a **fluent chain** — each step returns a value you can call the next method on, so the whole pipeline reads as a single expression: filter, then transform.

### Interchangeable Types

`Mapper[T]` is defined as `type Mapper[T any] []T` — a [defined type](https://go.dev/ref/spec#Type_definitions), not a wrapper. `[]T` and `Mapper[T]` convert implicitly in either direction, so you choose how much to expose:

```go
// Public API — hide the dependency. Callers never see fluentfp types.
func ActiveNames(users []User) []string {
    return slice.From(users).KeepIf(User.IsActive).ToString(User.Name)
}

// Internal — embrace it. Accepting Mapper saves From() calls across a chain of helpers.
func transform(users slice.Mapper[User]) slice.Mapper[User] {
    return users.KeepIf(User.IsActive).Convert(User.Normalize)
}
```

The public pattern keeps fluentfp as an implementation detail — callers don't import it, and its types don't appear in intellisense. Internal code can pass Mappers between helpers to avoid repeated `From()` wrapping.

### Method Expressions

Go lets you reference a method by its type name, creating a function value where the receiver becomes the first argument:

```go
func (u User) IsActive() bool  // method
func(User) bool                // method expression: User.IsActive
```

`KeepIf` expects `func(T) bool` — `User.IsActive` is exactly that:

```go
names := slice.From(users).KeepIf(User.IsActive).ToString(User.Name)
```

Without method expressions, every predicate needs a wrapper: `func(u User) bool { return u.IsActive() }`.

For `[]*User` slices, the method expression is `(*User).IsActive`.

See [naming patterns](naming-in-hof.md) for when to use method expressions vs named functions vs closures.

## What It Looks Like

Three highlights from the **[showcase](docs/showcase.md)**.

### Conditional Struct Fields

Go struct literals let you build and return a value in one statement — fluentfp keeps it that way when fields are conditional.

<table>
<tr><th>Before</th><th>After</th></tr>
<tr><td><pre><code class="language-go">var level string
if overdue {
    level = "critical"
} else {
    level = "info"
}
var icon string
if overdue {
    icon = "!"
} else {
    icon = "✓"
}
return Alert{
    Message: msg,
    Level:   level,
    Icon:    icon,
}
</code></pre></td><td><pre><code class="language-go">return Alert{
    Message: msg,
    Level:   value.Of("critical").When(overdue).Or("info"),
    Icon:    value.Of("!").When(overdue).Or("✓"),
}
</code></pre></td></tr>
</table>

Go has no inline conditional expression. `value.Of` fills that gap — each field resolves in place, so the struct literal stays a single statement. *From [hashicorp/consul](https://github.com/hashicorp/consul/blob/554b4ba24f86/agent/agent.go#L2482-L2530).*

### Sort, Trim, and Map-to-Slice

A network monitor's "top N processes by metric" function: convert a map to a slice, sort by a mode-dependent key, take the first N.

<table>
<tr><th>Before (18 lines)</th><th>After (2-line body)</th></tr>
<tr><td><pre><code class="language-go">func (s *Snapshot) TopNProcesses(n int, mode ViewMode) []ProcessesResult {
    var items []ProcessesResult
    for k, v := range s.Processes {
        items = append(items, NewResult(k, v))
    }
    switch mode {
    case ModeTableBytes:
        sort.Slice(items, func(i, j int) bool {
            return items[i].TotalBytes() &gt; items[j].TotalBytes()
        })
    case ModeTablePackets:
        sort.Slice(items, func(i, j int) bool {
            return items[i].TotalPackets() &gt; items[j].TotalPackets()
        })
    }
    if len(items) &lt; n {
        n = len(items)
    }
    return items[:n]
}
</code></pre></td><td><pre><code class="language-go">var sortKey = map[ViewMode]func(ProcessesResult) int{
    ModeTableBytes:   ProcessesResult.TotalBytes,
    ModeTablePackets: ProcessesResult.TotalPackets,
}
func (s *Snapshot) TopNProcesses(n int, mode ViewMode) []ProcessesResult {
    desc := slice.Desc(sortKey[mode])
    return kv.Map(s.Processes, NewResult).Sort(desc).Take(n)
}
</code></pre></td></tr>
</table>

`kv.Map` replaces the map-to-slice loop. Two duplicated `sort.Slice` closures with `func(i, j int) bool` become `.Sort(desc)` — method expressions in a map replace the switch. `.Take(n)` replaces the four-line bounds check. No index variables means no `items[i]`-vs-`items[j]` misreference. *From [chenjiandongx/sniffer](https://github.com/chenjiandongx/sniffer/blob/master/stat.go#L72-L93).*

### Bounded Concurrent Requests

Fetch weather for a list of cities with at most 10 simultaneous goroutines.

<table>
<tr><th>Before — errgroup (21 lines)</th><th>After (2 lines)</th></tr>
<tr><td><pre><code class="language-go">func Cities(ctx context.Context, cities ...string) ([]*Info, error) {
    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(10)
    res := make([]*Info, len(cities))
    for i, city := range cities {
        g.Go(func() error {
            info, err := City(ctx, city)
            if err != nil {
                return err
            }
            res[i] = info
            return nil
        })
    }
    if err := g.Wait(); err != nil {
        return nil, err
    }
    return res, nil
}
</code></pre></td><td><pre><code class="language-go">func Cities(ctx context.Context, cities ...string) ([]*Info, error) {
    results := slice.FanOut(ctx, 10, cities, City)
    return result.CollectAll(results)
}
</code></pre></td></tr>
</table>

`FanOut` replaces the goroutine-launching loop, closure captures, result-slot bookkeeping, and error aggregation. `City` passes directly — no wrapper needed. Unlike errgroup, FanOut recovers panics per item (as `*result.PanicError` with stack trace) and preserves every item's outcome. *From the [errgroup pattern](https://encore.dev/blog/advanced-go-concurrency).*

## Why fluentfp

**Type-safe end-to-end.** go-linq gives you `[]any` back — cast it and hope you got the type right. lo requires `func(T, int)` callbacks, so every stdlib function needs a wrapper to discard the unused index. fluentfp uses generics throughout: `Mapper[T]` is `[]T` with methods. If it compiles, the types are right.

**Bugs you can't write.** You can't get an off-by-one in a predicate because there's no index. You can't shadow a loop variable because there's no loop. You can't forget to initialize an accumulator because there's no accumulator. These aren't hypothetical — they're the bug classes code review catches every week. fluentfp makes them structurally impossible.

**Works with Go, not against it.** Mappers are slices — callers never import fluentfp. Options use comma-ok (`.Get() (T, bool)`), the same pattern as map lookups and type assertions. `either.Fold` gives you exhaustive dispatch the compiler enforces — miss a branch and it doesn't compile. `must.BeNil` makes invariant enforcement explicit. Mutation, channels, and hot paths stay as loops.

## Performance

Chains pre-allocate with `make([]T, 0, len(input))` internally — the same thing a well-written loop does. Throughput and allocations are identical to pre-allocated loops.

| Operation | Pre-allocated Loop | Chain |
|-----------|-------------------|-------|
| Filter only | 1 alloc | 1 alloc |
| Filter + Map | 2 allocs | 2 allocs |

*1000 elements. See [full benchmarks](methodology.md#benchmark-results).*

Multi-step chains pay one allocation per stage. If you're counting nanoseconds, use a raw loop. The other 95% of your loops aren't hot paths.

## Measurable Impact

| Codebase Type | Code Reduction | Complexity Reduction |
|---------------|----------------|---------------------|
| Mixed (typical) | 12% | 26% |
| Pure pipeline | 47% | 95% |

*Individual loops see up to 6x line reduction. Codebase-wide averages are lower because not every line is a loop. Complexity measured via `scc`. See [methodology](methodology.md#code-metrics-tool-scc).*

## When to Use Loops

fluentfp replaces mechanical loops — iteration scaffolding around a predicate and a transform. It doesn't try to replace loops that do structural work:

```go
// Mutation in place — fluentfp operates on copies, not originals
for i := range items {
    if items[i].ID == target {
        items[i].Status = "done"
        break
    }
}
```

Channel consumption (`for msg := range ch`), complex control flow (early return, labeled break), and performance-critical hot paths also stay as loops.

## Packages

Packages are independent — import one or all.

| Package | Purpose | Key Functions |
|---------|---------|---------------|
| [slice](slice/) | Collection transforms | `KeepIf`, `RemoveIf`, `Fold`, `FanOut` |
| [kv](kv/) | Map transforms | `KeepIf`, `MapValues`, `Map`, `Values` |
| [option](option/) | Nil safety | `Of`, `Get`, `Or`, `NonZero`, `NonNil` |
| [either](either/) | Sum types | `Left`, `Right`, `Fold`, `Map` |
| [result](result/) | Typed error handling | `Ok`, `Err`, `CollectAll`, `CollectOk` |
| [must](must/) | Invariant enforcement | `Get`, `BeNil`, `Of` |
| [value](value/) | Conditional value selection | `Of().When().Or()` |
| [stream](stream/) | Lazy sequences | `Generate`, `Unfold`, `Take`, `Collect` |
| [hof](hof/) | Function combinators | `Pipe`, `Bind`, `Throttle`, `OnErr` |
| [pair](tuple/pair/) | Zip slices | `Zip`, `ZipWith` |
| [lof](lof/) | Lower-order function wrappers | `Len`, `Println`, `Identity` |

Zero reflection. Zero global state. Zero build tags.

## Further Reading

- [Full Analysis](analysis.md) - Technical deep-dive with examples
- [Methodology](methodology.md) - How claims were measured
- [Nil Safety](nil-safety.md) - The billion-dollar mistake and Go
- [Naming Functions](naming-in-hof.md) - Function naming patterns for HOF use
- [Library Comparison](comparison.md) - How fluentfp compares to alternatives
- [Real-World Showcase](docs/showcase.md) - Before/after rewrites from GitHub projects

## Recent Additions

- **v0.41.0**: `stream` package — lazy memoized persistent sequences (`Generate`, `Unfold`, `Take`, `KeepIf`, `Seq`)
- **v0.40.0**: `result` package — `Result[R]` type with `Ok`/`Err`, `CollectAll`/`CollectOk`, `PanicError`
- **v0.40.0**: `slice.FanOut`/`FanOutEach` — bounded concurrent traversal with per-item results and panic recovery
- **NEW**: `hof` package — `Pipe`, `Bind`, `Throttle`, `OnErr` function combinators
- **NEW**: `kv.MapValues`, `Entries.KeepIf`/`RemoveIf` — map value transforms and entry filtering

## License

fluentfp is licensed under the MIT License. See [LICENSE](LICENSE) for more details.
