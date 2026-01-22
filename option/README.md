# option: nil-safe optional values

Represent values that may be absent. Options enforce checking before use—nil panics become compile-time errors.

```go
host := config.GetHost().Or("localhost")  // default if absent
```

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/option) for complete API documentation. For the full discussion of nil safety, see [Nil Safety in Go](../nil-safety.md). For function naming patterns, see [Naming Functions for Higher-Order Functions](../naming-in-hof.md).

## Quick Start

```go
import "github.com/binaryphile/fluentfp/option"

// Create options
found := option.Of("hello")           // ok
missing := option.NotOkString         // not-ok

// Extract with defaults
hello := found.Or("fallback")
fallback := missing.Or("fallback")

// Check and extract
if val, ok := found.Get(); ok {
    // use val
}
```

## Types

`Basic[T]` holds an optional value—either "ok" (has value) or "not-ok" (absent).

Type aliases `String`, `Int`, `Bool` are shorthand for `Basic[string]`, `Basic[int]`, etc. See [Type Aliases](#type-aliases) for the full list.

## API Reference

### Constructors

| Function | Signature | Purpose | Example |
|----------|-----------|---------|---------|
| `Of` | `Of[T](T) Basic[T]` | Create ok option | `option.Of(user)` |
| `New` | `New[T](T, bool) Basic[T]` | From value + ok flag | `option.New(val, ok)` |
| `NotOk` | `NotOk[T]() Basic[T]` | Create not-ok option | `option.NotOkString` |
| `IfProvided` | `IfProvided[T comparable](T) Basic[T]` | Not-ok if zero | `option.IfProvided(name)` |
| `IfNotZero` | `IfNotZero[T ZeroChecker](T) Basic[T]` | Not-ok if IsZero() | `option.IfNotZero(time)` |
| `FromOpt` | `FromOpt[T](*T) Basic[T]` | From pointer | `option.FromOpt(ptr)` |
| `Getenv` | `Getenv(string) String` | From env var | `option.Getenv("PORT")` |

### Extraction Methods

| Method | Signature | Purpose | Example |
|--------|-----------|---------|---------|
| `.Get` | `.Get() (T, bool)` | Comma-ok unwrap | `user, ok := option.FromOpt(ptr).Get()` |
| `.IsOk` | `.IsOk() bool` | Check if ok | `if option.IfProvided(name).IsOk()` |
| `.MustGet` | `.MustGet() T` | Value or panic | `user = option.Of(admin).MustGet()` |
| `.Or` | `.Or(T) T` | Value or default | `port = option.Getenv("PORT").Or("8080")` |
| `.OrCall` | `.OrCall(func() T) T` | Lazy default | `port = option.Getenv("PORT").OrCall(findPort)` |
| `.OrZero` | `.OrZero() T` | Value or zero | `name = option.IfProvided(input).OrZero()` |
| `.OrEmpty` | `.OrEmpty() T` | Alias for strings | `name = option.IfProvided(input).OrEmpty()` |
| `.OrFalse` | `.OrFalse() bool` | For option.Bool | `enabled = option.New(flag, ok).OrFalse()` |
| `.ToOpt` | `.ToOpt() *T` | Convert to pointer | `ptr = option.Of(user).ToOpt()` |

### Filtering Methods

| Method | Signature | Purpose | Example |
|--------|-----------|---------|---------|
| `.KeepOkIf` | `.KeepOkIf(func(T) bool) Basic[T]` | Not-ok if false | `active = option.Of(user).KeepOkIf(User.IsActive)` |
| `.ToNotOkIf` | `.ToNotOkIf(func(T) bool) Basic[T]` | Not-ok if true | `valid = option.Of(user).ToNotOkIf(User.IsExpired)` |

### Mapping Methods

| Method | Signature | Purpose | Example |
|--------|-----------|---------|---------|
| `.ToSame` | `.ToSame(func(T) T) Basic[T]` | Transform, same type | `normalized = option.Of(user).ToSame(User.Normalize)` |
| `.ToString` | `.ToString(func(T) string) String` | Transform to string | `name = option.Of(user).ToString(User.Name)` |
| `.ToInt` | `.ToInt(func(T) int) Int` | Transform to int | `age = option.Of(user).ToInt(User.Age)` |
| `Map` | `Map[T,R](Basic[T], func(T)R) Basic[R]` | Transform to any type | `role = option.Map(option.Of(user), User.Role)` |

Other `To[Type]` methods: `ToAny`, `ToBool`, `ToByte`, `ToError`, `ToRune`

### Side Effects

| Method | Signature | Purpose | Example |
|--------|-----------|---------|---------|
| `.Call` | `.Call(func(T))` | Execute if ok | `option.Of(user).Call(User.Save)` |

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

### Advanced: Domain Option Types

For domain-specific behavior (conditional lifecycle management, dependency injection), see the [advanced option example](../examples/advanced_option.go).

## See Also

For typed failure values instead of absent, see [either](../either/).
