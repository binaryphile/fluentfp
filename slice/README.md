# slice: fluent collection operations

Eliminate loop bugs with type-safe collection operations. Method expressions make pipelines read like intent—**fluent** operations chain method calls on a single line, no intermediate variables, no loop scaffolding.

```go
actives := slice.From(users).KeepIf(User.IsActive).ToString(User.GetName)
```

See the [main README](../README.md) for when to use fluentfp and performance characteristics. See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/slice) for complete API documentation. For function naming patterns, see [Naming Functions for Higher-Order Functions](../naming-in-hof.md).

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

## Types

`Mapper[T]` wraps a slice for fluent operations. Create with `From()`, chain methods, use as a regular slice.

`MapperTo[R,T]` adds `.To()` for mapping to arbitrary type R. Create with `MapTo[R]()`.

```go
users := slice.From(rawUsers)       // Mapper[User]
names := users.ToString(User.Name)  // Mapper[string]
```

## API Reference

### Factory Functions

| Function | Signature | Purpose | Example |
|----------|-----------|---------|---------|
| `From` | `From[T]([]T) Mapper[T]` | Create Mapper from slice | `slice.From(users)` |
| `MapTo` | `MapTo[R,T]([]T) MapperTo[R,T]` | Create MapperTo for type R | `slice.MapTo[User](ids)` |

### Mapper Methods

| Method | Signature | Purpose | Example |
|--------|-----------|---------|---------|
| `KeepIf` | `.KeepIf(func(T) bool) Mapper[T]` | Keep matching | `actives = slice.From(users).KeepIf(User.IsActive)` |
| `RemoveIf` | `.RemoveIf(func(T) bool) Mapper[T]` | Remove matching | `current = slice.From(users).RemoveIf(User.IsExpired)` |
| `TakeFirst` | `.TakeFirst(n int) Mapper[T]` | First n elements | `top10 = slice.From(users).TakeFirst(10)` |
| `Convert` | `.Convert(func(T) T) Mapper[T]` | Map to same type | `normalized = slice.From(users).Convert(User.Normalize)` |
| `ToString` | `.ToString(func(T) string) Mapper[string]` | Map to string | `names = slice.From(users).ToString(User.Name)` |
| `ToInt` | `.ToInt(func(T) int) Mapper[int]` | Map to int | `ages = slice.From(users).ToInt(User.Age)` |
| `Each` | `.Each(func(T))` | Side-effect iteration | `slice.From(users).Each(User.Save)` |
| `Len` | `.Len() int` | Count elements | `count = slice.From(users).Len()` |

Other `To[Type]` methods: `ToAny`, `ToBool`, `ToByte`, `ToError`, `ToFloat32`, `ToFloat64`, `ToInt32`, `ToInt64`, `ToRune`

### MapperTo Additional Method

| Method | Signature | Purpose | Example |
|--------|-----------|---------|---------|
| `To` | `.To(func(T) R) Mapper[R]` | Map to type R | `users = slice.MapTo[User](ids).To(FetchUser)` |

### Standalone Functions

| Function | Signature | Purpose | Example |
|----------|-----------|---------|---------|
| `Fold` | `Fold[T,R]([]T, R, func(R,T) R) R` | Reduce to single value | See [Fold](#fold) |
| `Unzip2` | `Unzip2[T,A,B]([]T, func(T)A, func(T)B) (Mapper[A], Mapper[B])` | Extract 2 fields | `names, ages = slice.Unzip2(users, User.Name, User.Age)` |
| `Unzip3` | `Unzip3[T,A,B,C](...)` | Extract 3 fields | — |
| `Unzip4` | `Unzip4[T,A,B,C,D](...)` | Extract 4 fields | — |

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
sumFloat64 := func(acc, n float64) float64 { return acc + n }
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

## When NOT to Use slice

- **Early exit needed** — `KeepIf` iterates entire slice; use loop with `break` for first match
- **Accumulating into maps** — No fluent equivalent; use `Fold` or a loop
- **Performance-critical hot paths** — Profile first; each chain operation allocates
- **Single simple operation** — `for _, u := range users` may be clearer than `slice.From(users).Each(...)`

## When Loops Are Necessary

- **Channel consumption**: `for r := range ch`
- **Complex control flow**: break, continue, early return
- **Index-dependent logic**: comparing adjacent elements, position-aware output

## See Also

- For zipping slices together, see [pair](../tuple/pair/)
- For library comparison, see [comparison.md](../comparison.md)
