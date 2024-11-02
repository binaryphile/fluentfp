package toboolmet

// SliceOf derives from slice.
// It is usable anywhere a slice is, but provides additional fluent fp methods.
type SliceOf[T comparable] []T

// Contains returns whether the slice contains the specified element.
func (ts SliceOf[T]) Contains(t T) bool {
	return ts.IndexOf(t) != -1
}

// Convert applies fn to each element of the slice, returning a new slice of the same element type with the results.
// It is the same as map, but without changing the element type of the slice.
func (ts SliceOf[T]) Convert(fn func(T) T) SliceOf[T] {
	results := make([]T, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// Each applies fn to each member of ts.
func (ts SliceOf[T]) Each(fn func(T)) {
	for _, t := range ts {
		fn(t)
	}
}

// IndexOf returns the index of the specified element in the slice, or -1 if not found.
func (ts SliceOf[T]) IndexOf(t T) int {
	for i, v := range ts {
		if t == v {
			return i
		}
	}

	return -1
}

// KeepIf returns the slice of elements from ts for which fn returns true.
func (ts SliceOf[T]) KeepIf(fn func(T) bool) SliceOf[T] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if fn(t) {
			results = append(results, t)
		}
	}

	return results
}

// Len returns the length of the slice.
func (ts SliceOf[T]) Len() int {
	return len(ts)
}

// RemoveIf returns the slice of elements from ts for which fn returns false.
func (ts SliceOf[T]) RemoveIf(fn func(T) bool) SliceOf[T] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if !fn(t) {
			results = append(results, t)
		}
	}

	return results
}

// TakeFirst returns the first n elements of ts.
func (ts SliceOf[T]) TakeFirst(n int) SliceOf[T] {
	if n > len(ts) {
		n = len(ts)
	}

	return ts[:n]
}

// ToBools applies the provided function to each element of the slice, mapping it to a Slice of type bool.
func (ts SliceOf[T]) ToBools(fn func(T) bool) SliceOf[bool] {
	return Map(ts, fn)
}
