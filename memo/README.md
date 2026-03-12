# memo

Memoize function results with concurrent-safe caches.

Use `memo` when you want to cache successful results by input. For lazy sequences with shared evaluation, use `stream`. For eager collection transforms, use `slice`.

```go
// Before: manual cache with lock choreography
var (
    mu    sync.RWMutex
    cache = make(map[string]Result)
)
func lookup(key string) Result {
    mu.RLock()
    if v, ok := cache[key]; ok {
        mu.RUnlock()
        return v
    }
    mu.RUnlock()
    v := expensiveLookup(key)
    mu.Lock()
    cache[key] = v
    mu.Unlock()
    return v
}

// After: one line — concurrent-safe, but see Concurrency note below
lookup := memo.Fn(expensiveLookup)
```

## What It Looks Like

```go
// Lazy initialization — computed once, cached permanently
getConfig := memo.Of(loadConfig)
cfg := getConfig()  // computes
cfg = getConfig()   // returns cached result
```

```go
// Fallible function — errors retry, successes cached
fetch := memo.FnErr(callAPI)
v, err := fetch("key")  // calls API
v, err = fetch("key")   // cached (if first call succeeded)
```

```go
// Bounded cache — LRU eviction at capacity
lookup := memo.FnWith(expensiveLookup, memo.NewLRU[string, Result](1000))
```

## Concurrency Semantics

**`Of` guarantees single evaluation.** Concurrent callers wait for the in-flight evaluation — only one goroutine computes the result. This is true single-flight behavior.

**Keyed wrappers (`Fn`, `FnWith`, `FnErr`, `FnErrWith`) do not coalesce concurrent misses.** The cache is concurrent-safe, but multiple goroutines requesting the same uncached key may all compute it simultaneously. If deduplication matters (expensive API calls, stampede-prone workloads), use `golang.org/x/sync/singleflight` or add external coordination.

## Retry on Panic, Retry on Error

**`Of` retries panicked functions.** Unlike `sync.Once`, which poisons permanently on panic, `Of` resets to pending — subsequent calls retry the function. If initialization panics due to a transient condition (network, file not ready), the next call gets another chance.

**`FnErr` retries errors.** Only successful results are cached. If the underlying function returns an error, subsequent calls retry rather than returning the cached error. This matches the assumption that errors are transient — if you need to cache errors, use `Fn` with `result.Result[V]` as the value type.

**`Of`: successes are permanent.** Once `Of`'s function returns successfully, the result is cached and the original function is released for GC. Keyed wrappers retain `fn` for future uncached keys.

## Operations

**Memoize**
- `Of[T any](fn func() T) func() T` — zero-arg; single-flight; retry-on-panic
- `Fn[K comparable, V any](fn func(K) V) func(K) V` — keyed; unbounded cache
- `FnErr[K comparable, V any](fn func(K) (V, error)) func(K) (V, error)` — keyed; caches successes only
- `FnWith[K comparable, V any](fn func(K) V, cache Cache[K, V]) func(K) V` — keyed; custom cache
- `FnErrWith[K comparable, V any](fn func(K) (V, error), cache Cache[K, V]) func(K) (V, error)` — keyed; custom cache, caches successes only

**Caches**
- `Cache[K comparable, V any]` — interface: `Load(K) (V, bool)`, `Store(K, V)`
- `NewMap[K comparable, V any]() Cache[K, V]` — unbounded, concurrent-safe (sync.RWMutex)
- `NewLRU[K comparable, V any](capacity int) Cache[K, V]` — LRU eviction, concurrent-safe (sync.Mutex)

All functions panic on nil inputs (`fn`, `cache`). `NewLRU` panics on non-positive capacity. `K` must be `comparable` (map key constraint).

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/memo) for complete API documentation and the [main README](../README.md) for installation.
