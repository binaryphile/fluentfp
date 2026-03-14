package seq

import "github.com/binaryphile/fluentfp/tuple/pair"

// Enumerate pairs each element with its zero-based index, lazily.
// The index resets on each iteration. Safe for repeated use.
func Enumerate[T any](s Seq[T]) Seq[pair.Pair[int, T]] {
	return Seq[pair.Pair[int, T]](func(yield func(pair.Pair[int, T]) bool) {
		i := 0

		for v := range s {
			if !yield(pair.Of(i, v)) {
				return
			}

			i++
		}
	})
}
