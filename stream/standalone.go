package stream

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
