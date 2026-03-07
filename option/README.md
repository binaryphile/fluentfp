# option

Optional values that move absence checks from runtime to the type system.

```go
// Before: four lines to safely extract a map value with a default
token, ok := headers["Authorization"]
if !ok {
    token = "none"
}

// After
token := option.Lookup(headers, "Authorization").Or("none")
```

Four lines become one.

## What It Looks Like

```go
// Environment with default
port := option.Getenv("PORT").Or("8080")
```

```go
// Conditional pipeline
name := userOption.KeepIf(User.IsActive).ToString(User.Name).Or("unknown")
```

```go
// Comma-ok extraction
if user, ok := userOption.Get(); ok {
    fmt.Println(user.Name)
}
```

```go
// Side effect — fires only if ok
userOption.IfOk(User.Save)
```

```go
// Before: three separate absence checks, then assemble
host := record.RawHost()
if host == "" {
    host = "localhost"
}
port := os.Getenv("PORT")
if port == "" {
    port = "8080"
}
name, ok := labels["name"]
if !ok {
    name = "default"
}
return Config{Host: host, Port: port, Name: name}

// After: every field resolves inline
return Config{
    Host: record.Host().Or("localhost"),
    Port: option.Getenv("PORT").Or("8080"),
    Name: option.Lookup(labels, "name").Or("default"),
}
```

```go
// Tri-state boolean — option.Bool is Option[bool]
type ScanResult struct {
    IsConnected option.Bool  // true, false, or unknown
}
connected := result.IsConnected.OrFalse()  // unknown → false
```

```go
// Nullable field — return option instead of zero value
func (r Record) Host() option.String {
    return option.NonZero(r.host)
}
// Caller decides how to handle absence
addr := record.Host().Or("localhost")
```

## One Type for All of Go's "Maybe" Patterns

Go represents absence three different ways: `*T` (nil), zero values (`""`, `0`), and comma-ok returns (`map` lookup, type assertion). All three let you skip the check and use the value directly — the failure shows up at runtime, not compile time.

`Option[T]` unifies them. Factory functions bridge each Go pattern into a single chainable type:

- `NonNil(ptr)` — pointer-based absence
- `NonZero(count)`, `NonEmpty(name)` — zero-value absence
- `Lookup(m, key)`, `New(val, ok)` — comma-ok absence
- `Getenv("PORT")` — environment variable absence

Once you have a `Option[T]`, the same API works regardless of where the value came from: `.Or("default")`, `.KeepIf(valid)`, `.ToString(format)`, `.Get()`.

## Operations

`Option[T]` holds an optional value — ok or not-ok. Type aliases `String`, `Int`, `Bool`, etc. are shorthand for common types, with pre-declared not-ok values (`NotOkString`, `NotOkInt`, etc.). JSON serialization via `MarshalJSON`/`UnmarshalJSON` (ok → value, not-ok → null).

- **Create**: `Of`, `New`, `NotOk`, `NonZero`, `NonEmpty`, `NonNil`, `Getenv`, `Lookup`
- **Create + Transform**: `NonZeroCall`, `NonEmptyCall`, `NonNilCall` — check presence and apply fn in one step
- **Extract**: `Get`, `IsOk`, `MustGet`, `Or`, `OrCall`, `OrZero`, `OrEmpty`, `OrFalse`
- **Transform**: `Convert`, `Map`, `ToString`, `ToInt`, other `To*`, `ToOpt`
- **Filter**: `KeepIf`, `RemoveIf`
- **Side effects**: `IfOk`, `IfNotOk`, `Lift`

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/option) for complete API documentation, the [main README](../README.md) for installation, [Nil Safety in Go](../nil-safety.md) for the full discussion, and the [showcase](../docs/showcase.md) for real-world rewrites.
