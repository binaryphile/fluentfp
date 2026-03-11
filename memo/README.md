# memo

Memoize function results — cache once, return forever.

Use `memo` when you have expensive or repeated computations that should only run once per input. For lazy sequences with shared evaluation, use `stream`. For eager collection transforms, use `slice`.

```go
// Before: 14 lines of lock choreography
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

// After: one line
lookup := memo.Fn(expensiveLookup)
```

## What It Looks Like

```go
// Lazy initialization — computed once, cached forever
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

```go
// Cached transform in a fluent chain
cachedLookup := memo.Fn(lookupName)
names := slice.From(ids).Convert(cachedLookup)
```

## Retry on Panic, Retry on Error

**`Of` retries panicked functions.** Unlike `sync.Once`, which poisons permanently on panic, `Of` resets to pending — subsequent calls retry the function. If initialization panics due to a transient condition (network, file not ready), the next call gets another chance.

**`FnErr` retries errors.** Only successful results are cached. If the underlying function returns an error, subsequent calls retry rather than returning the cached error. This matches the assumption that errors are transient — if you need to cache errors, use `Fn` with `result.Result[V]` as the value type.

**Successes are permanent.** Once a function returns successfully, the result is cached and the original function is released for GC.

## Operations

**Memoize**
- `Of[T](fn func() T) func() T` — zero-arg; retry-on-panic
- `Fn[K, V](fn func(K) V) func(K) V` — keyed; unbounded cache
- `FnErr[K, V](fn func(K) (V, error)) func(K) (V, error)` — keyed; caches successes only
- `FnWith[K, V](fn func(K) V, cache Cache[K, V]) func(K) V` — keyed; custom cache
- `FnErrWith[K, V](fn func(K) (V, error), cache Cache[K, V]) func(K) (V, error)` — keyed; custom cache, caches successes only

**Caches**
- `Cache[K, V]` — interface: `Load(K) (V, bool)`, `Store(K, V)`
- `NewMap[K, V]() Cache[K, V]` — unbounded, concurrent-safe (sync.RWMutex)
- `NewLRU[K, V](capacity int) Cache[K, V]` — LRU eviction, concurrent-safe (sync.Mutex)

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/memo) for complete API documentation and the [main README](../README.md) for installation.
