package seq

import (
	"iter"

	"github.com/binaryphile/fluentfp/tuple/pair"
)

// Map applies fn to each element, returning a Seq of a different type.
// Standalone because Go methods cannot introduce additional type parameters.
// Panics if fn is nil.
func Map[T, R any](s Seq[T], fn func(T) R) Seq[R] {
	if fn == nil {
		panic("seq.Map: fn must not be nil")
	}

	if s == nil {
		return Empty[R]()
	}

	return Seq[R](func(yield func(R) bool) {
		for v := range s {
			if !yield(fn(v)) {
				return
			}
		}
	})
}

// FlatMap applies fn to each element of s and concatenates the resulting Seqs.
// Standalone because Go methods cannot introduce additional type parameters.
// Panics if fn is nil.
func FlatMap[T, R any](s Seq[T], fn func(T) Seq[R]) Seq[R] {
	if fn == nil {
		panic("seq.FlatMap: fn must not be nil")
	}

	if s == nil {
		return Empty[R]()
	}

	return Seq[R](func(yield func(R) bool) {
		for v := range s {
			inner := fn(v)
			if inner == nil {
				continue
			}

			for r := range inner {
				if !yield(r) {
					return
				}
			}
		}
	})
}

// Concat returns a Seq that yields all elements of a followed by all elements of b.
func Concat[T any](a, b Seq[T]) Seq[T] {
	if a == nil && b == nil {
		return Empty[T]()
	}

	return Seq[T](func(yield func(T) bool) {
		if a != nil {
			for v := range a {
				if !yield(v) {
					return
				}
			}
		}

		if b != nil {
			for v := range b {
				if !yield(v) {
					return
				}
			}
		}
	})
}

// Zip returns a Seq of pairs from corresponding elements of a and b.
// Truncates to the shorter sequence. Note: a is the driving side —
// if b is shorter, one extra element of a is consumed before truncation
// is detected. For side-effectful or single-use sources, be aware of
// this left-consumption bias.
func Zip[A, B any](a Seq[A], b Seq[B]) Seq[pair.Pair[A, B]] {
	if a == nil || b == nil {
		return Empty[pair.Pair[A, B]]()
	}

	return Seq[pair.Pair[A, B]](func(yield func(pair.Pair[A, B]) bool) {
		next, stop := iter.Pull(iter.Seq[B](b))
		defer stop()

		for va := range a {
			vb, ok := next()
			if !ok {
				return
			}

			if !yield(pair.Of(va, vb)) {
				return
			}
		}
	})
}

// Scan reduces a Seq like Fold, but yields all intermediate accumulator values.
// It includes the initial value as the first element (scanl semantics),
// so the result has len(s)+1 elements for a finite Seq.
// Standalone because Go methods cannot introduce additional type parameters.
// Panics if fn is nil.
func Scan[T, R any](s Seq[T], initial R, fn func(R, T) R) Seq[R] {
	if fn == nil {
		panic("seq.Scan: fn must not be nil")
	}

	return Seq[R](func(yield func(R) bool) {
		acc := initial
		if !yield(acc) {
			return
		}

		if s == nil {
			return
		}

		for v := range s {
			acc = fn(acc, v)
			if !yield(acc) {
				return
			}
		}
	})
}

// FilterMap applies fn to each element and keeps only the results where fn returns true.
// Combines filtering and type-changing transformation in a single lazy pass.
// Fully streaming with O(1) state. Works with infinite sequences.
// Panics if fn is nil.
func FilterMap[T, R any](s Seq[T], fn func(T) (R, bool)) Seq[R] {
	if fn == nil {
		panic("seq.FilterMap: fn must not be nil")
	}

	if s == nil {
		return Empty[R]()
	}

	return Seq[R](func(yield func(R) bool) {
		for v := range s {
			if r, ok := fn(v); ok {
				if !yield(r) {
					return
				}
			}
		}
	})
}

// Contains returns true if target is in the sequence. Short-circuits on first match.
// On infinite sequences, terminates only if a match is found.
// Note: for float types, NaN != NaN, so Contains(s, NaN) is always false.
// Standalone because the comparable constraint cannot be expressed on the Seq[T any] receiver.
func Contains[T comparable](s Seq[T], target T) bool {
	if s == nil {
		return false
	}

	for v := range s {
		if v == target {
			return true
		}
	}

	return false
}

// Chunk groups elements into slices of at most size elements.
// The last chunk may have fewer than size elements.
// Each emitted slice is a stable snapshot with independent backing storage.
// Buffers up to size elements. Works with infinite sequences.
// Panics if size <= 0.
func Chunk[T any](s Seq[T], size int) Seq[[]T] {
	if size <= 0 {
		panic("seq.Chunk: size must be > 0")
	}

	if s == nil {
		return Empty[[]T]()
	}

	return Seq[[]T](func(yield func([]T) bool) {
		buf := make([]T, 0, size)

		for v := range s {
			buf = append(buf, v)

			if len(buf) == size {
				if !yield(buf) {
					return
				}

				buf = make([]T, 0, size)
			}
		}

		if len(buf) > 0 {
			if !yield(buf) {
				return
			}
		}
	})
}

// Fold reduces a Seq to a single value by applying fn to an accumulator
// and each element. Requires a finite sequence.
// Standalone because Go methods cannot introduce additional type parameters.
// Panics if fn is nil.
func Fold[T, R any](s Seq[T], initial R, fn func(R, T) R) R {
	if fn == nil {
		panic("seq.Fold: fn must not be nil")
	}

	acc := initial

	if s == nil {
		return acc
	}

	for v := range s {
		acc = fn(acc, v)
	}

	return acc
}
