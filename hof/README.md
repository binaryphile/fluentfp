# hof

Build complex functions from simple ones — compose, partially apply, or wrap with concurrency control.

```go
// Before: a named function just to combine two transforms
func normalize(s string) string {
    return strings.ToLower(strings.TrimSpace(s))
}

// After: compose directly
normalize := hof.Pipe(strings.TrimSpace, strings.ToLower)
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
```

## hof vs lof

`hof` *builds* functions — it takes functions and returns new functions. `lof` *is* functions — it wraps Go builtins (`len`, `fmt.Println`) as first-class values for use in chains.

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

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/hof) for complete API documentation, the [main README](../README.md) for installation, and [lof](../lof/) for builtin adapters.
