package stream

// From creates a lazy stream from a slice. Each tail closure captures a subslice
// view of the original — the backing array may be retained until those closures
// are evaluated or become unreachable.
func From[T any](ts []T) Stream[T] {
	if len(ts) == 0 {
		return Stream[T]{}
	}

	rest := ts[1:]

	return Stream[T]{cell: &cell[T]{
		head: ts[0],
		tail: func() *cell[T] {
			s := From(rest)
			return s.cell
		},
	}}
}

// Of creates a lazy stream from variadic arguments. Delegates to From.
func Of[T any](vs ...T) Stream[T] {
	return From(vs)
}

// Generate creates an infinite stream: seed, fn(seed), fn(fn(seed)), ...
// The seed is the first element (eager); subsequent elements apply fn lazily.
// Panics if fn is nil.
func Generate[T any](seed T, fn func(T) T) Stream[T] {
	if fn == nil {
		panic("stream.Generate: fn must not be nil")
	}

	return Stream[T]{cell: &cell[T]{
		head: seed,
		tail: func() *cell[T] {
			s := Generate(fn(seed), fn)
			return s.cell
		},
	}}
}

// Repeat creates an infinite stream where every element is v.
// Uses a self-referencing cell — O(1) memory regardless of traversal length.
func Repeat[T any](v T) Stream[T] {
	c := &cell[T]{
		head:  v,
		state: cellForced,
	}
	c.next = c

	return Stream[T]{cell: c}
}

// Unfold creates a stream by repeatedly applying a step function to a seed.
// Each call to fn returns (element, nextSeed, ok). When ok is false, the stream
// ends. Unfold is the dual of Fold: Fold consumes a stream to a value, Unfold
// produces a stream from a value.
//
// The first step is evaluated eagerly — a panic on the first call fails at
// construction and is not retryable. Subsequent steps are lazy and retryable.
func Unfold[T, S any](seed S, fn func(S) (T, S, bool)) Stream[T] {
	if fn == nil {
		panic("stream.Unfold: fn must not be nil")
	}

	v, next, ok := fn(seed)
	if !ok {
		return Stream[T]{}
	}

	return Stream[T]{cell: &cell[T]{
		head: v,
		tail: func() *cell[T] {
			s := Unfold(next, fn)
			return s.cell
		},
	}}
}
