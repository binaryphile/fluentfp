package stream

// KeepIf returns a stream containing only elements where fn returns true.
// Eagerly scans to the first match (head-eager constraint); the rest is lazy.
// On an infinite stream with no matching element, this will not terminate.
// Panics if fn is nil.
func (s Stream[T]) KeepIf(fn func(T) bool) Stream[T] {
	if fn == nil {
		panic("stream.KeepIf: fn must not be nil")
	}

	cur := s
	for !cur.IsEmpty() {
		if fn(cur.cell.head) {
			matched := cur

			return Stream[T]{cell: &cell[T]{
				head: matched.cell.head,
				tail: func() *cell[T] {
					result := matched.Tail().KeepIf(fn)
					return result.cell
				},
			}}
		}

		cur = cur.Tail()
	}

	return Stream[T]{}
}

// RemoveIf returns a stream containing only elements where fn returns false.
// It is the complement of KeepIf.
// Panics if fn is nil.
func (s Stream[T]) RemoveIf(fn func(T) bool) Stream[T] {
	if fn == nil {
		panic("stream.RemoveIf: fn must not be nil")
	}

	notFn := func(v T) bool { return !fn(v) }

	return s.KeepIf(notFn)
}

// Transform applies fn to each element, returning a stream of results.
// Eagerly transforms the current head; tail transforms are deferred.
// Same-type transform — use standalone Map for cross-type mapping.
// Panics if fn is nil.
func (s Stream[T]) Transform(fn func(T) T) Stream[T] {
	if fn == nil {
		panic("stream.Transform: fn must not be nil")
	}

	if s.cell == nil {
		return Stream[T]{}
	}

	head := fn(s.cell.head)
	cur := s

	return Stream[T]{cell: &cell[T]{
		head: head,
		tail: func() *cell[T] {
			result := cur.Tail().Transform(fn)
			return result.cell
		},
	}}
}

// Take returns a stream of at most n elements. Negative n returns empty.
func (s Stream[T]) Take(n int) Stream[T] {
	if n <= 0 || s.cell == nil {
		return Stream[T]{}
	}

	if n == 1 {
		return Stream[T]{cell: &cell[T]{head: s.cell.head}}
	}

	cur := s
	remaining := n - 1

	return Stream[T]{cell: &cell[T]{
		head: cur.cell.head,
		tail: func() *cell[T] {
			result := cur.Tail().Take(remaining)
			return result.cell
		},
	}}
}

// TakeWhile returns elements while fn returns true. Stops at the first false.
// Evaluates the predicate on the current head immediately; tail evaluation is deferred.
// Panics if fn is nil.
func (s Stream[T]) TakeWhile(fn func(T) bool) Stream[T] {
	if fn == nil {
		panic("stream.TakeWhile: fn must not be nil")
	}

	if s.cell == nil {
		return Stream[T]{}
	}

	if !fn(s.cell.head) {
		return Stream[T]{}
	}

	cur := s

	return Stream[T]{cell: &cell[T]{
		head: cur.cell.head,
		tail: func() *cell[T] {
			result := cur.Tail().TakeWhile(fn)
			return result.cell
		},
	}}
}

// Drop skips the first n elements. Forces skipped cells eagerly.
// Negative n returns the stream unchanged.
func (s Stream[T]) Drop(n int) Stream[T] {
	cur := s
	for i := 0; i < n && !cur.IsEmpty(); i++ {
		cur = cur.Tail()
	}

	return cur
}

// DropWhile skips elements while fn returns true. Forces skipped cells eagerly.
// On an infinite stream where fn always returns true, this will not terminate.
// Panics if fn is nil.
func (s Stream[T]) DropWhile(fn func(T) bool) Stream[T] {
	if fn == nil {
		panic("stream.DropWhile: fn must not be nil")
	}

	cur := s
	for !cur.IsEmpty() && fn(cur.cell.head) {
		cur = cur.Tail()
	}

	return cur
}
