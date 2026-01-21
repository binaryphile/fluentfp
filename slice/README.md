# slice: fluent collection operations

Eliminate loop bugs with type-safe collection operations. Method expressions make pipelines read like intent:

```go
actives := slice.From(users).KeepIf(User.IsActive).ToString(User.GetName)
```

See the [main README](../README.md) for when to use fluentfp and performance characteristics. See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/slice) for complete API documentation.

## Quick Start

```bash
go get github.com/binaryphile/fluentfp
```

```go
import "github.com/binaryphile/fluentfp/slice"

// Filter and extract
names := slice.From(users).KeepIf(User.IsActive).ToString(User.GetName)

// Map to arbitrary type
users := slice.MapTo[User](emails).To(UserFromEmail)

// Reduce
total := slice.Fold(amounts, 0.0, sumFloat64)
```

## API Reference

### Factory Functions

| Function | Signature | Purpose |
|----------|-----------|---------|
| `From` | `From[T]([]T) Mapper[T]` | Create Mapper from slice |
| `MapTo` | `MapTo[R,T]([]T) MapperTo[R,T]` | Create MapperTo for mapping to type R |

### Mapper Methods

| Method | Signature | Purpose |
|--------|-----------|---------|
| `KeepIf` | `.KeepIf(func(T) bool) Mapper[T]` | Keep elements matching predicate |
| `RemoveIf` | `.RemoveIf(func(T) bool) Mapper[T]` | Remove elements matching predicate |
| `TakeFirst` | `.TakeFirst(n int) Mapper[T]` | First n elements |
| `Convert` | `.Convert(func(T) T) Mapper[T]` | Map to same type |
| `ToString` | `.ToString(func(T) string) Mapper[string]` | Map to string |
| `ToInt` | `.ToInt(func(T) int) Mapper[int]` | Map to int |
| `Each` | `.Each(func(T))` | Side-effect iteration |
| `Len` | `.Len() int` | Count elements |

Other `To[Type]` methods: `ToAny`, `ToBool`, `ToByte`, `ToError`, `ToFloat32`, `ToFloat64`, `ToInt32`, `ToInt64`, `ToRune`

### MapperTo Additional Method

| Method | Signature | Purpose |
|--------|-----------|---------|
| `To` | `.To(func(T) R) Mapper[R]` | Map to arbitrary type R |

### Standalone Functions

| Function | Signature | Purpose |
|----------|-----------|---------|
| `Fold` | `Fold[T,R]([]T, R, func(R,T) R) R` | Reduce slice to single value |
| `Unzip2` | `Unzip2[T,A,B]([]T, func(T)A, func(T)B) (Mapper[A], Mapper[B])` | Extract 2 fields in one pass |
| `Unzip3` | `Unzip3[T,A,B,C](...)` | Extract 3 fields |
| `Unzip4` | `Unzip4[T,A,B,C,D](...)` | Extract 4 fields |

### Type Aliases

For use with `make()`: `Any`, `Bool`, `Byte`, `Error`, `Int`, `Rune`, `String`

```go
words := make(slice.String, 0, 10)
```

## Method Expressions

Method expressions let you pass methods directly to higher-order functions:

```go
slice.From(users).KeepIf(User.IsActive)  // User.IsActive is func(User) bool
```

**Receiver type must match slice element type.** Value receivers work with `[]T`; pointer receivers require `[]*T`:

```go
// Works: value receiver + value slice
func (u User) IsActive() bool { return u.Active }
slice.From(users).KeepIf(User.IsActive)  // ✓

// Fails: pointer receiver + value slice
func (u *User) IsActive() bool { return u.Active }
slice.From(users).KeepIf(User.IsActive)  // ✗ type mismatch
```

## Naming Functions in Chains

### Preference Hierarchy

1. **Method expressions** — `User.IsActive` (cleanest)
2. **Named functions** — with leading comment
3. **Inline lambdas** — trivial one-time use only

### Decision Flowchart

```
Is there a method on the type?
  YES → Method expression: User.IsActive
  NO  → Has domain meaning?
        YES → Named function + comment
        NO  → Trivial?
              YES → Inline lambda OK
              NO  → Name it anyway
```

### Comment Format

Start with function name, succinct sentence on return value:

```go
// completedAfterCutoff reports whether ticket was completed after cutoff.
completedAfterCutoff := func(t Ticket) bool { return t.CompletedTick >= cutoff }

// sumFloat64 adds two float64 values.
sumFloat64 := func(acc, x float64) float64 { return acc + x }
```

### Naming Patterns

**Predicates** (`func(T) bool`):

