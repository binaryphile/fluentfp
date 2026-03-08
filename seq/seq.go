package seq

import (
	"iter"
	"slices"
)

// Seq is a lazy iterator over iter.Seq[T] with method chaining.
// Zero value yields no elements. Range works directly:
//
//	for v := range seq.From(data).KeepIf(pred) { ... }
//
// Unlike stream.Stream (memoized), Seq pipelines re-evaluate on each
// Collect or range — standard iter.Seq behavior.
//
// Use .Iter() when a function expects iter.Seq[T].
type Seq[T any] iter.Seq[T]

// Iter returns the underlying iter.Seq[T] for interop with stdlib
// and other libraries.
func (s Seq[T]) Iter() iter.Seq[T] {
	return iter.Seq[T](s)
}

// From creates a Seq from a slice.
// Returns nil for nil or empty input.
func From[T any](ts []T) Seq[T] {
	if len(ts) == 0 {
		return nil
	}

	return Seq[T](slices.Values(ts))
}

// Of creates a Seq from variadic arguments.
func Of[T any](vs ...T) Seq[T] {
	return From(vs)
}

// FromIter wraps an existing iter.Seq[T] as a Seq for method chaining.
func FromIter[T any](s iter.Seq[T]) Seq[T] {
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
