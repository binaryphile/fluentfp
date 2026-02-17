# fluentfp

**Fluent functional programming for Go.**

The thinnest abstraction that eliminates mechanical bugs from Go. Chain type-safe operations on slices, options, and sum types — no loop scaffolding, no intermediate variables, no reflection.

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp) for complete API documentation.

## Quick Start

```bash
go get github.com/binaryphile/fluentfp
```

```go
// Before: 5 lines of mechanics around 1 line of intent
var names []string                         // state
for _, u := range users {                  // iteration
    if u.IsActive() {                      // predicate
        names = append(names, u.Name)      // accumulation
    }
}

// After: intent only
names := slice.From(users).KeepIf(User.IsActive).ToString(User.GetName)
```

## Look What You Can Write

### Exhaustive Sum Types
Add a third case — the compiler breaks the build until you handle it.
```go
msg := either.Fold(result,
    func(err string) string { return "Error: " + err },
    func(val int) string    { return fmt.Sprintf("Got: %d", val) },
)
```

### Nil Safety
```go
name := option.Of(user).Map(User.GetProfile).Map(Profile.GetNickname).Or("Guest")
```

### Invariant Enforcement
Silent `_ = fn()` failures become explicit panics.
```go
must.BeNil(os.Setenv("PORT", port))
port := must.Get(strconv.Atoi(os.Getenv("PORT")))
```

### Conditional Initialization
```go
vm := HeaderVM{
    Title:  title,
    Color:  value.Of("red").When(critical).Or("green"),
    Icon:   value.Of("!").When(critical).Or("✓"),
}
```

## Why It Exists

Loops force you to manage state, bounds, and mutation manually — four failure modes before you've expressed your actual logic.

| Bug Class | With Loops | With fluentfp |
|-----------|-----------|---------------|
| Accumulator error | You manage state | `Fold` manages state |
| Defer in loop | Loop body accumulates | No loop body |
| Index typo | You manage index math | Predicates operate on values |
| Off-by-one | You manage bounds | Iterate collection, not indices |
| Ignored error | `_ = fn()` | `must.BeNil(fn())` |

## Performance

Single-pass operations match tuned loops. Multi-step chains allocate per stage — same tradeoff as any builder pattern in Go. If you're counting nanoseconds, write a loop.

| Operation | Loop | Chain | Result |
|-----------|------|-------|--------|
| Filter only | 5.6 μs | 5.5 μs | **Equal** |
| Filter + Map | 3.1 μs | 7.6 μs | Loop 2.5× faster |

Benchmarks compare against pre-allocated loops. In production, most loops use naive `append` — chains beat those. See [full benchmarks](methodology.md#benchmark-results).

## Measurable Impact

| Codebase Type | Code Reduction | Complexity Reduction |
|---------------|----------------|---------------------|
| Mixed (typical) | 12% | 26% |
| Pure pipeline | 47% | 95% |

*Complexity measured via `scc` (cyclomatic complexity approximation). See [methodology](methodology.md#code-metrics-tool-scc).*

## Real-World Usage

Used in production in [era](https://codeberg.org/binaryphile/era) (semantic memory CLI) and [sofdevsim](https://github.com/binaryphile/sofdevsim-2026) (TUI simulator with DORA metrics and event sourcing). Between them: 40+ files, ~90 call sites, every pattern from tag filtering to exhaustive mode dispatch to crisis-detection state machines.

```go
// Tag filtering — identical idiom across two storage backends
filterSet := slice.String(opts.Tags).ToSet()
inFilterSet := func(tag string) bool { return filterSet[tag] }
if len(opts.Tags) > 0 && !slice.From(m.Tags).Any(inFilterSet) {
    continue
}
```

Each project imports only what it needs. era uses one package (`slice`). sofdevsim uses five (`slice`, `option`, `either`, `must`, `value`). Same library, different surface area — take what fits your domain.

## Packages

| Package | Purpose | Key Functions |
|---------|---------|---------------|
| [slice](slice/) | Collection transforms | `KeepIf`, `RemoveIf`, `Fold`, `ToString` |
| [option](option/) | Nil safety | `Of`, `Get`, `Or`, `IfNotZero`, `IfNotNil` |
| [either](either/) | Sum types | `Left`, `Right`, `Fold`, `Map` |
| [must](must/) | Fallible funcs → HOF args | `Get`, `BeNil`, `Of` |
| [value](value/) | Conditional value selection | `Of().When().Or()` |
| [pair](tuple/pair/) | Zip slices | `Zip`, `ZipWith` |
| [lof](lof/) | Lower-order function wrappers | `Len`, `Println`, `StringLen` |

Zero reflection. Zero global state. Zero build tags. Just Go code that compiles and stays compiled.

## Further Reading

- [Full Analysis](analysis.md) - Technical deep-dive with examples
- [Methodology](methodology.md) - How claims were measured
- [Nil Safety](nil-safety.md) - The billion-dollar mistake and Go
- [Naming Functions](naming-in-hof.md) - Function naming patterns for HOF use
- [Library Comparison](comparison.md) - How fluentfp compares to alternatives

## Recent Additions

- **v0.14.0**: `value` package replaces `ternary` — value-first conditional selection
- **v0.12.0**: **BREAKING** — `MapperTo.To` renamed to `MapperTo.Map` for clarity
- **v0.8.0**: `either` package (Left/Right sum types), `ToInt32`/`ToInt64` (slice package)
- **v0.7.0**: `IfNotZero` for comparable types (option package)
- **v0.6.0**: `Fold`, `Unzip2/3/4`, `Zip`/`ZipWith` (pair package)
- **v0.5.0**: `ToFloat64`, `ToFloat32`

## License

fluentfp is licensed under the MIT License. See [LICENSE](LICENSE) for more details.
