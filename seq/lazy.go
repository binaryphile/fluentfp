package seq

// KeepIf returns a Seq containing only elements where fn returns true.
// Panics if fn is nil.
func (s Seq[T]) KeepIf(fn func(T) bool) Seq[T] {
	if fn == nil {
		panic("seq.KeepIf: fn must not be nil")
	}

	if s == nil {
		return Empty[T]()
	}

	return Seq[T](func(yield func(T) bool) {
		for v := range s {
			if fn(v) && !yield(v) {
				return
			}
		}
	})
}

// RemoveIf returns a Seq containing only elements where fn returns false.
// It is the complement of KeepIf.
// Panics if fn is nil.
func (s Seq[T]) RemoveIf(fn func(T) bool) Seq[T] {
	if fn == nil {
		panic("seq.RemoveIf: fn must not be nil")
	}

	if s == nil {
		return Empty[T]()
	}

	return Seq[T](func(yield func(T) bool) {
		for v := range s {
			if !fn(v) && !yield(v) {
				return
			}
		}
	})
}

// Transform applies fn to each element, returning a Seq of results.
// Same-type transform — use standalone Map for cross-type mapping.
// Panics if fn is nil.
func (s Seq[T]) Transform(fn func(T) T) Seq[T] {
	if fn == nil {
		panic("seq.Transform: fn must not be nil")
	}

	if s == nil {
		return Empty[T]()
	}

	return Seq[T](func(yield func(T) bool) {
		for v := range s {
			if !yield(fn(v)) {
				return
			}
		}
	})
}

// Take returns a Seq yielding at most n elements.
// If n <= 0, yields nothing.
func (s Seq[T]) Take(n int) Seq[T] {
	if s == nil || n <= 0 {
		return Empty[T]()
	}

	return Seq[T](func(yield func(T) bool) {
		remaining := n

		for v := range s {
			remaining--

			if !yield(v) || remaining <= 0 {
				return
			}
		}
	})
}

// Drop returns a Seq that skips the first n elements.
// If n <= 0, yields all elements.
func (s Seq[T]) Drop(n int) Seq[T] {
	if s == nil {
		return Empty[T]()
	}

	if n <= 0 {
		return s
	}

	return Seq[T](func(yield func(T) bool) {
		skipped := 0

		for v := range s {
			if skipped < n {
				skipped++
				continue
			}

			if !yield(v) {
				return
			}
		}
	})
}

// TakeWhile returns a Seq yielding elements while fn returns true.
// Stops at the first element where fn returns false.
// Panics if fn is nil.
func (s Seq[T]) TakeWhile(fn func(T) bool) Seq[T] {
	if fn == nil {
		panic("seq.TakeWhile: fn must not be nil")
	}

	if s == nil {
		return Empty[T]()
	}

	return Seq[T](func(yield func(T) bool) {
		for v := range s {
			if !fn(v) || !yield(v) {
				return
			}
		}
	})
}

// Intersperse inserts sep between every adjacent pair of elements.
// Empty and single-element sequences pass through unchanged.
// Fully streaming with O(1) state. Works with infinite sequences.
func (s Seq[T]) Intersperse(sep T) Seq[T] {
	if s == nil {
		return Empty[T]()
	}

	return Seq[T](func(yield func(T) bool) {
		first := true

		for v := range s {
			if !first {
				if !yield(sep) {
					return
				}
			}

			if !yield(v) {
				return
			}

			first = false
		}
	})
}

// DropWhile returns a Seq that skips elements while fn returns true,
// then yields the rest.
// Panics if fn is nil.
func (s Seq[T]) DropWhile(fn func(T) bool) Seq[T] {
	if fn == nil {
		panic("seq.DropWhile: fn must not be nil")
	}

	if s == nil {
		return Empty[T]()
	}

	return Seq[T](func(yield func(T) bool) {
		dropping := true

		for v := range s {
			if dropping {
				if fn(v) {
					continue
				}

				dropping = false
			}

			if !yield(v) {
				return
			}
		}
	})
}
