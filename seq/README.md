# seq

Fluent chains on Go's `iter.Seq` — lazy, re-evaluating, stdlib-compatible.

Use `seq` when you want laziness, early termination, or `iter.Seq` interop. Prefer `slice` for eager in-memory collection work. For memoized lazy sequences, use `stream`.

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
// Lazy filter + limit — stops after finding 5 matches
top5 := seq.From(items).KeepIf(Item.IsActive).Take(5).Collect()
```

```go
// Infinite sequence — always use Take or TakeWhile before a terminal op
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

```go
// FlatMap — expand each item into a sub-sequence and flatten
allTags := seq.FlatMap(items, Item.Tags).Collect()
```

```go
// Zip — pair corresponding elements from two sequences
ranked := seq.Zip(seq.From(names), seq.From(scores)).Collect()
```

```go
// Scan — running totals as a lazy sequence
totals := seq.Scan(seq.From(amounts), 0.0, sumFloat64).Collect()
```

## Re-Evaluation

Seq pipelines re-evaluate on every terminal call. There is no caching:

```go
evens := seq.From(numbers).KeepIf(isEven)
a := evens.Collect()  // runs the filter
b := evens.Collect()  // runs the filter again
```

This means seq pipelines are lightweight descriptions — no hidden state, no memoization overhead. But if the source is expensive or has side effects, each terminal call pays the full cost. Replayability depends on the source — `FromIter` wraps the given `iter.Seq` as-is, and stateful or single-use sources may not produce the same results on re-invocation.

For cached evaluation, use `stream` instead.

## Behavior Notes

The zero value of `Seq[T]` is nil. It is **not safe for direct range** — use `Empty`, `From`, or other constructors. All constructors and Seq-returning operations return non-nil Seqs safe for range. Lazy operations are nil-safe on the receiver and return empty (non-nil) Seqs, enabling safe chaining. `From(nil)` and `From([]T{})` both return empty Seqs. `Collect()` on a nil Seq returns nil.

`Every` and `None` return true on empty or nil input (vacuous truth). `Find` returns `option.Option[T]` — not-ok if no match is found.

`Convert` is a same-type transform (method). `Map` is a cross-type transform (standalone, because Go methods can't introduce additional type parameters).

**Non-termination:** `Collect`, `Each`, and `Fold` on infinite sequences will not terminate. Always use `Take` or `TakeWhile` to bound infinite sequences before calling a terminal operation.

**Zip left-consumption bias:** `Zip` drives iteration from the first sequence. If the second sequence is shorter, one extra element from the first is consumed before exhaustion is detected. For side-effectful or single-use sources, be aware of this asymmetry.

All callback-taking functions panic on nil callbacks. `FlatMap` treats nil inner Seqs as empty.

## When to Use Seq vs Stream vs Slice

| | seq | stream | slice |
|---|---|---|---|
| **Evaluation** | Lazy, re-evaluates each terminal call | Lazy, memoized (cached) | Eager (immediate) |
| **Persistence** | Re-invokes source on each terminal call | Persistent — forced cells are shared | Persistent — slices are values |
| **Memory** | No intermediate collections during lazy chaining | Retains forced cells | Full slice per step |
| **Best for** | Stdlib interop, laziness, early termination | Shared evaluation, infinite sequences | Finite in-memory collections |
| **Interop** | `iter.Seq[T]` (Go stdlib) | `.Seq()` bridge to iter.Seq | `[]T` (Go native) |

## Operations

**Create**: `From`, `FromIter`, `Of`, `Generate`, `Repeat`

**Lazy** (return Seq): `KeepIf`, `RemoveIf`, `Convert` (same-type), `Take`, `Drop`, `TakeWhile`, `DropWhile`, `Map` (cross-type, standalone), `FlatMap` (standalone), `Concat` (standalone), `Zip` (standalone), `Scan` (standalone)

**Terminal** (force evaluation): `Collect`, `Find` (returns `option.Option[T]`), `Any`, `Every`, `None`, `Each`, `Fold` (standalone)

**Unwrap**: `Iter` — return to `iter.Seq[T]` for stdlib interop

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/seq) for complete API documentation, the [main README](../README.md) for installation, and [stream](../stream/) for memoized lazy sequences.
