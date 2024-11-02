package toboolmet

// MappableRawSliceOf derives from slice.
// It is usable anywhere a slice is, but provides additional fluent fp methods.
// It is raw because it takes any type, so it is unable to provide methods like Contains,
// which relies on comparability.
type MappableRawSliceOf[T any, R any] []T

// Each applies fn to each member of ts.
func (ts MappableRawSliceOf[T, R]) Each(fn func(T)) {
	for _, t := range ts {
		fn(t)
	}
}

// KeepIf returns the slice of elements from ts for which fn returns true.
func (ts MappableRawSliceOf[T, R]) KeepIf(fn func(T) bool) MappableRawSliceOf[T, R] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if fn(t) {
			results = append(results, t)
		}
	}

	return results
}

// Len returns the length of the slice.
func (ts MappableRawSliceOf[T, R]) Len() int {
	return len(ts)
}

// Map applies fn to each element of the slice, returning a slice of the return type with the results.
func (ts MappableRawSliceOf[T, R]) Map(fn func(T) R) RawSliceOf[T] {
	results := make([]T, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// RemoveIf returns the slice of elements from ts for which fn returns false.
func (ts MappableRawSliceOf[T, R]) RemoveIf(fn func(T) bool) MappableRawSliceOf[T, R] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if !fn(t) {
			results = append(results, t)
		}
	}

	return results
}

// TakeFirst returns the first n elements of ts.
func (ts MappableRawSliceOf[T, R]) TakeFirst(n int) MappableRawSliceOf[T, R] {
	if n > len(ts) {
		n = len(ts)
	}

	return ts[:n]
}

// ToBools applies the provided function to each element of the slice, mapping it to a Slice of type bool.
func (ts MappableRawSliceOf[T, R]) ToBools(fn func(T) bool) MappableSliceOf[bool, R] {
	return Map(ts, fn)
}
