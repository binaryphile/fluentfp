package toint32met

// MappableSliceOf derives from slice.
// It is usable anywhere a slice is, but provides additional fluent fp methods.
type MappableSliceOf[T comparable, R any] []T

// Contains returns whether the slice contains the specified element.
func (ts MappableSliceOf[T, R]) Contains(t T) bool {
	return ts.IndexOf(t) != -1
}

// Convert applies fn to each element of the slice, returning a new slice of the same element type with the results.
// It is the same as map, but without changing the element type of the slice.
func (ts MappableSliceOf[T, R]) Convert(fn func(T) T) MappableSliceOf[T, R] {
	results := make([]T, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// Each applies fn to each member of ts.
func (ts MappableSliceOf[T, R]) Each(fn func(T)) {
	for _, t := range ts {
		fn(t)
	}
}

// IndexOf returns the index of the specified element in the slice, or -1 if not found.
func (ts MappableSliceOf[T, R]) IndexOf(t T) int {
	for i, v := range ts {
		if t == v {
			return i
		}
	}

	return -1
}

// KeepIf returns the slice of elements from ts for which fn returns true.
func (ts MappableSliceOf[T, R]) KeepIf(fn func(T) bool) MappableSliceOf[T, R] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if fn(t) {
			results = append(results, t)
		}
	}

	return results
}

// Map applies fn to each element of the slice, returning a slice of the return type with the results.
func (ts MappableSliceOf[T, R]) Map(fn func(T) R) SliceOf[T] {
	results := make([]T, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// Len returns the length of the slice.
func (ts MappableSliceOf[T, R]) Len() int {
	return len(ts)
}

// RemoveIf returns the slice of elements from ts for which fn returns false.
func (ts MappableSliceOf[T, R]) RemoveIf(fn func(T) bool) MappableSliceOf[T, R] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if !fn(t) {
			results = append(results, t)
		}
	}

	return results
}

// TakeFirst returns the first n elements of ts.
func (ts MappableSliceOf[T, R]) TakeFirst(n int) MappableSliceOf[T, R] {
	if n > len(ts) {
		n = len(ts)
	}

	return ts[:n]
}

// ToInt32s applies the provided function to each element of the slice, mapping it to a Slice of type int32.
func (ts MappableSliceOf[T, R]) ToInt32s(fn func(T) int32) MappableSliceOf[int32, R] {
	return Map(ts, fn)
}
