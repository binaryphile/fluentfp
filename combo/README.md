# combo

Combinatorial constructions: Cartesian products, permutations, combinations, and power sets.

Standalone package returning plain slices. Bridge with `slice.From()` for fluent chains.

```go
pairs := combo.CartesianProduct(colors, sizes)  // []pair.Pair[string, string]
```

## What It Looks Like

```go
// All pairs from two slices
pairs := combo.CartesianProduct([]int{1, 2}, []string{"a", "b"})
// [{1 a} {1 b} {2 a} {2 b}]
```

```go
// Transform pairs without intermediate allocation
labels := combo.CartesianProductWith(sizes, colors, makeLabel)
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

## Operations

- `CartesianProduct[A, B]([]A, []B) []pair.Pair[A, B]` — all pairs
- `CartesianProductWith[A, B, R]([]A, []B, func(A, B) R) []R` — all pairs, transformed
- `Permutations[T]([]T) [][]T` — all orderings
- `Combinations[T]([]T, int) [][]T` — k-element subsets
- `PowerSet[T]([]T) [][]T` — all subsets

Nil or empty inputs return nil for Cartesian products and `[[]]` for Permutations, Combinations(k=0), and PowerSet (the empty set is the one result).

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/combo) for complete API documentation and the [main README](../README.md) for installation.
