package seq

import (
	"slices"

	"github.com/binaryphile/fluentfp/option"
)

// Collect materializes the Seq into a slice. Requires a finite sequence.
func (s Seq[T]) Collect() []T {
	if s == nil {
		return nil
	}

	return slices.Collect(s.Iter())
}

// Find returns the first element where fn returns true.
// Short-circuits on first match.
// Panics if fn is nil.
func (s Seq[T]) Find(fn func(T) bool) option.Option[T] {
	if fn == nil {
		panic("seq.Find: fn must not be nil")
	}

	if s == nil {
		return option.Option[T]{}
	}

	for v := range s {
		if fn(v) {
			return option.Of(v)
		}
	}

	return option.Option[T]{}
}

// Any returns true if fn returns true for at least one element.
// Short-circuits on first match.
// Panics if fn is nil.
func (s Seq[T]) Any(fn func(T) bool) bool {
	if fn == nil {
		panic("seq.Any: fn must not be nil")
	}

	if s == nil {
		return false
	}

	for v := range s {
		if fn(v) {
			return true
		}
	}

	return false
}

// Every returns true if fn returns true for all elements.
// Returns true for an empty Seq (vacuous truth).
// Short-circuits on first mismatch.
// Panics if fn is nil.
func (s Seq[T]) Every(fn func(T) bool) bool {
	if fn == nil {
		panic("seq.Every: fn must not be nil")
	}

	if s == nil {
		return true
	}

	for v := range s {
		if !fn(v) {
			return false
		}
	}

	return true
}

// None returns true if fn returns false for all elements.
// Returns true for an empty Seq (vacuous truth).
// Panics if fn is nil.
func (s Seq[T]) None(fn func(T) bool) bool {
	if fn == nil {
		panic("seq.None: fn must not be nil")
	}

	if s == nil {
		return true
	}

	for v := range s {
		if fn(v) {
			return false
		}
	}

	return true
}

// Each applies fn to every element for side effects. Requires a finite sequence.
// Panics if fn is nil.
func (s Seq[T]) Each(fn func(T)) {
	if fn == nil {
		panic("seq.Each: fn must not be nil")
	}

	if s == nil {
		return
	}

	for v := range s {
		fn(v)
	}
}