| Pattern | Example |
|---------|---------|
| `Is[Condition]` | `IsValid`, `IsExpired` |
| `[Subject][Verb]` | `TicketIsComplete` |
| `Type.Method` | `User.IsActive` |

**Reducers** (`func(R, T) R`):

| Pattern | Example |
|---------|---------|
| `sum[Type]` | `sumFloat64` |
| `max[Type]` | `maxDuration` |
| `[verb][Subject]` | `accumulateRevenue` |

### Taking It Too Far

- Don't wrap `User.IsActive` — already named
- 3-4 chain operations is fine — split for meaning, not length
- Don't comment obvious folds like `sum`
- Don't extract predicates for simple `if` statements — `if u.IsActive()` is clearer than `if isActive(u)`
- Don't name single field access — `func(u User) string { return u.Name }` is fine inline

## Pipeline Formatting

**Single operation** — one line:

```go
names := slice.From(users).ToString(User.GetName)
```

**Multiple operations** — one per line:

```go
result := slice.From(items).
    KeepIf(Item.IsValid).
    RemoveIf(Item.IsExpired).
    ToInt(Item.Score)
```

**Split at conceptual boundaries** when chains get long:

```go
validCurrent := slice.From(items).
    KeepIf(Item.IsValid).
    RemoveIf(Item.IsExpired)
scores := validCurrent.
    ToInt(Item.Score).
    KeepIf(aboveThreshold)
```

## Patterns

### Type Alias for Domain Slices

```go
type SliceOfUsers = slice.Mapper[User]

var users SliceOfUsers = fetchUsers()
actives := users.KeepIf(User.IsActive)
```

Avoids repeated `slice.From()` calls.

### Method Expression Chaining

```go
devices := slice.From(rawDevices).
    Convert(Device.Normalize).
    KeepIf(Device.IsValid)
```

### Field Extraction

```go
macs := devices.ToString(Device.GetMAC)
```

### Counting

```go
activeCount := slice.From(users).KeepIf(User.IsActive).Len()
```

**Note:** Allocates intermediate slice. For hot paths, use a manual loop.

### Fold

```go
// sumFloat64 adds two float64 values.
sumFloat64 := func(acc, x float64) float64 { return acc + x }
total := slice.Fold(amounts, 0.0, sumFloat64)

// indexByMAC adds a device to the map keyed by its MAC address.
indexByMAC := func(m map[string]Device, d Device) map[string]Device {
    m[d.MAC] = d
    return m
}
byMAC := slice.Fold(devices, make(map[string]Device), indexByMAC)
```

### Unzip

Extract multiple fields in one pass (more efficient than separate `ToX` calls):

```go
leadTimes, deployFreqs, mttrs, cfrs := slice.Unzip4(history,
    Record.GetLeadTime,
    Record.GetDeployFreq,
    Record.GetMTTR,
    Record.GetChangeFailRate,
)
```

### Zip (pair package)

```go
import "github.com/binaryphile/fluentfp/tuple/pair"

pairs := pair.Zip(names, scores)
// []pair.X[string, int]{{V1: "Alice", V2: 95}, ...}

results := pair.ZipWith(names, scores, formatScore)
// []string{"Alice: 95", ...}
```

Both panic if slices have different lengths.

## When NOT to Use slice

- **Early exit needed** — `KeepIf` iterates entire slice; use loop with `break` for first match
- **Accumulating into maps** — No fluent equivalent; use `Fold` or a loop
- **Performance-critical hot paths** — Profile first; each chain operation allocates
- **Single simple operation** — `for _, u := range users` may be clearer than `slice.From(users).Each(...)`

## When Loops Are Necessary

- **Channel consumption**: `for r := range ch`
- **Complex control flow**: break, continue, early return
- **Index-dependent logic**: comparing adjacent elements, position-aware output

## Appendix: Library Comparison

| Library | Type-Safe | Concise | Method Exprs | Fluent |
|---------|-----------|---------|--------------|--------|
| binaryphile/fluentfp | ✅ | ✅ | ✅ | ✅ |
| samber/lo | ✅ | ❌ | ❌ | ❌ |
| thoas/go-funk | ❌ | ✅ | ✅ | ❌ |
| ahmetb/go-linq | ❌ | ❌ | ❌ | ✅ |
| rjNemo/underscore | ✅ | ✅ | ✅ | ❌ |

**fluentfp vs lo:** `lo` passes indexes to all callbacks, requiring wrappers. Non-fluent style requires intermediate variables. See [examples/comparison/main.go](../examples/comparison/main.go).
