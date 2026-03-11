# heap

Persistent pairing heap (priority queue) parameterized by a comparator. Based on Stone's *Algorithms for Functional Programming* (Ch 4).

All operations return new heaps — the original is never modified. Use `slice.Asc` and `slice.Desc` to build comparators.

## What It Looks Like

```go
// Min-heap by priority
h := heap.New(slice.Asc(Task.Priority))
h = h.Insert(task1).Insert(task2).Insert(task3)

// Peek at minimum
next := h.Min()  // option.Option[Task]

// Pop minimum and continue
task, rest, ok := h.Pop()

// Drain in sorted order
sorted := h.Collect()

// Merge two heaps in O(1)
combined := h1.Merge(h2)
```

## Persistence

Every operation returns a new heap. The old heap remains valid and unchanged:

```go
h1 := heap.From(items, cmp)
h2 := h1.Insert(x)      // h1 still has original elements
h3 := h1.DeleteMin()     // h1 still has original elements
```

## Operations

**Create**
- `New(cmp)` — empty heap with comparator
- `From(items, cmp)` — heap from slice

**Core** (return new heap)
- `.Insert(t)` — add element, O(1)
- `.DeleteMin()` — remove minimum, O(log n) amortized
- `.Merge(other)` — combine two heaps, O(1)

**Query**
- `.Min()` — peek at minimum (`option.Option[T]`), O(1)
- `.Pop()` — minimum + remaining heap + ok, O(log n) amortized
- `.IsEmpty()` — true if no elements
- `.Len()` — element count

**Consume**
- `.Collect()` — all elements in sorted order, O(n log n)

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/heap) for complete API documentation and the [main README](../README.md) for installation.
