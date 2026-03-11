# combo

Combinatorial constructions: Cartesian products, permutations, combinations, and power sets.

Standalone package returning plain slices. Bridge with `slice.From()` for fluent chains.

```go
// Before: nested loop to generate all size/color combinations
var pairs []Option
for _, size := range sizes {
    for _, color := range colors {
        pairs = append(pairs, Option{size, color})
    }
}

// After: one call
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
// Chain with slice.From for filtering/sorting
valid := slice.From(combo.CartesianProduct(keys, ids)).KeepIf(isCompatible)
```

## Growth Rates

Results grow fast. Know the size before calling:

| Function | Results | n=5 | n=10 | n=15 |
|----------|---------|-----|------|------|
| Permutations | n! | 120 | 3.6M | 1.3T |
| PowerSet | 2^n | 32 | 1,024 | 32,768 |
| Combinations(n, k) | C(n, k) | C(5,2)=10 | C(10,5)=252 | C(15,7)=6,435 |
| CartesianProduct | a * b | depends on inputs | | |

Permutations above ~12 elements will exhaust memory. PowerSet and Combinations are practical for larger inputs.

## Operations

- `CartesianProduct[A, B]([]A, []B) []pair.Pair[A, B]` — all pairs
- `CartesianProductWith[A, B, R]([]A, []B, func(A, B) R) []R` — all pairs, transformed
- `Permutations[T]([]T) [][]T` — all orderings
- `Combinations[T]([]T, int) [][]T` — k-element subsets
- `PowerSet[T]([]T) [][]T` — all subsets

Nil or empty inputs return nil for Cartesian products and `[[]]` for Permutations, Combinations(k=0), and PowerSet (the empty set is the one result).

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/combo) for complete API documentation and the [main README](../README.md) for installation.
