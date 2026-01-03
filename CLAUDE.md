@/home/ted/projects/share/tandem-protocol/tandem-protocol.md

# FluentFP - Functional Programming Library for Go

## Development Environment

- **Language**: Go
- **Package Management**: Go modules

## Code Style: FluentFP

### slice Package - Complete API

```go
import "github.com/binaryphile/fluentfp/slice"

// Factory functions
slice.From(ts []T) Mapper[T]           // For mapping to built-in types
slice.MapTo[R](ts []T) MapperTo[R,T]   // For mapping to arbitrary type R

// Mapper[T] methods (also on MapperTo)
.KeepIf(fn func(T) bool) Mapper[T]     // Filter: keep matching
.RemoveIf(fn func(T) bool) Mapper[T]   // Filter: remove matching
.Convert(fn func(T) T) Mapper[T]       // Map to same type
.TakeFirst(n int) Mapper[T]            // First n elements
.Each(fn func(T))                      // Side-effect iteration
.Len() int                             // Count elements

// Mapping methods (return Mapper of target type)
.ToAny(fn func(T) any) Mapper[any]
.ToBool(fn func(T) bool) Mapper[bool]
.ToByte(fn func(T) byte) Mapper[byte]
.ToError(fn func(T) error) Mapper[error]
.ToFloat32(fn func(T) float32) Mapper[float32]
.ToFloat64(fn func(T) float64) Mapper[float64]
.ToInt(fn func(T) int) Mapper[int]
.ToRune(fn func(T) rune) Mapper[rune]
.ToString(fn func(T) string) Mapper[string]

// MapperTo[R,T] additional method
.To(fn func(T) R) Mapper[R]            // Map to type R

// Standalone functions
slice.Fold[T, R](ts []T, initial R, fn func(R, T) R) R
slice.Unzip2[T, A, B](ts []T, fa func(T) A, fb func(T) B) (Mapper[A], Mapper[B])
slice.Unzip3[T, A, B, C](...)
slice.Unzip4[T, A, B, C, D](...)
```

### slice Patterns

```go
// Count matching elements
count := slice.From(tickets).KeepIf(Ticket.IsActive).Len()

// Extract field to strings
ids := slice.From(tickets).ToString(Ticket.GetID)

// Method expressions for clean chains
actives := slice.From(users).
    Convert(User.Normalize).
    KeepIf(User.IsValid)

// Fold - reduce slice to single value
total := slice.Fold(amounts, 0.0, func(acc, x float64) float64 { return acc + x })

// Unzip - extract multiple fields in one pass
a, b, c, d := slice.Unzip4(items, Item.GetA, Item.GetB, Item.GetC, Item.GetD)
```

### option Package

```go
import "github.com/binaryphile/fluentfp/option"

// Creating options
option.Of(t T) Basic[T]                // Always ok
option.New(t T, ok bool) Basic[T]      // Conditional ok
option.IfProvided(t T) Basic[T]        // Ok if non-zero value
option.FromOpt(ptr *T) Basic[T]        // From pointer (nil = not-ok)

// Using options
.Get() (T, bool)                       // Comma-ok unwrap
.Or(t T) T                             // Value or default
.OrZero() T                            // Value or zero
.OrEmpty() T                           // Alias for strings
.OrFalse() bool                        // For option.Bool
.Call(fn func(T))                      // Side-effect if ok

// Pre-defined types
option.String, option.Int, option.Bool, option.Error
```

### must Package

```go
import "github.com/binaryphile/fluentfp/must"

must.Get(t T, err error) T             // Return or panic
must.BeNil(err error)                  // Panic if error
must.Getenv(key string) string         // Env var or panic
must.Of(fn func(T) (R, error)) func(T) R  // Wrap fallible func
```

### ternary Package

```go
import "github.com/binaryphile/fluentfp/ternary"

ternary.If[R](cond bool).Then(t R).Else(e R) R
ternary.If[R](cond bool).ThenCall(fn).ElseCall(fn) R  // Lazy
```

### lof Package (Lower-Order Functions)

```go
import "github.com/binaryphile/fluentfp/lof"

lof.Println(s string)      // Wraps fmt.Println for Each
lof.Len(ts []T) int        // Wraps len
```

### pair Package (Tuples)

```go
import "github.com/binaryphile/fluentfp/tuple/pair"

pair.X[V1, V2]             // Struct with V1, V2 fields
pair.Of(v1, v2) X[V1,V2]   // Construct a pair
pair.Zip(as, bs) []X[A,B]           // Combine into pairs
pair.ZipWith(as, bs, fn) []R        // Combine and transform
```

### Named vs Inline Functions

