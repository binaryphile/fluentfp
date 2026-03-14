package stream

import "github.com/binaryphile/fluentfp/tuple/pair"

// Map applies fn to each element, returning a stream of a different type.
// Eagerly transforms the current head; tail transforms are deferred.
// Standalone because Go methods cannot introduce the extra type parameter R.
// Panics if fn is nil.
func Map[T, R any](s Stream[T], fn func(T) R) Stream[R] {
	if fn == nil {
		panic("stream.Map: fn must not be nil")
	}

	if s.cell == nil {
		return Stream[R]{}
	}

	head := fn(s.cell.head)
	cur := s

	return Stream[R]{cell: &cell[R]{
		head: head,
		tail: func() *cell[R] {
			result := Map(cur.Tail(), fn)
			return result.cell
		},
	}}
}

// FlatMap applies fn to each element of s and concatenates the resulting streams.
// Head-eager: scans forward to find the first non-empty inner stream.
// Standalone because Go methods cannot introduce the extra type parameter R.
// Panics if fn is nil.
func FlatMap[T, R any](s Stream[T], fn func(T) Stream[R]) Stream[R] {
	if fn == nil {
		panic("stream.FlatMap: fn must not be nil")
	}

	cur := s
	for !cur.IsEmpty() {
		inner := fn(cur.cell.head)
		if !inner.IsEmpty() {
			matchedOuter := cur
			matchedInner := inner
			return Stream[R]{cell: &cell[R]{
				head: matchedInner.cell.head,
				tail: func() *cell[R] {
					result := Concat(matchedInner.Tail(), FlatMap(matchedOuter.Tail(), fn))
					return result.cell
				},
			}}
		}
		cur = cur.Tail()
	}

	return Stream[R]{}
}

// Concat returns a stream that yields all elements of a followed by all elements of b.
func Concat[T any](a, b Stream[T]) Stream[T] {
	if a.IsEmpty() {
		return b
	}

	matched := a
	return Stream[T]{cell: &cell[T]{
		head: a.cell.head,
		tail: func() *cell[T] {
			result := Concat(matched.Tail(), b)
			return result.cell
		},
	}}
}

// Zip returns a stream of pairs from corresponding elements of a and b.
// Truncates to the shorter stream.
func Zip[A, B any](a Stream[A], b Stream[B]) Stream[pair.Pair[A, B]] {
	if a.IsEmpty() || b.IsEmpty() {
		return Stream[pair.Pair[A, B]]{}
	}

	matchedA := a
	matchedB := b
	return Stream[pair.Pair[A, B]]{cell: &cell[pair.Pair[A, B]]{
		head: pair.Of(a.cell.head, b.cell.head),
		tail: func() *cell[pair.Pair[A, B]] {
			result := Zip(matchedA.Tail(), matchedB.Tail())
			return result.cell
		},
	}}
}

// Scan reduces a stream like Fold, but returns a stream of all intermediate
// accumulator values. Includes the initial value as the first element (scanl semantics).
// Standalone because Go methods cannot introduce the extra type parameter R.
// Panics if fn is nil.
func Scan[T, R any](s Stream[T], initial R, fn func(R, T) R) Stream[R] {
	if fn == nil {
		panic("stream.Scan: fn must not be nil")
	}

	if s.IsEmpty() {
		return Stream[R]{cell: &cell[R]{
			head:  initial,
			state: cellForced,
		}}
	}

	matched := s
	return Stream[R]{cell: &cell[R]{
		head: initial,
		tail: func() *cell[R] {
			next := fn(initial, matched.cell.head)
			result := Scan(matched.Tail(), next, fn)
			return result.cell
		},
	}}
}

// Fold reduces a stream to a single value by applying fn progressively.
// Standalone because Go methods cannot introduce the extra type parameter R.
// Requires a finite stream. Panics if fn is nil.
func Fold[T, R any](s Stream[T], initial R, fn func(R, T) R) R {
	if fn == nil {
		panic("stream.Fold: fn must not be nil")
	}

	acc := initial
	cur := s

	for !cur.IsEmpty() {
		acc = fn(acc, cur.cell.head)
		cur = cur.Tail()
	}

	return acc
}
