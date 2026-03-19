# call

Resilience decorators for communicating with runtime dependencies. Named after effectful call decorators — communication over unreliable channels. "Breaker, breaker."

All decorators wrap `func(context.Context, T) (R, error)` and return the same signature, so they compose by stacking:

```go
// Classify → retry → circuit break — each wraps the previous
classified := call.MapErr(fetchUser, classifyError)
retried := call.Retry(3, call.ExponentialBackoff(100*time.Millisecond), isTransient, classified)
safeFetch := call.WithBreaker(breaker, retried)

// Caller sees the same signature as fetchUser
resp, err := safeFetch(ctx, url)
```

## What It Looks Like

```go
// Retry with exponential backoff, only for transient errors
backoff := call.ExponentialBackoff(100 * time.Millisecond)
fetcher := call.Retry(3, backoff, isTransient, fetchData)
```

```go
// Circuit breaker — trips after 5 consecutive failures, resets after 30s
breaker := call.NewBreaker(call.BreakerConfig{
    ResetTimeout: 30 * time.Second,
    ReadyToTrip:  call.ConsecutiveFailures(5),
})
safeFetch := call.WithBreaker(breaker, fetchFromAPI)
resp, err := safeFetch(ctx, url)  // returns call.ErrOpen when tripped
```

```go
// Bound concurrency — at most 5 in-flight API calls
callAPI := call.Throttle(5, fetchFromAPI)
```

```go
// Bound by total cost — large items consume more budget
fetchData := call.ThrottleWeighted(100, estimateSize, fetchFromAPI)
```

```go
// Cancel remaining work on first error
failFast := call.OnErr(fetchURL, func(_ error) { cancel() })
```

```go
// Transform errors without changing the function signature
annotated := call.MapErr(fetchUser, classifyError)
```

```go
// Debounce rapid calls, execute once after quiet period
d := call.NewDebouncer(500*time.Millisecond, saveConfig)
defer d.Close()
d.Call(cfg)
```

## Operations

**Circuit Breaking**
- `NewBreaker(cfg BreakerConfig) *Breaker` — 3-state: closed → open → half-open → closed
- `WithBreaker[T, R](b *Breaker, fn) fn` — wrap fn with breaker protection
- `ConsecutiveFailures(n int) func(Snapshot) bool` — ReadyToTrip predicate
- `ErrOpen` — sentinel error when breaker rejects

**Retry**
- `Retry[T, R](maxAttempts, backoff, shouldRetry, fn) fn` — retry on error with pluggable backoff
- `ConstantBackoff(delay) Backoff` — fixed delay
- `ExponentialBackoff(initial) Backoff` — full jitter: random in [0, initial * 2^n)

**Concurrency Control**
- `Throttle[T, R](n, fn) fn` — bound by call count
- `ThrottleWeighted[T, R](capacity, cost, fn) fn` — bound by total cost

**Side-Effect Wrappers**
- `OnErr[T, R](fn, onErr) fn` — call handler on error
- `MapErr[T, R](fn, mapper) fn` — transform errors

**Debounce**
- `NewDebouncer[T](wait, fn, opts...) *Debouncer[T]` — trailing-edge coalescer
- `MaxWait(d) DebounceOption` — cap maximum deferral

All context-aware wrappers return `ctx.Err()` on cancellation. `WithBreaker` does not count `context.Canceled` as a failure. All functions panic on nil inputs.

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/call) for complete API documentation and the [orders example](../examples/orders/) for a full integration demo.
