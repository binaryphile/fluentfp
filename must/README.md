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

When you do handle the error, `must.BeNil` replaces the three-line `if err != nil` block with one. `must.Get` goes further — four lines (declare, call, check, panic) become one.

## What It Looks Like

```go
// Initialization — errors here mean a bug, not a runtime condition
db := must.Get(sql.Open("postgres", cfg.DSN))
err := db.Ping()
must.BeNil(err)
```

```go
// Required environment
home := must.Getenv("HOME")
```

```go
// Pipeline adapter — wraps func(T)(R, error) into func(T) R
mustAtoi := must.Of(strconv.Atoi)
ints := slice.From(strings).ToInt(mustAtoi)
```

## Making Assumptions Visible

Every `_ = fn()` is a hidden invariant — you're assuming the error won't happen. If it does, the program continues in a corrupt state with no indication of what went wrong.

`must` makes the assumption explicit. If it's wrong, you find out immediately.

Don't recover from `must` panics. An invariant violation means the program is in a state you didn't anticipate — recovering and continuing from that state is how silent corruption happens. If you're tempted to recover, the operation isn't an invariant — use `(T, error)` returns and handle the error explicitly.

**Convention:** For error-only returns, keep the `err` assignment — it reads like Go:

```go
err := db.Ping()
must.BeNil(err)
```

Use `must.Get` when you need the value: `db := must.Get(sql.Open(...))`. Prefix `must.Of`-wrapped variables with `must` to signal panic behavior: `mustAtoi := must.Of(strconv.Atoi)`.

## Operations

- `Get(T, error) T` — extract value or panic
- `Get2(T, T2, error) (T, T2)` — two-value variant
- `BeNil(error)` — panic if non-nil
- `Getenv(key) string` — env var or panic
- `Of(func(T)(R, error)) func(T) R` — wrap for higher-order use

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/must) for complete API documentation, the [main README](../README.md) for installation, and [either](../either/) for typed alternatives without panics.
