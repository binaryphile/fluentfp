# fluentfp

**Fluent functional programming for Go.**

The thinnest abstraction that eliminates mechanical bugs from Go. Chain type-safe operations on slices, options, and sum types — no loop scaffolding, no intermediate variables, no reflection.

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp) for complete API documentation.

## Quick Start

Requires Go 1.21+.

```bash
go get github.com/binaryphile/fluentfp
```

```go
// Before: 4 lines of scaffolding, 2 closing braces, 1 line of intent
var names []string                         // state
for _, u := range users {                  // iteration
    if u.IsActive() {                      // predicate
        names = append(names, u.Name)      // accumulation
    }
}

// After: intent only
names := slice.From(users).KeepIf(User.IsActive).ToString(User.GetName)
```

Seven lines become one — a 7× reduction. Every closing brace marks a nesting level, and nesting depth is how tools like [`scc`](https://github.com/boyter/scc) approximate cyclomatic complexity.

- **Interchangeable** (assignable) — `Mapper[T]` has underlying type `[]T`. No conversion needed in either direction.
- **Generics** — 100% type-safe. No `any`, no reflection, no type assertions.
- **Method expressions** — pass `User.IsActive` directly. No wrapper closures.
- **Comma-ok** — `Find`, `IndexWhere` return `option` with `.Get()` → `(value, ok)`.

## What It Looks Like

### Conditional Initialization
```go
vm := HeaderVM{
    Title:  title,
    Color:  value.Of(warn).When(critical).Or(calm),
    Icon:   value.Of("!").When(critical).Or("✓"),
}
```

### Environment Configuration
```go
port := option.Getenv("PORT").Or("8080")
```

### Invariant Enforcement
```go
err := os.Setenv("KEY", value)
must.BeNil(err)

port := must.Get(strconv.Atoi(os.Getenv("PORT")))
```

## Why It Exists

Go's mechanical patterns — loops, nil checks, ignored errors — each carry bug classes that fluentfp eliminates structurally.

| Bug Class | With Loops | With fluentfp |
|-----------|-----------|---------------|
| Accumulator error | You manage state | `Fold` manages state |
| Defer in loop | Loop body accumulates | No loop body |
| Index typo | You manage index math | Predicates operate on values |
| Off-by-one | You manage bounds | Iterate collection, not indices |
| Ignored error | `_ = fn()` | `must.BeNil(fn())` |

## Performance

Most loops use `var out []T` with `append`, which re-allocates as the slice grows. Chains pre-allocate automatically.

| Operation | Append Loop | Chain |
|-----------|------------|-------|
| Filter only | 10 allocs | **1 alloc** |
| Filter + Map | 10 allocs | **2 allocs** |

*1000 elements. See [full benchmarks](methodology.md#benchmark-results).*

You could pre-allocate manually with `make([]T, 0, len(input))` — but almost nobody does. Chains give you that optimization as a side effect of writing clearer code. Multi-step chains pay one allocation per stage; throughput is comparable since the bottleneck is the work inside your predicate, not the chain machinery.

## Measurable Impact

| Codebase Type | Code Reduction | Complexity Reduction |
|---------------|----------------|---------------------|
| Mixed (typical) | 12% | 26% |
| Pure pipeline | 47% | 95% |

*Individual loops see up to 7× line reduction (as above). Codebase-wide averages are lower because not every line is a loop. Complexity measured via `scc`. See [methodology](methodology.md#code-metrics-tool-scc).*

## Real-World Usage

Used in production in [era](https://codeberg.org/binaryphile/era) (semantic memory CLI) and [sofdevsim](https://github.com/binaryphile/sofdevsim-2026) (TUI simulator with DORA metrics and event sourcing). Between them: 40+ files, ~90 call sites, every pattern from tag filtering to exhaustive mode dispatch to crisis-detection state machines.

```go
// Tag filtering — identical idiom across two storage backends
if !slice.String(m.Tags).Matches(opts.Tags) {
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
| [must](must/) | Invariant enforcement | `Get`, `BeNil`, `Of` |
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
