# wrap

Resilience decorators for context-aware functions. Wrap a function with retry and circuit breaking — chainable methods, no type arguments needed.

```go
safeFetch := wrap.Func(fetchUser).
    Retry(3, wrap.ExpBackoff(time.Second), nil).
    Breaker(breaker)
```

## What It Looks Like

```go
// Circuit breaker — trips after 5 consecutive failures, resets after 30s
breaker := wrap.NewBreaker(wrap.BreakerConfig{
    ResetTimeout: 30 * time.Second,
    ReadyToTrip:  wrap.ConsecutiveFailures(5),
})
safeFetch := wrap.Func(fetchFromAPI).Breaker(breaker)
resp, err := safeFetch(ctx, url)  // returns wrap.ErrCircuitOpen when tripped
```

```go
// Retry transient errors, then circuit-break the dependency
resilient := wrap.Func(fetchData).
    Retry(3, wrap.ExpBackoff(100*time.Millisecond), isTransient).
    Breaker(breaker)
```

```go
// Error observation and transformation
observed := wrap.Func(fetchUser).OnError(logErr).MapError(annotate)
```

```go
// Custom decorators via Apply
wrap.Func(fn).Retry(3, backoff, nil).Apply(myCustomDecorator)
```

## Methods on Fn[T, R]

| Method | Purpose |
|--------|---------|
| `Breaker(b)` | Circuit breaker — shared `*Breaker` from `NewBreaker(cfg)` |
| `MapError(mapper)` | Transform errors via mapping function |
| `OnError(handler)` | Side-effect handler on error |
| `Retry(max, backoff, pred)` | Retry on error with backoff strategy |
| `Apply(ds...)` | Apply custom `Decorator` values |

## Backoff

`ExpBackoff(initial)` — randomized exponential: uniform random in [0, initial * 2^n). Spreads retries to minimize collisions under contention.

## Circuit Breaker

- `NewBreaker(cfg) *Breaker` — 3-state: closed → open → half-open → closed
- `ConsecutiveFailures(n) func(Snapshot) bool` — ReadyToTrip predicate
- `ErrCircuitOpen` — sentinel error when breaker rejects

All context-aware decorators return `ctx.Err()` on cancellation. Circuit breaker does not count `context.Canceled` as a failure.

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/wrap) for complete API documentation.
