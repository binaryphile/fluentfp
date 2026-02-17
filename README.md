# fluentfp

**Fluent functional programming for Go.**

Chain type-safe operations on slices, options, and sum types — no loop scaffolding, no intermediate variables, no reflection. The bugs you can't write are the bugs you'll never debug.

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

## The Problem

Loops mix intent with mechanics. Every loop manually manages state, bounds, mutation, and control flow — four failure modes before you've expressed your actual logic.

- **Accumulator errors**: forgot to increment, wrong variable
- **Defer in loop**: resources pile up until function returns
- **Index typos**: `i+i` instead of `i+1`
- **Off-by-one**: `i <= n` instead of `i < n`
- **Ignored errors**: `_ = fn()` silently continues when "impossible" errors occur

These bugs compile, pass review, and ship. They recur in every codebase because the construct permits them.

## The Solution

Remove the mechanics and the bugs have nowhere to live. *Correctness by construction* — design code so errors can't occur.

| Bug Class | With Loops | With fluentfp |
|-----------|-----------|---------------|
| Accumulator error | You manage state | `Fold` manages state |
| Defer in loop | Loop body accumulates | No loop body |
| Index typo | You manage index math | Predicates operate on values |
| Off-by-one | You manage bounds | Iterate collection, not indices |
| Ignored error | `_ = fn()` silent failure | `must.BeNil(fn())` explicit invariant |

## Real-World Usage

### [era](https://github.com/binaryphile/era) — Semantic Memory for AI Agents

One package (`slice`), four call sites, identical idiom across in-memory and SQLite-vec backends:

```go
filterSet := slice.String(opts.Tags).ToSet()
inFilterSet := func(tag string) bool { return filterSet[tag] }

if len(opts.Tags) > 0 && !slice.From(m.Tags).Any(inFilterSet) {
    continue
}
```

Also: `SortByDesc(...).TakeFirst(limit)` for top-K retrieval.

### [sofdevsim](https://github.com/binaryphile/sofdevsim-2026) — Software Development Simulator

Five packages, 37 files. Each mapped to a distinct domain problem:

```go
// Exhaustive mode dispatch — compiler-enforced for both modes
header := either.Fold(a.mode,
    func(eng EngineMode) HeaderVM { ... },
    func(_ ClientMode) HeaderVM { ... },
)
```

Also: `Unzip4` for single-pass multi-field extraction, `must.Get` for invariant enforcement, `Convert` for immutable updates, `value.Of().When().Or()` for conditional selection, `option.Lift` for conditional logic on optionals, `ToString` for rendering (12 TUI call sites).

Where FP doesn't fit — early exits, complex state machines — the codebase uses imperative loops with comments citing the specific guide section that says not to.

## When to Use Each

**Use fluentfp for:** filter/map/fold, field extraction, data pipelines, API transforms, immutable updates, conditional value selection.

**Use a loop for:** channel consumption (`for r := range ch`), complex control flow (break, continue, early return), index-dependent logic.

## Performance

Chains beat the loops you actually ship — the ones that use naive `append` instead of pre-allocating. The benchmark below compares against tuned loops with pre-allocation. In production, nobody writes those in handlers.

| Operation | Loop | Chain | Result |
|-----------|------|-------|--------|
| Filter only | 5.6 μs | 5.5 μs | **Equal** |
| Filter + Map | 3.1 μs | 7.6 μs | Loop 2.5× faster |

Single operations match tuned loops. Multi-step chains allocate per step — the same tradeoff as any builder pattern in Go. See [full benchmarks](methodology.md#benchmark-results).

## Measurable Impact

| Codebase Type | Code Reduction | Complexity Reduction |
|---------------|----------------|---------------------|
| Mixed (typical) | 12% | 26% |
| Pure pipeline | 47% | 95% |

*Complexity measured via `scc` (cyclomatic complexity approximation). See [methodology](methodology.md#code-metrics-tool-scc).*

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

## Package Examples

```go
// slice — filter, extract, reduce
names := slice.From(users).KeepIf(User.IsActive).ToString(User.GetName)
total := slice.Fold(amounts, 0.0, sumFloat64)

// option — nil safety with explicit optionality
opt := option.IfNotZero(name)       // ok if non-zero
user := opt.Or(defaultUser)         // value or fallback

// either — sum types, exhaustive matching
msg := either.Fold(result, formatErr, formatOk)

// must — error invariant enforcement
must.BeNil(os.Setenv("KEY", value)) // panics if error, replaces _ =
ints := slice.From(strs).ToInt(must.Of(strconv.Atoi))

// value — conditional selection (Go's missing ternary)
days := value.Of(tick).When(tick < 7).Or(7)
```

## The Familiarity Discount

A `for` loop you've seen 10,000 times feels instant to parse—but only because you've amortized the cognitive load through repetition. fluentfp expresses intent without mechanics; the simplicity is inherent, not learned. Be aware of this discount when comparing approaches.

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
