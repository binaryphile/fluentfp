# stream

Lazy, memoized, persistent sequences. Zero goroutines, zero channels.

Use `stream` when you need lazy evaluation — infinite sequences, early termination, or deferred computation over expensive elements. For finite in-memory collections, use `slice` instead.

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
// Lazy from slice — useful when only a subset is needed
first := stream.From(largeSlice).KeepIf(isValid).First()
```

## Head-Eager, Tail-Lazy

When a stream cell exists, its head value is already computed. Only the tail is deferred. This means:

- `KeepIf` eagerly scans to the first match — you get immediate feedback
- `Take(n)` builds a chain of n cells, each with a lazy tail
- `Collect()` forces all remaining thunks and materializes to a slice

Cells are memoized via state machine transitions — once a tail thunk evaluates, the result is cached for all future accesses. Thread-safe without holding locks during thunk execution.

## When to Use Stream vs Slice

| Use stream when... | Use slice when... |
|-------------------|------------------|
| Sequence is infinite | Collection is finite and in memory |
| Only first N elements needed | All elements will be consumed |
| Elements are expensive to compute | Elements are cheap or pre-computed |
| Lazy composition matters | Eager execution is fine |

## Operations

**Create**: `From`, `Of`, `Generate`, `Repeat`, `Unfold`

**Lazy** (return Stream): `KeepIf`, `Convert`, `Take`, `TakeWhile`, `Drop`, `DropWhile`, `Map` (standalone)

**Terminal** (force evaluation): `Each`, `Collect`, `Find`, `Any`, `Seq`, `Fold` (standalone)

**Access**: `IsEmpty`, `First`, `Tail`

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/stream) for complete API documentation, the [main README](../README.md) for installation, and the [showcase](../docs/showcase.md) for real-world comparisons.
