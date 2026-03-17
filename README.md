# fluentfp

**Fluent functional programming for Go.**

A thin abstraction that removes common loop-scaffolding bugs from Go. Chain type-safe operations on slices, options, and sum types — no loop scaffolding, no intermediate variables.

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp) for API docs and the **[showcase](docs/showcase.md)** for 16 before/after rewrites from real GitHub projects.

Zero reflection. Zero global state. Zero build tags.

## Quick Start

Requires Go 1.26+.

```bash
go get github.com/binaryphile/fluentfp
```

```go
import "github.com/binaryphile/fluentfp/slice"

// Before: scaffolding around one line of intent
var names []string                         // state
for _, u := range users {                  // iteration
    if u.IsActive() {                      // predicate
        names = append(names, u.Name())    // accumulation
    }
}

// After: intent only
names := slice.From(users).KeepIf(User.IsActive).ToString(User.Name)
```

That's a **fluent chain** — each step returns a value you can call the next method on, so the whole pipeline reads as a single expression: filter, then transform.

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

The **[showcase](docs/showcase.md)** has 16 more, including [Sort, Trim, and Map-to-Slice](docs/showcase.md#sort-and-trim-boilerplate--chenjiandongxsniffer).

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
    Level:   option.When(overdue, "critical").Or("info"),
    Icon:    option.When(overdue, "!").Or("✓"),
}
</code></pre></td></tr>
</table>

Go has no inline conditional expression. `option.When` fills that gap — each field resolves in place, so the struct literal stays a single statement. *From [hashicorp/consul](https://github.com/hashicorp/consul/blob/554b4ba24f86/agent/agent.go#L2482-L2530).*

### Bounded Concurrent Requests

Fetch weather for a list of cities with at most 10 simultaneous goroutines.

<table>
<tr><th>Before — errgroup (21 lines)</th><th>After (1 line)</th></tr>
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
    return slice.FanOutAll(ctx, 10, cities, City)
}
</code></pre></td></tr>
</table>

`FanOutAll` is all-or-nothing: on first error it cancels remaining work and returns that error. `City` passes directly — no wrapper needed. Panics in callbacks are recovered as `*rslt.PanicError` with stack trace.

When you need per-item outcomes instead of all-or-nothing, use `FanOut`:

```go
results := slice.FanOut(ctx, 10, cities, City)
infos, errs := rslt.CollectOkAndErr(results)  // gather successes and failures separately
```

