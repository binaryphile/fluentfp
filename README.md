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
// Before: 3 lines of scaffolding, 2 closing braces, 1 line of intent
var names []string                         // state
for _, u := range users {                  // iteration
    if u.IsActive() {                      // predicate
        names = append(names, u.Name)      // accumulation
    }
}

// After: intent only
names := slice.From(users).KeepIf(User.IsActive).ToString(User.GetName)
```

Six lines become one.

That's a **fluent chain** — each step returns a value you can call the next method on, so the whole pipeline reads as a single expression: filter, then transform.

Every closing brace marks a nesting level, and nesting depth is how tools like [`scc`](https://github.com/boyter/scc) approximate cyclomatic complexity.

- **Interchangeable** — pass `[]User` in, get `[]string` back. No wrapping, no unwrapping.
- **Generics** — 100% type-safe. No `any`, no reflection, no type assertions.
- **Method expressions** — pass `User.IsActive` directly. No wrapper closures.
- **Comma-ok** — `Find`, `IndexWhere` return `option` with `.Get()` → `(value, ok)`.

### Interchangeable Types

```go
// The chain returns Mapper[string], but the function returns []string — no conversion
func activeNames(users []User) []string {
    return slice.From(users).KeepIf(User.IsActive).ToString(User.Name)
}
```

`Mapper[T]` is defined as `type Mapper[T any] []T`. Everything that works on `[]T` works on `Mapper[T]` — index, `range`, `append`, `len`, pass to functions, return from functions. No conversion needed in either direction. Keep `[]T` in your function signatures and use `From()` at the point of use — fluentfp stays an implementation detail. Other Go FP libraries use `[]any` internally, requiring type assertions on both ends. See [comparison](comparison.md).

Full treatment in [It's Just a Slice](slice/#its-just-a-slice).

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

### Conditional Initialization
```go
return Alert{
    Message: msg,
    Level:   value.Of("critical").When(overdue).Or("info"),
    Icon:    value.Of("!").When(overdue).Or("✓"),
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
| Ignored error | `_ = fn()` | `err := fn()`<br>`must.BeNil(err)` |

## Performance

Chains pre-allocate with `make([]T, 0, len(input))` internally — the same thing a well-written loop does. Throughput and allocations are identical to pre-allocated loops.

| Operation | Pre-allocated Loop | Chain |
|-----------|-------------------|-------|
| Filter only | 1 alloc | 1 alloc |
| Filter + Map | 2 allocs | 2 allocs |

*1000 elements. See [full benchmarks](methodology.md#benchmark-results).*

Multi-step chains pay one allocation per stage. Execution time varies — single-stage chains match raw loops, but multi-stage chains add overhead from intermediate slices and function call indirection. If you're counting nanoseconds, use a raw loop.

## Measurable Impact

| Codebase Type | Code Reduction | Complexity Reduction |
|---------------|----------------|---------------------|
| Mixed (typical) | 12% | 26% |
| Pure pipeline | 47% | 95% |

*Individual loops see up to 6× line reduction (as above). Codebase-wide averages are lower because not every line is a loop. Complexity measured via `scc`. See [methodology](methodology.md#code-metrics-tool-scc).*

## Adopt What Fits

Packages are independent — import one or all. A CLI might use only `slice` and `must`. A domain with sum-type state might add `either` and `option`. Same library, different surface area.

## When to Use Loops

The filter+map chain in Quick Start is a mechanical loop — iteration scaffolding around a predicate and a transform. fluentfp replaces those.

It doesn't try to replace loops that do structural work. The most common: mutation in place.

```go
// Find by ID, update, break — fluentfp operates on copies, not originals
for i := range items {
    if items[i].ID == target {
        items[i].Status = "done"
        break
    }
}
```

fluentfp builds new slices from old ones (functional transforms). This loop modifies an element in the original slice by index — a fundamentally different operation.

Channel consumption (`for msg := range ch`), complex control flow (early return, labeled break), and performance-critical hot paths also stay as loops.

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

Zero reflection. Zero global state. Zero build tags.

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
