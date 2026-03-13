# hof

Build complex functions from simple ones — compose, partially apply, or wrap with concurrency control.

```go
// Compose at the call site when the combination is obvious
normalize := hof.Pipe(strings.TrimSpace, strings.ToLower)
slice.From(inputs).Convert(normalize)
```

## What It Looks Like

```go
// Partial application — fix one argument
add := func(a, b int) int { return a + b }
add5 := hof.Bind(add, 5)
slice.From(nums).Convert(add5)
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
// Bound concurrency — at most 5 in-flight API calls
callAPI := hof.Throttle(5, fetchFromAPI)
resp, err := callAPI(ctx, url)  // blocks until a slot opens
```

```go
// Bound by total cost — large items consume more budget
fetchData := hof.ThrottleWeighted(100, estimateSize, fetchFromAPI)
```

```go
// Cancel remaining work on first error
ctx, cancel := context.WithCancel(parentCtx)
defer cancel()
failFast := hof.OnErr(fetchURL, cancel)
```

```go
// Retry with exponential backoff — composable with Throttle
resilientFetch := hof.Retry(3, hof.ExponentialBackoff(100*time.Millisecond), fetchFromAPI)
resp, err := resilientFetch(ctx, url)
```

## hof vs lof

`hof` *builds* functions — it takes functions and returns new functions. `lof` *is* functions — it wraps Go builtins and standard library functions (`len`, `fmt.Println`) as first-class values for use in chains.

`hof.Pipe` builds a transform; `lof.Len` is a transform.

## Operations

**Composition**
- `Pipe[A, B, C](f func(A) B, g func(B) C) func(A) C` — left-to-right: `Pipe(f, g)(x)` = `g(f(x))`

**Partial Application**
- `Bind[A, B, C](f func(A, B) C, a A) func(B) C` — fix first arg
- `BindR[A, B, C](f func(A, B) C, b B) func(A) C` — fix second arg

**Independent Application**
- `Cross[A, B, C, D](f func(A) C, g func(B) D) func(A, B) (C, D)` — apply separate fns to separate args

**Building Blocks**
- `Eq[T comparable](target T) func(T) bool` — equality predicate factory

**Concurrency Control**
- `Throttle[T, R](n int, fn func(context.Context, T) (R, error)) func(context.Context, T) (R, error)` — bound by call count
- `ThrottleWeighted[T, R](capacity int, cost func(T) int, fn func(context.Context, T) (R, error)) func(context.Context, T) (R, error)` — bound by total cost

**Side-Effect Wrappers**
- `OnErr[T, R](fn func(context.Context, T) (R, error), onErr func()) func(context.Context, T) (R, error)` — call handler on error

**Retry**
- `Retry[T, R](maxAttempts int, backoff Backoff, fn) func(context.Context, T) (R, error)` — retry on error with pluggable backoff
- `ConstantBackoff(delay time.Duration) Backoff` — fixed delay between retries
- `ExponentialBackoff(initial time.Duration) Backoff` — full jitter: random in [0, initial * 2^n)

All functions panic on nil inputs. `Throttle`, `ThrottleWeighted`, and `Retry` panic on non-positive limits. `ThrottleWeighted` also panics per-call if `cost` returns a non-positive value or one exceeding capacity. `ExponentialBackoff` panics if initial <= 0.

`Throttle` and `ThrottleWeighted` return functions that are safe for concurrent use from multiple goroutines. All context-aware wrappers (`Throttle`, `ThrottleWeighted`, `Retry`) return `ctx.Err()` on cancellation rather than blocking indefinitely.

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/hof) for complete API documentation, the [main README](../README.md) for installation, and [lof](../lof/) for builtin adapters.
