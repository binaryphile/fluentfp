package seq

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
