# seq

Fluent chains on Go's `iter.Seq` — lazy, re-evaluating, stdlib-compatible.

Use `seq` when you have an `iter.Seq` and want to filter, transform, or collect it without a manual loop. For finite in-memory slices, use `slice` instead. For memoized lazy sequences, use `stream`.

```go
// Before: manual loop to filter an iterator
var active []string
for k := range maps.Keys(configs) {
    if isActive(k) {
        active = append(active, k)
    }
}

// After: fluent chain on the iterator
active := seq.FromIter(maps.Keys(configs)).KeepIf(isActive).Collect()
```

## What It Looks Like

```go
// Lazy filter + limit — stops after finding 5 matches (doesn't filter the whole slice)
top5 := seq.From(items).KeepIf(Item.IsActive).Take(5).Collect()
```

```go
// Infinite sequence
naturals := seq.Generate(0, func(n int) int { return n + 1 })
first10 := naturals.Take(10).Collect()
```

```go
// Range works directly — no .Iter() needed
for v := range seq.From(data).KeepIf(isValid) {
    fmt.Println(v)
}
```

```go
// .Iter() when passing to functions expecting iter.Seq
slices.Collect(seq.From(data).KeepIf(isValid).Iter())
```

```go
// Cross-type map (standalone — Go methods can't introduce type params)
names := seq.Map(users, User.Name).Collect()
```

## Re-Evaluation

Seq pipelines re-evaluate on every terminal call. There is no caching:

```go
evens := seq.From(numbers).KeepIf(isEven)
a := evens.Collect()  // runs the filter
b := evens.Collect()  // runs the filter again
```

This means seq pipelines are lightweight descriptions — no hidden state, no memoization overhead. But if the source is expensive or has side effects, each terminal call pays the full cost.

For cached evaluation, use `stream` instead.

## When to Use Seq vs Stream vs Slice

| | seq | stream | slice |
|---|---|---|---|
| **Evaluation** | Lazy, re-evaluates each terminal call | Lazy, memoized (cached) | Eager (immediate) |
| **Persistence** | Reusable — re-evaluates each time | Persistent — forced cells are shared | Persistent — slices are values |
| **Memory** | No intermediate allocations | Retains forced cells | Full slice per step |
| **Best for** | Stdlib interop, simple lazy chains | Infinite sequences, shared evaluation | Finite in-memory collections |
| **Interop** | `iter.Seq[T]` (Go stdlib) | `.Seq()` bridge to iter.Seq | `[]T` (Go native) |

**Rule of thumb:** Use `slice` for finite in-memory data (most cases). Use `seq` when you have an `iter.Seq` and want fluent chaining. Use `stream` when you need memoization or persistence.

## Operations

**Create**: `From`, `FromIter`, `Of`, `Generate`, `Repeat`

**Lazy** (return Seq): `KeepIf`, `RemoveIf`, `Convert`, `Take`, `Drop`, `TakeWhile`, `DropWhile`, `Map` (standalone)

**Terminal** (force evaluation): `Collect`, `Find`, `Any`, `Every`, `None`, `Each`, `Fold` (standalone)

**Unwrap**: `Iter` — return to `iter.Seq[T]` for stdlib interop

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/seq) for complete API documentation, the [main README](../README.md) for installation, and [stream](../stream/) for memoized lazy sequences.
