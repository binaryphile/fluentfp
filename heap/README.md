# heap

Persistent pairing heap (priority queue) parameterized by a comparator. Based on Stone's *Algorithms for Functional Programming* (Ch 4).

All mutating operations return new heaps — the original is never modified. Heaps share structure internally, so copies are cheap.

## What It Looks Like

```go
// Min-heap of ints
h := heap.New(lof.IntAsc)
h = h.Insert(3).Insert(1).Insert(2)

// Peek at minimum
if v, ok := h.Min().Get(); ok {
    fmt.Println(v)  // 1
}

// Pop minimum and continue
v, rest, ok := h.Pop()  // v=1, rest has {2, 3}

// Drain in sorted order
sorted := h.Collect()  // [1 2 3]

// Merge two heaps
combined := h.Merge(h2)
```

## When to Use heap

```go
// TopK returns the k largest items without sorting the full slice.
// Uses a min-heap of size k: insert each item, evict the smallest
// when the heap exceeds k. O(n log k) vs O(n log n) for full sort.
func TopK(items []int, k int) []int {
    h := heap.New(lof.IntAsc)
    for _, v := range items {
        h = h.Insert(v)
        if h.Len() > k {
            h = h.DeleteMin()
        }
    }
    return h.Collect()
}
```

Other common uses: merging K sorted sources (log files, database shards) and priority scheduling. For ephemeral priority queues where mutation is fine, use `container/heap`. This package is for when you need persistence — undo, branching, or concurrent readers with no locking.

## Comparator Contract

The comparator determines priority: `cmp(a, b) < 0` means `a` has higher priority than `b`. Use `slice.Asc` / `slice.Desc` or `lof.IntAsc` / `lof.IntDesc` to build comparators, or provide a custom `func(T, T) int`.

**Merge requires the same comparator.** `Merge` uses the receiver's comparator. If `other` was built with a different ordering, the merged heap's ordering is silently incorrect — no error or panic occurs.

## Persistence

Every mutating operation returns a new heap. The old heap remains valid and unchanged:

```go
h1 := heap.From(items, lof.IntAsc)
h2 := h1.Insert(x)      // h1 still has original elements
h3 := h1.DeleteMin()     // h1 still has original elements
```

## Zero Value and Panics

The zero value of `Heap[T]` is an empty heap without a comparator. Queries are safe: `IsEmpty`, `Len`, `Min`, and `Pop` return sensible defaults. Mutating operations (`Insert`, `DeleteMin`, `Merge`) panic because no comparator is available. Always construct with `New` or `From`.

`New` and `From` panic if `cmp` is nil.

## Operations

**Create**
- `New[T any](cmp func(T, T) int) Heap[T]` — empty heap with comparator
- `From[T any](items []T, cmp func(T, T) int) Heap[T]` — heap from slice (repeated Insert)

**Core** (return new heap)
- `.Insert(t)` — add element
- `.DeleteMin()` — remove minimum (returns empty heap with comparator if already empty)
- `.Merge(other)` — combine two heaps (both must use the same ordering)

**Query**
- `.Min()` — peek at minimum (`option.Option[T]`), not-ok if empty
- `.Pop()` — `(elem, rest, true)` or `(zero, empty, false)` if empty
- `.IsEmpty()` — true if no elements
- `.Len()` — element count

**Consume**
- `.Collect()` — all elements in sorted order (returns nil for empty heap)

Textbook pairing heap complexities: Insert and Merge are O(1), DeleteMin is O(log n) amortized. The persistent representation copies child-pointer slices during merge to preserve immutability, so constant factors are higher than an ephemeral implementation.

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/heap) for complete API documentation and the [main README](../README.md) for installation.
