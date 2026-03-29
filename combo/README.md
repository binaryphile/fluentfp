# combo

Combinatorial constructions: Cartesian products, permutations, combinations, and power sets.

Eager functions allocate the full result in memory as `slice.Mapper` — chain directly with `.KeepIf`, `.Convert`, etc. Seq variants (`SeqPermutations`, `SeqPowerSet`, etc.) return `seq.Seq` for lazy evaluation with early termination — use these when the full result is too large to materialize.

```go
// Before: nested loop to generate all size/color combinations
var pairs []Option
for _, size := range sizes {
    for _, color := range colors {
        pairs = append(pairs, Option{size, color})
    }
}

// After: one call — produces domain objects directly
pairs := combo.CartesianProductWith(sizes, colors, NewOption)
```

## What It Looks Like

```go
// All pairs from two slices
pairs := combo.CartesianProduct([]int{1, 2}, []string{"a", "b"})
// [{1 a} {1 b} {2 a} {2 b}]
```

```go
// All orderings (n! results)
combo.Permutations([]int{1, 2, 3})
// [[1 2 3] [1 3 2] [2 1 3] [2 3 1] [3 1 2] [3 2 1]]
```

```go
// k-element subsets — C(n,k) results
combo.Combinations([]string{"a", "b", "c", "d"}, 2)
// [[a b] [a c] [a d] [b c] [b d] [c d]]
```

```go
// All subsets (2^n results)
combo.PowerSet([]int{1, 2})
// [[] [2] [1] [1 2]]
```

```go
// Fluent chain — results are slice.Mapper, so chain directly
// Keep only 2-element combinations that sum to at least 5
combo.Combinations([]int{1, 2, 3, 4}, 2).KeepIf(sumsToAtLeast5)
// [[1 4] [2 3] [2 4] [3 4]]
```

## Growth Rates

All results are fully materialized in memory. Compute the result size before calling:

| Function | Results | n=5 | n=10 | n=15 |
|----------|---------|-----|------|------|
| Permutations | n! | 120 | 3.6M | 1.3T |
| PowerSet | 2^n | 32 | 1,024 | 32,768 |
| Combinations(n, k) | C(n, k) | C(5,2)=10 | C(10,5)=252 | C(15,7)=6,435 |
| CartesianProduct | a * b | depends on inputs | | |

Eager functions are impractical for large inputs — `Permutations` above ~10-12 elements, `PowerSet` above ~20. Use the `Seq` variants for lazy evaluation with early termination:

```go
// Lazy: only generates elements as needed
for perm := range combo.SeqPermutations(largeSlice) {
    if isInteresting(perm) {
        break // stops generation
    }
}
```

## Empty and Invalid Input

| Function | Empty/nil input | Invalid args |
|----------|----------------|--------------|
| `CartesianProduct` | `nil` if either input is empty/nil | — |
| `CartesianProductWith` | `nil` if either input is empty/nil | `fn` must not be nil |
| `Permutations` | `[[]]` (one empty permutation) | — |
| `Combinations` | `[[]]` for `k=0` | `nil` if `k < 0` or `k > len(items)` |
| `PowerSet` | `[[]]` (one empty subset) | — |

## Operations

### Eager (return `slice.Mapper`)

- `CartesianProduct[A, B any]([]A, []B) slice.Mapper[pair.Pair[A, B]]` — all pairs
- `CartesianProductWith[A, B, R any]([]A, []B, func(A, B) R) slice.Mapper[R]` — all pairs, transformed (avoids intermediate `pair.Pair` allocation)
- `Permutations[T any]([]T) slice.Mapper[[]T]` — all orderings
- `Combinations[T any]([]T, int) slice.Mapper[[]T]` — k-element subsets, preserving order
- `PowerSet[T any]([]T) slice.Mapper[[]T]` — all subsets

### Lazy (return `seq.Seq`)

- `SeqCartesianProduct[A, B any]([]A, []B) seq.Seq[pair.Pair[A, B]]` — lazy pairs
- `SeqCartesianProductWith[A, B, R any]([]A, []B, func(A, B) R) seq.Seq[R]` — lazy pairs, transformed
- `SeqPermutations[T any]([]T) seq.Seq[[]T]` — lazy orderings
- `SeqCombinations[T any]([]T, int) seq.Seq[[]T]` — lazy k-element subsets
- `SeqPowerSet[T any]([]T) seq.Seq[[]T]` — lazy subsets

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/combo) for complete API documentation and the [main README](../README.md) for installation.
