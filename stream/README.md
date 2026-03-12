# stream

Lazy, memoized, persistent sequences. Zero goroutines, zero channels.

Use `stream` when you need lazy evaluation — infinite sequences, early termination, or deferred computation over expensive elements. For finite in-memory collections, use `slice`. For lightweight `iter.Seq` chaining without caching, use `seq`.

```go
// Before: channel-based Fibonacci leaks a goroutine when you stop reading
func fib() <-chan int {
    ch := make(chan int)
    go func() {
        a, b := 0, 1
        for { ch <- a; a, b = b, a+b }
    }()
    return ch
}
// goroutine runs forever after consumer stops

// After: lazy stream — no goroutine, no channel, no leak
type pair struct{ a, b int }
fib := stream.Unfold(pair{0, 1}, func(p pair) (int, pair, bool) {
    return p.a, pair{p.b, p.a + p.b}, true
})
first10 := fib.Take(10).Collect()
```

## What It Looks Like

```go
// Infinite sequence of natural numbers
naturals := stream.Generate(0, func(n int) int { return n + 1 })
```

```go
// First 10 primes — lazy filter over infinite sequence
primes := stream.Generate(2, inc).KeepIf(isPrime).Take(10).Collect()
```

```go
// Cross-type lazy map (standalone — Go methods can't introduce type params)
names := stream.Map(users, User.Name).Collect()
```

```go
// Bridge to Go's range protocol
for v := range stream.Of(1, 2, 3).Seq() {
    fmt.Println(v)
}
```

```go
// Cursor-based pagination — each step returns a page and optional next cursor
pages := stream.Paginate(firstCursor, fetchPage)
allItems := stream.Map(pages, Page.Items).Collect()
```

## Head-Eager, Tail-Lazy

When a stream cell exists, its head value is already computed. Only the tail is deferred. This means:

- `KeepIf` eagerly scans forward until it finds a match (which may not terminate on infinite streams — see Caveats)
- `Take(n)` returns a stream capped at n elements — the current cell is available immediately, the remainder is produced lazily as tails are forced
- `Collect()` forces all remaining thunks and materializes to a slice

The zero value is an empty stream. `First()` returns a not-ok option on empty. `Tail()` returns empty on empty. `Collect()` returns nil on empty.

## Memoization and Persistence

Each tail thunk runs at most once on success. After evaluation, the result is cached — calling `.Collect()` twice on the same stream returns the same elements without re-computing them:

```go
s := stream.Generate(0, expensiveStep).Take(1000)
a := s.Collect()  // computes all 1000 steps
b := s.Collect()  // returns cached results — no recomputation
```

Multiple references to the same stream share the cache. This is what makes streams **persistent** — operations return new streams, but shared prefixes are computed once. Note that downstream operations (predicates, transforms) in derived streams are not shared — only the upstream forcing is deduplicated:

```go
s := stream.Generate(0, expensiveStep).Take(1000)
evens := s.KeepIf(isEven)   // s is the shared source
odds  := s.KeepIf(isOdd)    // same s — steps are computed once
```

**Retention cost:** Holding a reference to an early cell pins all forced suffix cells in memory. Release references when you're done to allow GC.

Thread-safe: concurrent forcing of the same stream is coordinated via state machine transitions — thunks execute outside internal locks.

## Caveats

**Retry-on-panic.** If a tail thunk panics, the cell resets to pending and a later `Tail()` call retries it. Avoid side effects in deferred computations unless retry is acceptable. "At most once" means at most once on success.

**Eager first step.** `Unfold` and `Paginate` compute the first element eagerly at construction time. Panics during that first step are not memoized or retried.

**Non-termination.** Some operations may not terminate on infinite streams:
- `KeepIf`, `Find`, `Any` — if no element matches
- `DropWhile` — if the predicate never becomes false
- `Collect`, `Each`, `Fold` — on any infinite stream

**Backing-array retention.** `From` captures subslice views of the input slice. The backing array may be retained until those stream nodes are forced or become unreachable. For very large slices, this can retain more memory than expected.

**Reentrancy.** Callbacks must not force the same cell being evaluated (directly or indirectly). This is inherent to memoized lazy evaluation.

All callback-taking functions panic on nil inputs.

## When to Use Stream vs Seq vs Slice

| Use stream when...                  | Use seq when...                    | Use slice when...                  |
| ----------------------------------- | ---------------------------------- | ---------------------------------- |
| Sequence is infinite                | You have an `iter.Seq` to chain    | Collection is finite and in memory |
| Multiple consumers share evaluation | Pipeline can re-evaluate each call | All elements will be consumed      |
| Elements are expensive to compute   | Lightweight, no caching needed     | Elements are cheap or pre-computed |
| Memoization and persistence matter  | No retention cost needed           | Eager execution is fine            |

## Operations

**Create**: `From`, `Of`, `Generate`, `Repeat`, `Unfold`, `Paginate`, `Prepend`, `PrependLazy`

**Lazy** (return Stream): `KeepIf`, `Convert`, `Take`, `TakeWhile`, `Drop`, `DropWhile`, `Map` (standalone)

**Terminal** (force evaluation): `Each`, `Collect`, `Find`, `Any`, `Seq`, `Fold` (standalone)

**Access**: `IsEmpty`, `First`, `Tail`

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/stream) for complete API documentation, the [main README](../README.md) for installation, and the [showcase](../docs/showcase.md) for real-world comparisons.
