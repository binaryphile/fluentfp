package stream

import (
	"iter"

	"github.com/binaryphile/fluentfp/option"
)

// Each calls fn for every element. Requires a finite stream.
// Panics if fn is nil.
func (s Stream[T]) Each(fn func(T)) {
	if fn == nil {
		panic("stream.Each: fn must not be nil")
	}

	cur := s
	for !cur.IsEmpty() {
		fn(cur.cell.head)
		cur = cur.Tail()
	}
}

// Collect materializes the stream into a slice. Requires a finite stream.
// Returns nil for an empty stream.
func (s Stream[T]) Collect() []T {
	if s.cell == nil {
		return nil
	}

	var result []T

	cur := s
	for !cur.IsEmpty() {
		result = append(result, cur.cell.head)
		cur = cur.Tail()
	}

	return result
}

// Find returns the first element where fn returns true.
// Short-circuits on match. On an infinite stream with no matching element,
// this will not terminate. Panics if fn is nil.
func (s Stream[T]) Find(fn func(T) bool) option.Option[T] {
	if fn == nil {
		panic("stream.Find: fn must not be nil")
	}

	cur := s
	for !cur.IsEmpty() {
		if fn(cur.cell.head) {
			return option.Of(cur.cell.head)
		}

		cur = cur.Tail()
	}

	return option.Option[T]{}
}

// Any returns true if fn returns true for any element.
// Short-circuits on first match. On an infinite stream with no matching element,
// this will not terminate. Panics if fn is nil.
func (s Stream[T]) Any(fn func(T) bool) bool {
	if fn == nil {
		panic("stream.Any: fn must not be nil")
	}

	cur := s
	for !cur.IsEmpty() {
		if fn(cur.cell.head) {
			return true
		}

		cur = cur.Tail()
	}

	return false
}

// Seq returns an iter.Seq that yields each element. Bridges to Go's range protocol.
// The returned closure captures the original stream head. During iteration, the
// loop variable advances without accumulating additional references, but the
// closure itself retains the stream root while it exists.
func (s Stream[T]) Seq() iter.Seq[T] {
	return func(yield func(T) bool) {
		cur := s
		for !cur.IsEmpty() {
			if !yield(cur.cell.head) {
				return
			}

			cur = cur.Tail()
		}
	}
}
