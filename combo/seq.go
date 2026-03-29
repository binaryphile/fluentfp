package combo

import (
	"github.com/binaryphile/fluentfp/seq"
	"github.com/binaryphile/fluentfp/tuple/pair"
)

// SeqCartesianProduct returns a lazy sequence of all pairs from a and b.
func SeqCartesianProduct[A, B any](a []A, b []B) seq.Seq[pair.Pair[A, B]] {
	return seq.Seq[pair.Pair[A, B]](func(yield func(pair.Pair[A, B]) bool) {
		for _, x := range a {
			for _, y := range b {
				if !yield(pair.Of(x, y)) {
					return
				}
			}
		}
	})
}

// SeqCartesianProductWith returns a lazy sequence applying fn to every (a, b) pair.
// Avoids intermediate Pair allocation when the caller transforms immediately.
// Panics if fn is nil.
func SeqCartesianProductWith[A, B, R any](a []A, b []B, fn func(A, B) R) seq.Seq[R] {
	if fn == nil {
		panic("combo.SeqCartesianProductWith: fn must not be nil")
	}

	return seq.Seq[R](func(yield func(R) bool) {
		for _, x := range a {
			for _, y := range b {
				if !yield(fn(x, y)) {
					return
				}
			}
		}
	})
}

// SeqPermutations returns a lazy sequence of all orderings of items.
// Returns n! results — use [seq.Seq.Take] for large inputs.
func SeqPermutations[T any](items []T) seq.Seq[[]T] {
	return seq.Seq[[]T](func(yield func([]T) bool) {
		if len(items) == 0 {
			yield([]T{})
			return
		}

		for i := range items {
			rest := make([]T, 0, len(items)-1)
			rest = append(rest, items[:i]...)
			rest = append(rest, items[i+1:]...)

			for perm := range SeqPermutations(rest) {
				if !yield(append([]T{items[i]}, perm...)) {
					return
				}
			}
		}
	})
}

// SeqCombinations returns a lazy sequence of all k-element subsets of items, preserving order.
// Returns C(n,k) results.
func SeqCombinations[T any](items []T, k int) seq.Seq[[]T] {
	return seq.Seq[[]T](func(yield func([]T) bool) {
		if k < 0 || k > len(items) {
			return
		}

		if k == 0 {
			yield([]T{})
			return
		}

		// Include items[0] in the subset
		for c := range SeqCombinations(items[1:], k-1) {
			if !yield(append([]T{items[0]}, c...)) {
				return
			}
		}

		// Exclude items[0] from the subset
		for c := range SeqCombinations(items[1:], k) {
			if !yield(c) {
				return
			}
		}
	})
}

// SeqPowerSet returns a lazy sequence of all subsets of items.
// Returns 2^n results — use [seq.Seq.Take] for large inputs.
func SeqPowerSet[T any](items []T) seq.Seq[[]T] {
	return seq.Seq[[]T](func(yield func([]T) bool) {
		if len(items) == 0 {
			yield([]T{})
			return
		}

		// Subsets without items[0]
		for subset := range SeqPowerSet(items[1:]) {
			if !yield(subset) {
				return
			}
		}

		// Subsets with items[0]
		for subset := range SeqPowerSet(items[1:]) {
			if !yield(append([]T{items[0]}, subset...)) {
				return
			}
		}
	})
}
