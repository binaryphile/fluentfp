# must: panic-on-error for invariants

Convert fallible operations to panics when errors indicate bugs, not runtime conditions. **If you're tempted to ignore an error with `_ =`, use `must` instead.**

An **invariant** is a precondition that, if violated, means the program has a bug—not a runtime error to handle.

```go
db := must.Get(sql.Open("postgres", dsn))
must.BeNil(db.Ping())
```

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/must) for complete API documentation.

## Quick Start

```go
import "github.com/binaryphile/fluentfp/must"

// Extract value or panic
db := must.Get(sql.Open("postgres", dsn))

// Panic if error non-nil
must.BeNil(db.Ping())

// Environment variable or panic
home := must.Getenv("HOME")

// Wrap fallible func for HOF use
mustAtoi := must.Of(strconv.Atoi)
ints := slice.From(strings).ToInt(mustAtoi)
```

## API Reference

| Function | Signature | Purpose | Example |
|----------|-----------|---------|---------|
| `Get` | `Get[T](t T, err error) T` | Value or panic | `must.Get(os.Open(path))` |
| `Get2` | `Get2[T,T2](t T, t2 T2, err error) (T, T2)` | Two-value variant | `must.Get2(fn())` |
| `BeNil` | `BeNil(err error)` | Panic if non-nil | `must.BeNil(db.Ping())` |
| `Getenv` | `Getenv(key string) string` | Env var or panic | `must.Getenv("HOME")` |
| `Of` | `Of[T,R](fn func(T)(R,error)) func(T)R` | Wrap for HOF use | `must.Of(strconv.Atoi)` |

## Naming Convention

When storing a must-wrapped function in a variable, prefix with `must` to signal panic behavior:

```go
mustAtoi := must.Of(strconv.Atoi)
ints := slice.From(strings).ToInt(mustAtoi)
```

This makes panic behavior visible at the call site.

For more naming patterns, see [Naming Functions for Higher-Order Functions](../naming-in-hof.md).

## Recovering from Panics

When using `must.Of` in pipelines, panics can be caught with defer/recover:

```go
func SafeParseAll(inputs []string) (results []int, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("parse failed: %v", r)
        }
    }()

    mustAtoi := must.Of(strconv.Atoi)
    results = slice.From(inputs).ToInt(mustAtoi)
    return results, nil
}
```

This pattern converts pipeline panics back to errors at the boundary. Use it when:
- Processing untrusted input that may contain invalid values
- You want FP-style pipelines but need error handling at the edges
- The calling code expects errors, not panics

## Patterns

### Initialization Sequences

Use `must.Get` and `must.BeNil` together for init/setup code where errors are fatal:

```go
func OpenDatabase(cfg Config) *Database {
    db := must.Get(sql.Open("postgres", cfg.DSN))
    must.BeNil(db.Ping())

    db.SetMaxIdleConns(cfg.IdleConns)
    db.SetMaxOpenConns(cfg.OpenConns)

    return &Database{db: db}
}
```

### Config Loading

Combine with config parsing where missing config is fatal:

```go
func LoadConfig() Config {
    var config Config
    must.BeNil(viper.ReadInConfig())
    must.BeNil(viper.Unmarshal(&config))
    return config
}
```

### Embedding Files

Use with `embed.FS` when embedded files must exist:

```go
//go:embed README.md
var files embed.FS

func GetReadme() string {
    return string(must.Get(files.ReadFile("README.md")))
}
```

### Other Common Uses

```go
// Time parsing with known-valid formats
timestamp := must.Get(time.Parse("2006-01-02 15:04:05", s.ScannedAt))

// Validation-only (discard result)
_ = must.Get(strconv.Atoi(configID))  // panics if invalid
```

## When NOT to Use must

- **Recoverable errors** — If the caller can handle it, return `error`
- **User input** — Never panic on external input; validate and return errors
- **Library code** — Libraries should return errors, not panic
- **Production request handlers** — One bad request shouldn't crash the server
- **Expected failures** — Network timeouts, file not found, etc. are expected; handle them

## Never Ignore Errors

When you write `_ = doSomething()`, you're implicitly declaring an invariant: "this won't error." But what if it does?

```go
_ = os.Setenv("PATH", newPath)  // Silent failure, program continues corrupted
```

The program continues in an invalid state. You've traded a clear failure for silent corruption.

`must` makes the invariant explicit:

```go
must.BeNil(os.Setenv("PATH", newPath))  // Invariant enforced, clear failure
```

**Every `_ = fn()` that discards an error should be `must.BeNil()` or `must.Get()`.** If you're confident the error "won't happen," prove it—by panicking if it does.

Tools like `errcheck` flag ignored errors as bugs.

## See Also

For typed error handling without panics, see [either](../either/).
