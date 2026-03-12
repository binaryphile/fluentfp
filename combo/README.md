# combo

Combinatorial constructions: Cartesian products, permutations, combinations, and power sets.

All functions eagerly allocate the full result in memory. `CartesianProduct` returns `pair.Pair` values; use `CartesianProductWith` to produce your own result type directly and avoid intermediate pairs. Bridge with `slice.From()` for fluent chains.

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

## Growth Rates

All results are fully materialized in memory. Compute the result size before calling:

| Function | Results | n=5 | n=10 | n=15 |
|----------|---------|-----|------|------|
| Permutations | n! | 120 | 3.6M | 1.3T |
| PowerSet | 2^n | 32 | 1,024 | 32,768 |
| Combinations(n, k) | C(n, k) | C(5,2)=10 | C(10,5)=252 | C(15,7)=6,435 |
| CartesianProduct | a * b | depends on inputs | | |

Use only for small inputs unless you've computed the result size and memory cost. `Permutations` becomes impractical above ~10-12 elements. `PowerSet` grows quickly. `Combinations` can also be large near midpoint `k`.

## Empty and Invalid Input

| Function | Empty/nil input | Invalid args |
|----------|----------------|--------------|
| `CartesianProduct` | `nil` if either input is empty/nil | — |
| `CartesianProductWith` | `nil` if either input is empty/nil | `fn` must not be nil |
| `Permutations` | `[[]]` (one empty permutation) | — |
| `Combinations` | `[[]]` for `k=0` | `nil` if `k < 0` or `k > len(items)` |
| `PowerSet` | `[[]]` (one empty subset) | — |

## Operations

- `CartesianProduct[A, B any]([]A, []B) []pair.Pair[A, B]` — all pairs
- `CartesianProductWith[A, B, R any]([]A, []B, func(A, B) R) []R` — all pairs, transformed (avoids intermediate `pair.Pair` allocation)
- `Permutations[T any]([]T) [][]T` — all orderings
- `Combinations[T any]([]T, int) [][]T` — k-element subsets, preserving order
- `PowerSet[T any]([]T) [][]T` — all subsets

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/combo) for complete API documentation and the [main README](../README.md) for installation.
