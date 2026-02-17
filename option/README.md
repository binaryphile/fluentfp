# option

Optional values that move absence checks from runtime to the type system.

```go
// Before: six lines to safely extract a map value with a default
var token string
if t, ok := headers["Authorization"]; ok {
    token = t
} else {
    token = "none"
}

// After
token := option.Lookup(headers, "Authorization").Or("none")
```

## What It Looks Like

```go
// Environment with default
port := option.Getenv("PORT").Or("8080")
```

```go
// Conditional pipeline
name := userOption.KeepOkIf(User.IsActive).ToString(User.Name).Or("unknown")
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
// Tri-state boolean — option.Bool is Basic[bool]
type ScanResult struct {
    IsConnected option.Bool  // true, false, or unknown
}
connected := result.IsConnected.OrFalse()  // unknown → false
```

## One Type for All of Go's "Maybe" Patterns

Go represents absence three different ways: `*T` (nil), zero values (`""`, `0`), and comma-ok returns (`map` lookup, type assertion). All three let you skip the check and use the value directly — the failure shows up at runtime, not compile time.

`Basic[T]` unifies them. Factory functions bridge each Go pattern into a single chainable type:

- `IfNotNil(ptr)` — pointer-based absence
- `IfNotZero(count)`, `IfNotEmpty(name)` — zero-value absence
- `Lookup(m, key)`, `New(val, ok)` — comma-ok absence
- `Getenv("PORT")` — environment variable absence

Once you have a `Basic[T]`, the same API works regardless of where the value came from: `.Or("default")`, `.KeepOkIf(valid)`, `.ToString(format)`, `.Get()`.

## Operations

`Basic[T]` holds an optional value — ok or not-ok. Type aliases `String`, `Int`, `Bool`, etc. are shorthand for common types, with pre-declared not-ok values (`NotOkString`, `NotOkInt`, etc.). JSON serialization via `MarshalJSON`/`UnmarshalJSON` (ok → value, not-ok → null).

- **Create**: `Of`, `New`, `NotOk`, `IfNotZero`, `IfNotEmpty`, `IfNotNil`, `Getenv`, `Lookup`
- **Extract**: `Get`, `IsOk`, `MustGet`, `Or`, `OrCall`, `OrZero`, `OrEmpty`, `OrFalse`
- **Transform**: `Convert`, `Map`, `ToString`, `ToInt`, other `To*`, `ToOpt`
- **Filter**: `KeepOkIf`, `ToNotOkIf`
- **Side effects**: `IfOk`, `IfNotOk`, `Lift`

For domain-specific option types with conditional method dispatch, see the [advanced option example](../examples/advanced_option.go).

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/option) for complete API documentation, the [main README](../README.md) for installation, and [Nil Safety in Go](../nil-safety.md) for the full discussion.