*From the [errgroup pattern](https://encore.dev/blog/advanced-go-concurrency).*

## Why fluentfp

**Type-safe end-to-end.** [go-linq](https://github.com/ahmetb/go-linq) gives you `[]any` back — cast it and hope you got the type right. [lo](https://github.com/samber/lo) requires `func(T, int)` callbacks, so every stdlib function needs a wrapper to discard the unused index. fluentfp uses generics throughout: `Mapper[T]` is `[]T` with methods. If it compiles, you avoid a class of cast, index, and callback-shape mistakes.

**Fewer places for bugs to hide.** No index means no off-by-one in a predicate. No loop variable means no shadowing. No accumulator means no forgetting to initialize one. These are the loop-scaffolding bug classes that code review catches regularly — fluentfp removes the scaffolding where they live.

**Works with Go, not against it.** Mappers are slices — callers of your functions don't need to import fluentfp. Options use comma-ok (`.Get() (T, bool)`), the same pattern as map lookups and type assertions. `either.Fold` gives you exhaustive dispatch the compiler enforces — miss a branch and it doesn't compile. `must.BeNil` makes invariant enforcement explicit. Mutation, channels, and hot paths stay as loops.

## Interchangeable Types

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

## Performance

A single chain step (filter OR map) matches a hand-written loop — `slice.From` is a zero-cost type conversion, and each operation pre-allocates with `make([]T, 0, len(input))`.

Multi-step chains pay one allocation per stage. A chain that filters then maps makes two passes and two allocations where a hand-written loop can fuse them into one. In [benchmarks](methodology.md#benchmark-results) (1000 elements), a two-step chain runs ~2.5× slower than the fused equivalent.

If you're counting nanoseconds in a hot path, fuse it in a loop. Most loops aren't hot paths — they're scaffolding that fluentfp eliminates.

## Measurable Impact

| Codebase Type   | Code Reduction | Complexity Reduction |
| --------------- | -------------- | -------------------- |
| Mixed (typical) | 12%            | 26%                  |
| Pure pipeline   | 47%            | 95%                  |

*Individual loops see up to 6x line reduction. Codebase-wide averages are lower because not every line is a loop. Complexity measured via `scc`. See [methodology](methodology.md#code-metrics-tool-scc).*

## When to Use Loops

fluentfp replaces mechanical loops — iteration scaffolding around a predicate and a transform. It doesn't try to replace loops that do structural work:

```go
// Mutation in place — fluentfp returns new slices, but elements are shared (shallow copy)
for i := range items {
    if items[i].ID == target {
        items[i].Status = "done"
        break
    }
}

// Channel consumption — direct range is simplest for straightforward use
for msg := range ch {
    handle(msg)
}

// Bridge to fluentfp when you want operators on a channel
seq.FromChannel(ctx, ch).KeepIf(valid).Take(10).Each(handle)

// Complex control flow — early return, labeled break
for _, item := range items {
    if item.IsTerminal() {
        return item.Result()
    }
}

// Hot paths — fuse filter+map into one pass when nanoseconds matter
out := make([]R, 0, len(input))
for _, v := range input {
    if keep(v) {
        out = append(out, transform(v))
    }
}
```

## Packages

Packages are independent — import one or all.

| Package             | Purpose                       | Key Functions                                  |
| ------------------- | ----------------------------- | ---------------------------------------------- |
| [slice](slice/)     | Collection transforms         | `KeepIf`, `RemoveIf`, `Fold`, `FanOutAll`      |
| [kv](kv/)           | Map transforms                | `KeepIf`, `MapValues`, `Map`, `Values`         |
| [option](option/)   | Optional values + conditionals | `Of`, `When`, `WhenCall`, `Or`, `NonZero`      |
| [either](either/)   | Sum types                     | `Left`, `Right`, `Fold`, `Convert`, `FlatMap`  |
| [rslt](rslt/)   | Typed error handling          | `Ok`, `Err`, `CollectAll`, `CollectOk`         |
| [must](must/)       | Invariant enforcement         | `Get`, `BeNil`, `Of`                           |
| [stream](stream/)   | Lazy memoized sequences       | `Generate`, `Unfold`, `Take`, `Collect`        |
| [seq](seq/)         | Fluent iter.Seq chains        | `From`, `KeepIf`, `Take`, `Collect`            |
| [heap](heap/)       | Persistent priority queue     | `New`, `Insert`, `Pop`, `Collect`              |
| [hof](hof/)         | Function combinators          | `Pipe`, `Bind`, `Throttle`, `Retry`, `Debouncer` |
| [pair](tuple/pair/) | Zip slices                    | `Zip`, `ZipWith`                               |
| [combo](combo/)     | Combinatorial constructions   | `CartesianProduct`, `Combinations`, `PowerSet` |
| [memo](memo/)       | Memoization                   | `Of`, `Fn`, `FnErr`, `NewLRU`                  |
| [toc](toc/)         | Constrained stage runner with pipeline composition | `Start`, `Pipe`, `NewBatcher`, `Submit`, `CloseInput`, `Wait` |
| [lof](lof/)         | Lower-order function wrappers | `Len`, `Println`, `Identity`, `Inc`            |

## Package Highlights

**[rslt](rslt/)** — typed error handling as values:

```go
r := rslt.Of(strconv.Atoi(input))  // wrap (int, error) → Result[int]
port := r.Or(8080)                   // value or default
```

**[seq](seq/)** — fluent chains on Go's `iter.Seq`:

```go
active := seq.FromIter(maps.Keys(configs)).KeepIf(isActive).Collect()
```

**[stream](stream/)** — lazy memoized sequences:

```go
naturals := stream.Generate(0, lof.Inc)
first10Squares := stream.Map(naturals, square).Take(10).Collect()
```

## Further Reading

- [Full Analysis](analysis.md) - Technical deep-dive with examples
- [Methodology](methodology.md) - How claims were measured
- [Nil Safety](nil-safety.md) - The billion-dollar mistake and Go
- [Naming Functions](naming-in-hof.md) - Function naming patterns for HOF use
- [Library Comparison](comparison.md) - How fluentfp compares to alternatives
- [Real-World Showcase](docs/showcase.md) - Before/after rewrites from GitHub projects

See [CHANGELOG](CHANGELOG.md) for version history.

## License

fluentfp is licensed under the MIT License. See [LICENSE](LICENSE) for more details.
