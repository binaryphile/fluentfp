package seq

import (
	"iter"
	"slices"
)

// Seq is a lazy iterator over iter.Seq[T] with method chaining.
// Range works directly:
//
//	for v := range seq.From(data).KeepIf(pred) { ... }
//
// Unlike stream.Stream (memoized), Seq pipelines re-evaluate on each
// Collect or range — standard iter.Seq behavior.
//
// The zero value (nil) is NOT safe for direct range — it will panic.
// Use [Empty], [From], or other constructors. All constructors and
// Seq-returning operations return non-nil Seqs safe for range.
//
// Use .Iter() when a function expects iter.Seq[T].
type Seq[T any] iter.Seq[T]

// Iter returns the underlying iter.Seq[T] for interop with stdlib
// and other libraries. Returns a no-op iterator if s is nil (zero value).
func (s Seq[T]) Iter() iter.Seq[T] {
	if s == nil {
		return iter.Seq[T](Empty[T]())
	}

	return iter.Seq[T](s)
}

// Empty returns a Seq that yields no elements. Safe for direct range.
func Empty[T any]() Seq[T] {
	return Seq[T](func(func(T) bool) {})
}

// From creates a Seq from a slice.
// Returns an empty Seq for nil or empty input.
func From[T any](ts []T) Seq[T] {
	if len(ts) == 0 {
		return Empty[T]()
	}

	return Seq[T](slices.Values(ts))
}

// Of creates a Seq from variadic arguments.
func Of[T any](vs ...T) Seq[T] {
	return From(vs)
}

// FromIter wraps an existing iter.Seq[T] as a Seq for method chaining.
// If s is nil, returns an empty Seq safe for range.
func FromIter[T any](s iter.Seq[T]) Seq[T] {
	if s == nil {
		return Empty[T]()
	}

	return Seq[T](s)
}

// Generate creates an infinite Seq: seed, fn(seed), fn(fn(seed)), ...
// Panics if fn is nil.
func Generate[T any](seed T, fn func(T) T) Seq[T] {
	if fn == nil {
		panic("seq.Generate: fn must not be nil")
	}

	return Seq[T](func(yield func(T) bool) {
		v := seed
		for {
			if !yield(v) {
				return
			}

			v = fn(v)
		}
	})
}

// Repeat creates an infinite Seq that yields v forever.
func Repeat[T any](v T) Seq[T] {
	return Seq[T](func(yield func(T) bool) {
		for {
			if !yield(v) {
				return
			}
		}
	})
}

// Unfold creates a Seq by repeatedly applying fn to a seed state.
// fn returns (element, nextState, continue). When continue is false,
// the sequence ends without emitting element. Fully lazy — fn is not
// called until iteration begins (unlike stream.Unfold, which eagerly
// evaluates the first element at construction time).
// Panics if fn is nil.
func Unfold[T, S any](seed S, fn func(S) (T, S, bool)) Seq[T] {
	if fn == nil {
		panic("seq.Unfold: fn must not be nil")
	}

	return Seq[T](func(yield func(T) bool) {
		state := seed
		for {
			v, next, ok := fn(state)
			if !ok {
				return
			}

			if !yield(v) {
				return
			}

			state = next
		}
	})
}

// FromNext creates a Seq from a pull-style iterator function.
// next is called repeatedly; each call returns (value, ok).
// When ok is false the sequence ends. next is not called again after
// returning false.
//
// Because next is typically stateful (a cursor), iterating the returned
// Seq a second time will see whatever state next is in — usually exhausted.
// Use .Collect() to materialize if you need multiple passes.
// Panics if next is nil.
func FromNext[T any](next func() (T, bool)) Seq[T] {
	if next == nil {
		panic("seq.FromNext: next must not be nil")
	}

	return Seq[T](func(yield func(T) bool) {
		for {
			v, ok := next()
			if !ok {
				return
			}

			if !yield(v) {
				return
			}
		}
	})
}
