# hof

Pure function combinators — compose, partially apply, or compare.

```go
normalize := hof.Pipe(strings.TrimSpace, strings.ToLower)
slice.From(inputs).Transform(normalize)
```

For decorators over context-aware calls (retry, circuit breaker, throttle), see the [call](../call/) package.

## What It Looks Like

```go
// Compose two functions left-to-right
normalize := hof.Pipe(strings.TrimSpace, strings.ToLower)
```

```go
// Partial application — fix one argument
add := func(a, b int) int { return a + b }
add5 := hof.Bind(add, 5)
slice.From(nums).Transform(add5)
```

```go
// Apply separate functions to separate values
both := hof.Cross(double, toUpper)
d, u := both(5, "hello")  // 10, "HELLO"
```

```go
// Equality predicate for Every/Any
allSkipped := slice.From(statuses).Every(hof.Eq(Skipped))
```

```go
// Debounce — coalesce rapid calls, execute once after quiet period
d := hof.NewDebouncer(500*time.Millisecond, saveConfig)
defer d.Close()
d.Call(cfg)
```

## Operations

**Composition**
- `Pipe[A, B, C](f func(A) B, g func(B) C) func(A) C` — left-to-right composition
- `Bind[A, B, C](f func(A, B) C, a A) func(B) C` — fix first arg
- `BindR[A, B, C](f func(A, B) C, b B) func(A) C` — fix second arg
- `Cross[A, B, C, D](f func(A) C, g func(B) D) func(A, B) (C, D)` — apply separate fns to separate args
- `Eq[T comparable](target T) func(T) bool` — equality predicate factory

**Debounce**
- `NewDebouncer[T](wait, fn, opts...) *Debouncer[T]` — trailing-edge coalescer
- `MaxWait(d) DebounceOption` — cap maximum deferral

All functions panic on nil inputs.

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/hof) for complete API documentation and the [main README](../README.md) for installation.
