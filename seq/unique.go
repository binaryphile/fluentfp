package seq

// Unique removes duplicate elements lazily, preserving first occurrence.
// Memory grows with the number of distinct elements seen — on infinite or
// high-cardinality streams, memory may grow without bound.
// On infinite repeating streams, Unique stalls once all distinct values have been
// emitted — requesting more elements than distinct values exist will never terminate.
// Note: for float types, NaN != NaN, so NaN values are never deduplicated.
// Standalone because the comparable constraint cannot be expressed on the Seq[T any] receiver.
func Unique[T comparable](s Seq[T]) Seq[T] {
	if s == nil {
		return Empty[T]()
	}

	return Seq[T](func(yield func(T) bool) {
		seen := make(map[T]struct{})

		for v := range s {
			if _, exists := seen[v]; exists {
				continue
			}

			seen[v] = struct{}{}

			if !yield(v) {
				return
			}
		}
	})
}

// UniqueBy removes duplicate elements lazily by extracted key, preserving first occurrence.
// Memory grows with the number of distinct keys seen — on infinite or
// high-cardinality streams, memory may grow without bound.
// Note: for float key types, NaN != NaN, so NaN keys are never deduplicated.
// Standalone because the comparable constraint on K cannot be expressed on the Seq[T any] receiver.
// Panics if fn is nil.
func UniqueBy[T any, K comparable](s Seq[T], fn func(T) K) Seq[T] {
	if fn == nil {
		panic("seq.UniqueBy: fn must not be nil")
	}

	if s == nil {
		return Empty[T]()
	}

	return Seq[T](func(yield func(T) bool) {
		seen := make(map[K]struct{})

		for v := range s {
			k := fn(v)
			if _, exists := seen[k]; exists {
				continue
			}

			seen[k] = struct{}{}

			if !yield(v) {
				return
			}
		}
	})
}
