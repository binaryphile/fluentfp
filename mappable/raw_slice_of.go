package mappable

import (
	"bitbucket.org/accelecon/charybdis/tools/avwob2drm/pkg/option"
)

type RawSliceOf[T any] []T

// ForEach applies fn to each element of ts.
func (ts RawSliceOf[T]) ForEach(fn func(T)) {
	for _, t := range ts {
		fn(t)
	}
}

// KeepIf returns the slice of elements from ts for which fn returns true.  It's the same as Filter would be.
func (ts RawSliceOf[T]) KeepIf(fn func(T) bool) RawSliceOf[T] {
	results := make([]T, 0, len(ts))

	for _, t := range ts {
		if fn(t) {
			results = append(results, t)
		}
	}

	return results
}

// Map returns the slice resulting from applying fn, whose return type is the same as the elements of ts, to each member of ts.
func (ts RawSliceOf[T]) Map(fn func(T) T) RawSliceOf[T] {
	results := make([]T, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToBool returns the slice resulting from applying fn, whose return type is bool, to each member of ts.
func (ts RawSliceOf[T]) MapToBool(fn func(T) bool) RawSliceOf[bool] {
	results := make([]bool, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToInt returns the slice resulting from applying fn, whose return type is int, to each member of ts.
func (ts RawSliceOf[T]) MapToInt(fn func(T) int) RawSliceOf[int] {
	results := make([]int, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToInt64 returns the slice resulting from applying fn, whose return type is int64, to each member of ts.
func (ts RawSliceOf[T]) MapToInt64(fn func(T) int64) RawSliceOf[int64] {
	results := make([]int64, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToString returns the slice resulting from applying fn, whose return type is string, to each member of ts.
func (ts RawSliceOf[T]) MapToString(fn func(T) string) RawSliceOf[string] {
	results := make([]string, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToSliceOfStrings returns the slice resulting from applying fn, whose return type is []string, to each member of ts.
func (ts RawSliceOf[T]) MapToSliceOfStrings(fn func(T) []string) SliceOfStringSlices {
	results := make([][]string, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToStringOption returns the slice resulting from applying fn, whose return type is string, to each member of ts.
func (ts RawSliceOf[T]) MapToStringOption(fn func(T) option.String) RawSliceOf[option.String] {
	results := make([]option.String, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// RemoveIf returns the slice of elements from ts for which fn returns false.
// It's the negation of what Filter would be, since it's not easy to write an in-line function for negation in Go.
func (ts RawSliceOf[T]) RemoveIf(fn func(T) bool) RawSliceOf[T] {
	results := make([]T, 0, len(ts))

	for _, t := range ts {
		if !fn(t) {
			results = append(results, t)
		}
	}

	return results
}
