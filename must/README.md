# must

Turn assumed preconditions into enforced invariants. Panics when violated.

Use `must` when failure indicates a programmer error or misconfiguration — not for recoverable runtime conditions.

```go
// Before: you're assuming this won't fail — but silently
_ = os.Setenv("PATH", newPath)

// After: same assumption, now enforced
err := os.Setenv("PATH", newPath)
must.BeNil(err)
```

`must.BeNil` replaces the three-line `if err != nil { panic(err) }` block with one line. `must.Get` goes further — four lines (declare, call, check, panic) become one.

## What It Looks Like

```go
// Initialization — errors here mean a bug or misconfiguration, not a runtime condition
re := must.Get(regexp.Compile(pattern))
port := must.Get(strconv.Atoi(os.Getenv("PORT")))
```

```go
// Required environment
home := must.NonEmptyEnv("HOME")
```

```go
// Pipeline adapter — wraps func(T)(R, error) into func(T) R
mustAtoi := must.Of(strconv.Atoi)
n := mustAtoi("42")  // 42, or panics if not a valid integer
```

## Making Assumptions Visible

Every `_ = fn()` is a hidden assumption — you're expecting the error won't happen. If it does, the program continues with no indication of what went wrong.

`must` makes the assumption explicit. If it's wrong, you find out immediately.

Don't recover from `must` panics and continue normal execution. An invariant violation means the program is in a state you didn't anticipate — recovering and continuing from that state is how silent corruption happens. Top-level recovery for logging, crash reporting, or test assertions is fine. If you're tempted to recover and continue, the operation isn't an invariant — use `(T, error)` returns and handle the error explicitly.

**Convention:** For error-only returns, keep the `err` assignment — it reads like Go:

```go
err := os.MkdirAll(dataDir, 0o755)
must.BeNil(err)
```

Use `must.Get` when you need the value: `re := must.Get(regexp.Compile(pattern))`. Prefix `must.Of`-wrapped variables with `must` to signal panic behavior: `mustAtoi := must.Of(strconv.Atoi)`.

## Operations

- `Get(T, error) T` — extract value or panic with the error
- `Get2(T, T2, error) (T, T2)` — two-value variant
- `BeNil(error)` — panic with the error if non-nil
- `NonEmptyEnv(key) string` — env var or panic if unset or empty (wraps `ErrEnvUnset` / `ErrEnvEmpty`)
- `Of(func(T)(R, error)) func(T) R` — wrap for higher-order use (panics immediately if fn is nil, wraps `ErrNilFunction`)

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/must) for complete API documentation, the [main README](../README.md) for installation, and [rslt](../rslt/) for typed error handling without panics.
