# option: nil-safe optional values

Represent values that may be absent. Options enforce checking before use—nil panics become compile-time errors.

```go
host := config.GetHost().Or("localhost")  // default if absent
```

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/option) for complete API documentation. For the full discussion of nil safety, see [Nil Safety in Go](../nil-safety.md).

## Quick Start

```go
import "github.com/binaryphile/fluentfp/option"

// Create options
found := option.Of("hello")           // ok option
missing := option.NotOk[string]()     // not-ok option

// Extract with defaults
value := found.Or("default")          // "hello"
value := missing.Or("default")        // "default"

// Check and extract
if val, ok := found.Get(); ok {
    // use val
}
```

## API Reference

### Constructors

| Function | Signature | Purpose |
|----------|-----------|---------|
| `Of` | `Of[T](T) Basic[T]` | Create ok option |
| `New` | `New[T](T, bool) Basic[T]` | Create from value + ok flag |
| `NotOk` | `NotOk[T]() Basic[T]` | Create not-ok option |
| `IfProvided` | `IfProvided[T comparable](T) Basic[T]` | Not-ok if zero value |
| `IfNotZero` | `IfNotZero[T ZeroChecker](T) Basic[T]` | Not-ok if `t.IsZero()` |
| `FromOpt` | `FromOpt[T](*T) Basic[T]` | From pointer (nil = not-ok) |
| `Getenv` | `Getenv(string) String` | From env var (empty = not-ok) |

### Extraction Methods

| Method | Signature | Purpose |
|--------|-----------|---------|
| `.Get` | `.Get() (T, bool)` | Comma-ok unwrap |
| `.IsOk` | `.IsOk() bool` | Check if ok |
| `.MustGet` | `.MustGet() T` | Value or panic |
| `.Or` | `.Or(T) T` | Value or default |
| `.OrCall` | `.OrCall(func() T) T` | Value or lazy default |
| `.OrZero` | `.OrZero() T` | Value or zero |
| `.OrEmpty` | `.OrEmpty() T` | Alias for strings |
| `.OrFalse` | `.OrFalse() bool` | For `option.Bool` |
| `.ToOpt` | `.ToOpt() *T` | Convert to pointer |

### Filtering Methods

| Method | Signature | Purpose |
|--------|-----------|---------|
| `.KeepOkIf` | `.KeepOkIf(func(T) bool) Basic[T]` | Not-ok if predicate false |
| `.ToNotOkIf` | `.ToNotOkIf(func(T) bool) Basic[T]` | Not-ok if predicate true |

### Mapping Methods

| Method | Signature | Purpose |
|--------|-----------|---------|
| `.ToSame` | `.ToSame(func(T) T) Basic[T]` | Transform, same type |
| `.ToString` | `.ToString(func(T) string) String` | Transform to string |
| `.ToInt` | `.ToInt(func(T) int) Int` | Transform to int |
| `Map` | `Map[T,R](Basic[T], func(T)R) Basic[R]` | Transform to any type (function, not method) |

Other `To[Type]` methods: `ToAny`, `ToBool`, `ToByte`, `ToError`, `ToRune`

### Side Effects

| Method | Signature | Purpose |
|--------|-----------|---------|
| `.Call` | `.Call(func(T))` | Execute if ok |

### Type Aliases

Pre-defined types: `String`, `Int`, `Bool`, `Error`, `Any`, `Byte`, `Rune`

Pre-defined not-ok values: `NotOkString`, `NotOkInt`, `NotOkBool`, `NotOkError`, `NotOkAny`, `NotOkByte`, `NotOkRune`

## Creating Options

```go
// From known value
opt := option.Of("hello")

// Conditional creation
opt := option.New(value, ok)

// From comparable (not-ok if zero)
opt := option.IfProvided(maybeEmpty)

// From pointer (not-ok if nil)
opt := option.FromOpt(ptr)

// From environment
port := option.Getenv("PORT").Or("8080")
```

### Non-Comparable Types

For types containing slices, maps, or funcs, implement `ZeroChecker`:

```go
type Registry struct {
    instances map[string]Instance
}

func (r Registry) IsZero() bool { return r.instances == nil }

opt := option.IfNotZero(registry)  // not-ok if IsZero() returns true
```

## Using Options

```go
// Get with comma-ok
if val, ok := opt.Get(); ok {
    // use val
}

// Get with default
val := opt.Or("default")
val := opt.OrCall(expensiveDefault)
val := opt.OrZero()

// Transform
length := opt.ToInt(func(s string) int { return len(s) })
upper := option.Map(opt, strings.ToUpper)

// Filter
nonEmpty := opt.KeepOkIf(func(s string) bool { return s != "" })

// Side effect
opt.Call(fmt.Println)
```

## When NOT to Use option

- **Go idiom `(T, error)`** — Don't replace error returns with option
- **Performance-critical paths** — Option adds a bool field; profile first
- **Simple nil checks** — If `if ptr != nil` is clear, don't over-engineer
- **When error context matters** — Option loses why something is absent

## Patterns

### Tri-State Boolean

```go
type ScanResult struct {
    IsConnected option.Bool  // true, false, or unknown
}

connected := result.IsConnected.OrFalse()  // unknown → false
```

### Nullable Database Fields

```go
func (r Record) GetHost() option.String {
    return option.IfProvided(r.NullableHost.String)
}
```
