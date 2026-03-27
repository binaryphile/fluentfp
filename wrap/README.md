# wrap

Resilience decorators for context-aware functions. Wrap a function with retry, circuit breaking, throttling, and error handling — all configured in one struct.

```go
safeFetch := wrap.Func(fetchUser).With(wrap.Features{
    Breaker:  breaker,
    Retry:    wrap.Retry(3, wrap.ExpBackoff(time.Second), isTransient),
    Throttle: wrap.Throttle(10),
})
```

The library controls decorator order (OnError → MapError → Retry → Breaker → Throttle). Nil fields are skipped.

## What It Looks Like

```go
// Circuit breaker — trips after 5 consecutive failures, resets after 30s
breaker := wrap.NewBreaker(wrap.BreakerConfig{
    ResetTimeout: 30 * time.Second,
    ReadyToTrip:  wrap.ConsecutiveFailures(5),
})
safeFetch := wrap.Func(fetchFromAPI).WithBreaker(breaker)
resp, err := safeFetch(ctx, url)  // returns wrap.ErrCircuitOpen when tripped
```

```go
// Chain convenience methods for single decorators
resilient := wrap.Func(fetchData).
    WithRetry(3, wrap.ExpBackoff(100*time.Millisecond), isTransient).
    WithBreaker(breaker).
    WithThrottle(10)
```

## Features Struct

| Field | Type | Factory | Skip value |
|-------|------|---------|------------|
| Breaker | `*Breaker` | `NewBreaker(cfg)` | nil |
| MapError | `func(error) error` | — | nil |
| OnError | `func(error)` | — | nil |
| Retry | `*RetryConfig` | `Retry(max, backoff, pred)` | nil |
| Throttle | `*ThrottleConfig` | `Throttle(n)` | nil |

`WithThrottleWeighted(capacity, cost)` is available as a method only — the cost function requires type `T`.

## Operations

**Circuit Breaking**
- `NewBreaker(cfg BreakerConfig) *Breaker` — 3-state: closed → open → half-open → closed
- `ConsecutiveFailures(n) func(Snapshot) bool` — ReadyToTrip predicate
- `ErrCircuitOpen` — sentinel error when breaker rejects

**Retry**
- `Retry(max, backoff, shouldRetry) *RetryConfig` — factory for Features
- `ExpBackoff(initial) Backoff` — randomized exponential: random in [0, initial * 2^n)

**Concurrency Control**
- `Throttle(n) *ThrottleConfig` — factory for Features (count-based)
- `WithThrottleWeighted(capacity, cost)` — method only (cost-based)

**Error Handling**
- `MapError` field — transform errors via mapping function
- `OnError` field — side-effect handler on error

All context-aware decorators return `ctx.Err()` on cancellation. Circuit breaker does not count `context.Canceled` as a failure. All functions panic on nil inputs.

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/wrap) for complete API documentation.
