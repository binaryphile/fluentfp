package fluent

type RawSliceFromTo[T, R any] []T

// ForEach applies fn to each member of ts.
func (ts RawSliceFromTo[T, R]) ForEach(fn func(T)) {
	for _, t := range ts {
		fn(t)
	}
}

// KeepIf returns the slice of elements from ts for which fn returns true.
func (ts RawSliceFromTo[T, R]) KeepIf(fn func(T) bool) RawSliceFromTo[T, R] {
	results := make([]T, 0, len(ts))

	for _, t := range ts {
		if fn(t) {
			results = append(results, t)
		}
	}

	return results
}

func (ts RawSliceFromTo[T, R]) MapTo(fn func(T) R) RawSliceOf[R] {
	results := make([]R, len(ts))

	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// MapToBool returns the slice resulting from applying fn, whose return type is bool, to each member of ts.
func (ts RawSliceFromTo[T, R]) MapToBool(fn func(T) bool) RawSliceFromTo[bool, R] {
	results := make([]bool, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToInt returns the slice resulting from applying fn, whose return type is int, to each member of ts.
func (ts RawSliceFromTo[T, R]) MapToInt(fn func(T) int) RawSliceFromTo[int, R] {
	results := make([]int, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToSliceOfStrings returns the slice resulting from applying fn, whose return type is []string, to each member of ts.
func (ts RawSliceFromTo[T, R]) MapToSliceOfStrings(fn func(T) []string) RawSliceFromTo[[]string, R] {
	results := make([][]string, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToString returns the slice resulting from applying fn, whose return type is string, to each member of ts.
func (ts RawSliceFromTo[T, R]) MapToString(fn func(T) string) RawSliceFromTo[string, R] {
	results := make([]string, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// RemoveIf returns the slice of elements from ts for which fn returns false.
func (ts RawSliceFromTo[T, R]) RemoveIf(fn func(T) bool) RawSliceFromTo[T, R] {
	results := make([]T, 0, len(ts))

	for _, t := range ts {
		if !fn(t) {
			results = append(results, t)
		}
	}

	return results
}