**Preference hierarchy** (best to worst):
1. **Method expressions** - `User.IsActive`, `Device.GetMAC` (cleanest, no function body)
2. **Named functions** - `isActive := func(u User) bool {...}` (readable, debuggable)
3. **Inline anonymous** - `func(u User) bool {...}` (only when trivial)

**When to name (vs inline):**

| Name when... | Inline when... |
|--------------|----------------|
| Reused (called 2+ times) | Trivial wrapper (single forwarding expression) |
| Complex (multiple statements) | Standard idiom (t.Run, http.HandlerFunc) |
| Has domain meaning | Context already explains intent |
| Stored for later use | |
| Captures outer variables | |

**Locality:** Define named functions close to first usage, not at package level.

#### Predicate Naming Patterns

| Pattern | When to use | Example |
|---------|-------------|---------|
| `Is[Condition]` | Simple check, subject obvious | `IsValidMAC` |
| `[Subject]Is[Condition]` | State check on specific type | `SliceOfScansIsEmpty` |
| `[Subject]Has[Condition](params)` | Parameterized predicate factory | `DeviceHasHWVersion("EX12")` |
| `Type.Is[Condition]` | Method expression | `Device.IsActive` |

#### Reducer Naming

```go
// sumFloat64 adds two float64 values.
sumFloat64 := func(acc, x float64) float64 { return acc + x }
total := slice.Fold(amounts, 0.0, sumFloat64)
```

### When Loops Are Still Necessary

1. **Channel consumption** - `for r := range chan` has no FP equivalent
2. **Complex control flow** - break/continue/early return within loop

## Testing: Khorikov Principles

### Khorikov's Four Quadrants

| Quadrant | Complexity | Collaborators | Test Strategy |
|----------|------------|---------------|---------------|
| **Domain/Algorithms** | High | Few | Unit test heavily (edge cases) |
| **Controllers** | Low | Many | ONE integration test per happy path |
| **Trivial** | Low | Few | **Don't test at all** |
| **Overcomplicated** | High | Many | Refactor first, then test |

### Domain/Algorithms: Unit Test Heavily

**FluentFP-specific domain code:**
- `slice.KeepIf`, `slice.RemoveIf` - conditional inclusion logic
- `slice.TakeFirst` - boundary handling (`if n > len`)
- `slice.Fold`, `slice.Unzip2/3/4` - accumulation and multi-output logic
- `option.New`, `option.IfProvided`, `option.FromOpt` - conditional construction
- `option.Or`, `option.OrCall`, `option.MustGet` - conditional extraction
- `option.KeepOkIf`, `option.ToNotOkIf` - double conditional (filter)
- `ternary.Else`, `ternary.ElseCall` - 3 code paths each

### Representative Pattern Testing

When multiple methods share **identical logic**, test ONE representative:
- `option.ToInt` covers all `option.ToX` methods (same if-ok-then-transform pattern)
- `option.Or` covers `OrZero`/`OrEmpty`/`OrFalse` (same if-ok-return-value pattern)
- `slice.KeepIf` on `Mapper` covers `MapperTo.KeepIf` (identical implementation)

### Trivial Code: Don't Test

- Loop + apply with no branching (e.g., `ToInt`, `ToString` - just iterate and call fn)
- Wrappers around stdlib (e.g., `lof.Len` wraps `len()`)
- Type aliases with identical logic (e.g., `OrZero`, `OrEmpty`, `OrFalse` are same impl)
- `slice.From`, `slice.MapTo` - just return input
- `option.Of`, `option.NotOk` - just construct struct
- `option.Get`, `option.IsOk` - just return fields
- `ternary.If`, `ternary.Then`, `ternary.ThenCall` - just store values

### Coverage Baseline (2026-01-03)

| Package | Coverage | Notes |
|---------|----------|-------|
| must | 100% | All domain (conditional + panic) |
| ternary | 100% | All domain code paths tested |
| option | 51.9% | Domain tested, trivial aliases untested |
| slice | 28.8% | Domain tested, ToX methods trivial |
| lof | 0.0% | All trivial wrappers - acceptable |

### Go Test Style

- Prefer **table-driven tests** for domain algorithms
- Use descriptive test names that explain the behavior being tested
- Group related assertions in subtests with `t.Run()`

Run tests: `go test ./...`
Run with coverage: `go test -cover ./...`

## Branching Strategy: Trunk-Based Development

- **Single trunk**: `main` is the only long-lived branch
- **develop branch**: For in-progress work before PR to main
- **Small, frequent commits**: Commit directly to develop when working
- **PR to main**: Create PR when ready to release
- **Tag releases**: Use semantic versioning (v0.6.0, etc.)
