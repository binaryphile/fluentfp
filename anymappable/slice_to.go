package anymappable

// SliceTo is an extended slice that specifies the return type used by the MapTo method.
// That is in addition to the usual methods of SliceOf.
type SliceTo[T, R any] []T

// KeepIf returns the slice of elements from ts for which fn returns true.
func (ts SliceTo[T, R]) KeepIf(fn func(T) bool) SliceTo[T, R] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if fn(t) {
			results = append(results, t)
		}
	}

	return results
}

// Map returns the slice resulting from applying fn, whose return type is the same as the elements of ts, to each member of ts.
func (ts SliceTo[T, R]) Map(fn func(T) T) SliceTo[T, R] {
	results := make([]T, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// MapTo returns the slice resulting from applying fn, whose return type is R, to each member of ts.
// Since there is no new target type for a new SliceTo, the return type is SliceOf.
func (ts SliceTo[T, R]) MapTo(fn func(T) R) SliceOf[R] {
	results := make([]R, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// MapToBool returns the slice resulting from applying fn, whose return type is bool, to each member of ts.
// The return type is SliceTo, preserving the originally-specified return type.
func (ts SliceTo[T, R]) MapToBool(fn func(T) bool) SliceTo[bool, R] {
	results := make([]bool, len(ts))
	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToInt returns the slice resulting from applying fn, whose return type is int, to each member of ts.
// The return type is SliceTo, preserving the originally-specified return type.
func (ts SliceTo[T, R]) MapToInt(fn func(T) int) SliceTo[int, R] {
	results := make([]int, len(ts))
	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToStrSlice returns the slice resulting from applying fn, whose return type is []string, to each member of ts.
// The return type is SliceTo, preserving the originally-specified return type.
func (ts SliceTo[T, R]) MapToStrSlice(fn func(T) []string) SliceTo[[]string, R] {
	results := make([][]string, len(ts))
	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToStr returns the slice resulting from applying fn, whose return type is string, to each member of ts.
// The return type is SliceTo, preserving the originally-specified return type.
func (ts SliceTo[T, R]) MapToStr(fn func(T) string) SliceTo[string, R] {
	results := make([]string, len(ts))
	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// RemoveIf returns the slice of elements from ts for which fn returns false.
func (ts SliceTo[T, R]) RemoveIf(fn func(T) bool) SliceTo[T, R] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if fn(t) {
			results = append(results, t)
		}
	}

	return results
}
