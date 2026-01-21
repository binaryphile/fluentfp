# fluentfp

Pragmatic functional programming for Go: fewer bugs, less code, predictable performance.

> **Summary:** Eliminate control structures, eliminate the bugs they enable.
> Mixed codebases see 26% complexity reduction; pure pipelines drop 95%.
> The win isn't lines saved—it's bugs that become unwritable.

fluentfp is a small set of composable utilities for data transformation and type safety in Go.

## Quick Start

```bash
go get github.com/binaryphile/fluentfp
```

```go
// Before: loop mechanics interleaved with intent
var names []string
for _, u := range users {
    if u.IsActive() {
        names = append(names, u.Name)
    }
}

// After: just intent
names := slice.From(users).KeepIf(User.IsActive).ToString(User.GetName)
```

## The Problem

Loop mechanics create bugs regardless of developer skill:

- **Accumulator errors**: forgot to increment, wrong variable
- **Defer in loop**: resources pile up until function returns
- **Index typos**: `i+i` instead of `i+1`

C-style loops add off-by-one errors: `i <= n` instead of `i < n`.

These bugs compile, pass review, and look correct. They continue to appear in highly-reviewed, very public projects. If the construct allows an error, it will eventually happen.

## The Solution

*Correctness by construction*: design code so errors can't occur.

| Bug Class | Why It Happens | fluentfp Elimination |
|-----------|----------------|---------------------|
| Accumulator error | Manual state tracking | `Fold` manages state |
| Defer in loop | Loop body accumulates | No loop body |
| Index typo | Manual index math | Predicates operate on values |
| Off-by-one (C-style) | Manual bounds | Iterate collection, not indices |

## Measurable Impact

| Codebase Type | Code Reduction | Complexity Reduction |
|---------------|----------------|---------------------|
| Mixed (typical) | 12% | 26% |
| Pure pipeline | 47% | 95% |

*Complexity measured via `scc` (cyclomatic complexity approximation). See [methodology](methodology.md#code-metrics-tool-scc).*

## Performance

| Operation | Loop | Chain | Result |
|-----------|------|-------|--------|
| Filter only | 5.6 μs | 5.5 μs | **Equal** |
| Filter + Map | 3.1 μs | 7.6 μs | Loop 2.5× faster |
| Count only | 0.26 μs | 7.6 μs | Loop 29× faster |

Single operations equal properly-written loops (both pre-allocate). In practice, many loops use naive append for simplicity—chains beat those. Multi-operation chains allocate per operation. See [full benchmarks](methodology.md#benchmark-results).

## When to Use fluentfp

**High yield** (adopt broadly):
- Data pipelines, ETL, report generators
- Filter/map/fold patterns
- Field extraction from collections

**Medium yield** (adopt selectively):
- API handlers with data transformation
- Config validation

**Low yield** (probably skip):
- I/O-heavy code with minimal transformation
- Graph/tree traversal
- Streaming/channel-based pipelines

## When to Use Loops

- **Channel consumption**: `for r := range ch`
- **Complex control flow**: break, continue, early return
- **Index-dependent logic**: when you need `i` for more than indexing

## Packages

| Package | Purpose | Key Functions |
|---------|---------|---------------|
| [slice](slice/) | Collection transforms | `KeepIf`, `RemoveIf`, `Fold`, `ToString` |
| [option](option/) | Nil safety | `Of`, `Get`, `Or`, `IfProvided`, `IfNotZero` |
| [either](either/) | Sum types | `Left`, `Right`, `Fold`, `Map` |
| [must](must/) | Fallible funcs → HOF args | `Get`, `BeNil`, `Of` |
| [ternary](ternary/) | Conditional expressions | `If().Then().Else()` |
| [pair](tuple/pair/) | Zip slices | `Zip`, `ZipWith` |

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
users := slice.MapTo[User](ids).To(FetchUser)

// Reduce
total := slice.Fold(amounts, 0.0, sumFloat64)
```

### option

Eliminate nil panics with explicit optionality:

```go
// Create
opt := option.Of(user)           // always ok
opt := option.IfProvided(name)   // ok if non-zero (comparable types)
opt := option.IfNotZero(reg)     // ok if !reg.IsZero() (ZeroChecker types)

// Extract
user, ok := opt.Get()            // comma-ok
user := opt.Or(defaultUser)      // with fallback
```

### either

Sum types for values that are one of two possible types:

```go
// Create
left := either.Left[string, int]("error")
right := either.Right[string, int](42)

// Extract with comma-ok
if val, ok := result.Get(); ok { /* use val */ }

// Pattern match with Fold
msg := either.Fold(result,
    func(err string) string { return "Error: " + err },
    func(val int) string { return fmt.Sprintf("Got: %d", val) },
)
```

### must

Convert fallible functions for use with higher-order functions:

```go
mustAtoi := must.Of(strconv.Atoi)
ints := slice.From(strings).ToInt(mustAtoi)
```

### ternary

Conditional expressions:

```go
status := ternary.If[string](done).Then("complete").Else("pending")
```

## The Familiarity Discount

A `for` loop you've seen 10,000 times feels instant to parse—but only because you've amortized the cognitive load through repetition. fluentfp expresses intent without mechanics; the simplicity is inherent, not learned. Be aware of this discount when comparing approaches.

## Further Reading

- [Full Analysis](analysis.md) - Technical deep-dive with examples
- [Methodology](methodology.md) - How claims were measured
- [Nil Safety](nil-safety.md) - The billion-dollar mistake and Go

## Recent Additions

- **v0.8.0**: `either` package (Left/Right sum types), `ToInt32`/`ToInt64` (slice package)
- **v0.7.0**: `IfNotZero`, `ZeroChecker` interface (option package)
- **v0.6.0**: `Fold`, `Unzip2/3/4`, `Zip`/`ZipWith` (pair package)
- **v0.5.0**: `ToFloat64`, `ToFloat32`

## License

fluentfp is licensed under the MIT License. See [LICENSE](LICENSE) for more details.
