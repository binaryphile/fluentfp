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

These numbers come from real projects. Here's what the code looks like in practice.

### era — Semantic Memory for AI Agents

[era](https://github.com/binaryphile/era) is a CLI and MCP server for AI agent memory backed by SQLite with vector search. It imports one package: `slice`.

**Tag filtering** — the same three-line idiom appears identically in both the in-memory test index and the SQLite-vec production index:

```go
filterSet := slice.String(opts.Tags).ToSet()
inFilterSet := func(tag string) bool { return filterSet[tag] }

// "do any of this memory's tags match the filter?"
if len(opts.Tags) > 0 && !slice.From(m.Tags).Any(inFilterSet) {
    continue
}
```

**Top-K retrieval** — sort by score descending, take the best results:

```go
slice.SortByDesc(results, func(r era.Result) float64 { return r.Score }).TakeFirst(limit)
```

One package, four call sites, one consistent idiom across two completely different storage backends.

### sofdevsim — Software Development Simulator

[sofdevsim](https://github.com/binaryphile/sofdevsim-2026) is a TUI-based simulation with DORA metrics, event sourcing, and office animations. It imports five packages across 37 files — each mapped to a distinct domain problem:

**Exhaustive mode dispatch** — seven `either.Fold` calls form the TUI's rendering strategy, compiler-enforced for both modes:

```go
header := either.Fold(a.mode,
    func(eng EngineMode) HeaderVM { ... },
    func(_ ClientMode) HeaderVM { ... },
)
```

**Single-pass multi-field extraction** — four DORA metrics for sparklines, one pass instead of four loops:

```go
leadTimes, deployFreqs, mttrs, cfrs := slice.Unzip4(dora.History,
    DORASnapshot.GetLeadTimeAvg,
    DORASnapshot.GetDeployFrequency,
    DORASnapshot.GetMTTR,
    DORASnapshot.GetChangeFailRate,
)
```

**Invariant enforcement** — errors in deterministic runs are bugs, not recoverable conditions:

```go
eng = must.Get(eng.AddDeveloper("dev-1", "Alice", 1.0))
eng = must.Get(eng.StartSprint())
```

Additional patterns: immutable slice updates with `Convert`, conditional value selection with `value.Of().When().Or()`, `option.Lift` for conditional logic on optional values, and `ToString` for rendering transforms (12 call sites across the TUI).

Where FP doesn't fit — early exits, complex state machines — the codebase uses imperative loops with comments citing the specific guide section that says not to. The discipline to leave a tool on the shelf is as important as knowing when to reach for it.

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

## Parallelism Readiness

Pure functions + immutable data = safe parallelism.

**Note:** fluentfp does not provide parallel operations. But the patterns it encourages—pure transforms, no shared state—are exactly what makes code *parallel-ready* when you need it.

```go
// With errgroup (idiomatic Go)
import "golang.org/x/sync/errgroup"

var g errgroup.Group
results := make([]Result, len(items))
for i, item := range items {
    i, item := i, item  // capture by value for closure
    g.Go(func() error {
        results[i] = transform(item)  // Safe: transform is pure, i is unique
        return nil
    })
}
g.Wait()
```

**Benchmarked crossover (Go, 8 cores):**

| N | Sequential | Parallel | Speedup | Verdict |
|---|------------|----------|---------|---------|
| 100 | 5.6μs | 9.3μs | 0.6× | Sequential wins |
| 1,000 | 56μs | 40μs | 1.4× | Parallel starts winning |
| 10,000 | 559μs | 200μs | 2.8× | Parallel wins |
| 100,000 | 5.6ms | 1.4ms | 4.0× | Parallel wins decisively |

**When to parallelize:**
- N > 1K items AND CPU-bound transform → yes
- N < 500 OR transform < 100ns → no (overhead dominates)
- I/O-bound (HTTP calls, disk) → yes (waiting is free to parallelize)

**Key insight:** The discipline investment—writing pure transforms—pays off when you need parallelism and don't have to refactor first.

*Reproduce these benchmarks: `go test -bench=. -benchmem ./examples/`*

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

## Installation

```bash
go get github.com/binaryphile/fluentfp
```

```go
import "github.com/binaryphile/fluentfp/slice"
import "github.com/binaryphile/fluentfp/option"
```

## Package Highlights

### slice

Fluent collection operations with method chaining:

```go
// Filter and extract
actives := slice.From(users).KeepIf(User.IsActive)
names := slice.From(users).ToString(User.GetName)

// Map to arbitrary types
users := slice.MapTo[User](ids).Map(FetchUser)

// Reduce
total := slice.Fold(amounts, 0.0, sumFloat64)
```

### option

Eliminate nil panics with explicit optionality:

```go
// Create
opt := option.Of(user)           // always ok
opt := option.IfNotZero(name)    // ok if non-zero (comparable types)
opt := option.IfNotNil(ptr)      // ok if not nil (pointer types)

// Extract
user, ok := opt.Get()            // comma-ok
user := opt.Or(defaultUser)      // with fallback
```

### either

Sum types for values that are one of two possible types:

```go
// Create
fail := either.Left[string, int]("fail")
ok42 := either.Right[string, int](42)

// Extract with comma-ok
if fortyTwo, ok := ok42.Get(); ok {
    fmt.Println(fortyTwo) // 42
}

// Fold: handle both cases exhaustively
// formatLeft returns an error message.
formatLeft := func(err string) string { return "Error: " + err }
// formatRight returns a success message.
formatRight := func(n int) string { return fmt.Sprintf("Got: %d", n) }

msg := either.Fold(ok42, formatLeft, formatRight)   // "Got: 42"
msg = either.Fold(fail, formatLeft, formatRight)    // "Error: fail"
```

### must

Make error invariants explicit. Every `_ = fn()` should be `must.BeNil(fn())`:

```go
_ = os.Setenv("KEY", value)           // Silent corruption if error
must.BeNil(os.Setenv("KEY", value))   // Invariant enforced
```

Also wraps fallible functions for HOF use:

```go
mustAtoi := must.Of(strconv.Atoi)
ints := slice.From(strings).ToInt(mustAtoi)
```

### value

Value-first conditional selection:

```go
// "value of CurrentTick when CurrentTick < 7, or 7"
days := value.Of(tick).When(tick < 7).Or(7)

// Lazy evaluation for expensive computations
config := value.OfCall(loadFromDB).When(useCache).Or(defaultConfig)
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
