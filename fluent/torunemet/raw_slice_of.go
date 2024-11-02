package torunemet

// RawSliceOf derives from slice.
// It is usable anywhere a slice is, but provides additional fluent fp methods.
// It is raw because it takes any type, so it is unable to provide methods like Contains,
// which relies on comparability.
type RawSliceOf[T any] []T

// Each applies fn to each member of ts.
func (ts RawSliceOf[T]) Each(fn func(T)) {
	for _, t := range ts {
		fn(t)
	}
}

// KeepIf returns the slice of elements from ts for which fn returns true.
func (ts RawSliceOf[T]) KeepIf(fn func(T) bool) RawSliceOf[T] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if fn(t) {
			results = append(results, t)
		}
	}

	return results
}

// Len returns the length of the slice.
func (ts RawSliceOf[T]) Len() int {
	return len(ts)
}

// Map applies fn to each element of the slice, returning a new slice of the same element type with the results.
// It is the same as Map, but without changing the element type of the slice.
func (ts RawSliceOf[T]) Map(fn func(T) T) RawSliceOf[T] {
	results := make([]T, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// RemoveIf returns the slice of elements from ts for which fn returns false.
func (ts RawSliceOf[T]) RemoveIf(fn func(T) bool) RawSliceOf[T] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if !fn(t) {
			results = append(results, t)
		}
	}

	return results
}

// TakeFirst returns the first n elements of ts.
func (ts RawSliceOf[T]) TakeFirst(n int) RawSliceOf[T] {
	if n > len(ts) {
		n = len(ts)
	}

	return ts[:n]
}

// ToRunes applies the provided function to each element of the slice, mapping it to a Slice of type rune.
func (ts RawSliceOf[T]) ToRunes(fn func(T) rune) SliceOf[rune] {
	return Map(ts, fn)
}
