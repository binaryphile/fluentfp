# fluentfp

Pragmatic functional programming for Go: fewer bugs, less code, predictable performance.

> **Summary:** Eliminate control structures, eliminate the bugs they enable.
> Mixed codebases see 26% complexity reduction; pure pipelines drop 95%.
> The win isn't lines saved—it's bugs that become unwritable.

fluentfp is a small set of composable utilities for slice-based data transformation in Go.

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

- **Index typos**: `i+i` instead of `i+1`
- **Defer in loop**: resources pile up until function returns
- **Accumulator errors**: forgot to increment, wrong variable
- **Off-by-one**: `i <= n` instead of `i < n`

These bugs compile, pass review, and look correct. They continue to appear in highly-reviewed, very public projects. If the construct allows an error, it will eventually happen.

## The Solution

*Correctness by construction*: design code so errors can't occur.

| Bug Class | Why It Happens | fluentfp Elimination |
|-----------|----------------|---------------------|
| Index typo | Manual index math | No index—predicates operate on values |
| Defer in loop | Loop body accumulates | No loop body |
| Accumulator error | Manual state tracking | `Fold` manages state |
| Off-by-one | Manual bounds | Iterate collection, not indices |

## Measurable Impact

| Codebase Type | Code Reduction | Complexity Reduction |
|---------------|----------------|---------------------|
| Mixed (typical) | 12% | 26% |
| Pure pipeline | 47% | 95% |

*Complexity measured via `scc` (cyclomatic complexity approximation). See [methodology](methodology.md#code-metrics-tool-scc).*

## Performance

| Operation | Loop | Chain | Winner |
|-----------|------|-------|--------|
| Filter only | 6.6 μs | 5.8 μs | **Chain 13% faster** |
| Filter + Map | 4.2 μs | 7.6 μs | Loop 45% faster |
| Count only | 0.26 μs | 7.3 μs | Loop 28× faster |

Single operations often equal or beat loops (fluentfp pre-allocates; naive loops dynamically append). Multi-operation chains allocate per operation. See [full benchmarks](methodology.md#benchmark-results).

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
| [option](option/) | Nil safety | `Of`, `Get`, `Or`, `IfProvided` |
| [must](must/) | Error-to-panic for init | `Get`, `BeNil`, `Of` |
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
opt := option.IfProvided(name)   // ok if non-zero

// Extract
user, ok := opt.Get()            // comma-ok
user := opt.Or(defaultUser)      // with fallback
```

### must

Convert fallible functions for init sequences:

```go
db := must.Get(sql.Open("postgres", dsn))
must.BeNil(db.Ping())
home := must.Getenv("HOME")
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

- **v0.6.0**: `Fold`, `Unzip2/3/4`, `Zip`/`ZipWith` (pair package)
- **v0.5.0**: `ToFloat64`, `ToFloat32`

## License

fluentfp is licensed under the MIT License. See [LICENSE](LICENSE) for more details.
