# option

Optional values that make absence explicit in the type.

`Option[T]` stores a value plus an `ok` flag. The zero value is not-ok, so it works without initialization.

```go
// Before: four lines to extract a map value with a default
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
port := option.Env("PORT").Or("8080")
```

```go
// Conditional pipeline
name := userOption.KeepIf(User.IsActive).ToString(User.Name).Or("unknown")
```

```go
// Comma-ok extraction
if user, ok := userOption.Get(); ok {
    fmt.Println(user.Name())
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
    Port: option.Env("PORT").Or("8080"),
    Name: option.Lookup(labels, "name").Or("default"),
}
```

```go
// Tri-state boolean — option.Bool is Option[bool]
type ScanResult struct {
    IsConnected option.Bool  // true, false, or unknown
}
connected := option.OrFalse(rslt.IsConnected)  // unknown → false
```

```go
// Nullable field — return option instead of zero value
func (r Record) Host() option.String {
    return option.NonZero(r.host)
}
// Caller decides how to handle absence
addr := record.Host().Or("localhost")
```

```go
// Pre-declared not-ok values read as intent in returns
func (db *DB) FindUser(id int) option.Int {
    row := db.QueryRow("SELECT id FROM users WHERE id = ?", id)
    var uid int
    if err := row.Scan(&uid); err != nil {
        return option.NotOkInt
    }
    return option.Of(uid)
}
```

## One Type for All of Go's "Maybe" Patterns

Go represents absence three different ways: `*T` (nil), zero values (`""`, `0`), and comma-ok returns (`map` lookup, type assertion). Each has a different failure mode — nil derefs panic, zero values are silently ambiguous, and ignored `ok` values lose the distinction between missing and present.

`Option[T]` unifies them. Factory functions bridge each Go pattern into a single chainable type:

- `NonNil(ptr)` — pointer-based absence
- `NonZero(count)`, `NonEmpty(name)` — zero-value absence (use only when zero/empty truly means absent in your domain)
- `Lookup(m, key)`, `New(val, ok)` — comma-ok absence
- `When(cond, val)`, `WhenCall(cond, fn)` — conditional construction (eager/lazy)
- `Env("PORT")` — environment variable absence (unset or empty)

Once you have an `Option[T]`, the same API works regardless of where the value came from: `.Or("default")`, `.KeepIf(valid)`, `.ToString(format)`, `.Get()`.

## Operations

`Option[T]` holds an optional value — ok or not-ok. Type aliases `String`, `Int`, `Bool`, etc. are shorthand for common types, with pre-declared not-ok values (`NotOkString`, `NotOkInt`, etc.). JSON serialization via `MarshalJSON`/`UnmarshalJSON` (ok → value, not-ok → null). SQL via `Value`/`Scan` (ok → value, not-ok → NULL). Note: both JSON and SQL collapse Ok(nil) and NotOk into the same representation (null/NULL) — a round-trip through serialization may lose the distinction.

- **Create**: `Of`, `New`, `When`, `WhenCall`, `NotOk`, `NonZero`, `NonEmpty`, `NonNil`, `NonErr`, `Env`, `Lookup`
- **Create + Transform**: `NonZeroCall`, `NonEmptyCall`, `NonNilCall` — check presence and apply fn in one step
- **Bridge**: `OkOr` (eager), `OkOrCall` (lazy) — convert `Option[T]` to `rslt.Result[T]`, treating absence as an error; `FlatMapResult` — apply a fallible function, absent → Ok(NotOk), present+invalid → Err
- **Extract**: `Get`, `IsOk`, `MustGet`, `Or`, `OrCall`, `OrElse`, `OrZero`, `OrEmpty`, `OrFalse` (standalone for `Option[bool]`)
- **Transform**: `Convert` (same type), `Map` (cross-type, standalone), `ToString`, `ToInt`, other `To*`, `ToOpt`
- **Filter**: `KeepIf`, `RemoveIf`
- **Side effects**: `IfOk`, `IfNotOk`, `Lift`

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/option) for complete API documentation, the [main README](../README.md) for installation, [Nil Safety in Go](../nil-safety.md) for the full discussion, and the [showcase](../docs/showcase.md) for real-world rewrites.
