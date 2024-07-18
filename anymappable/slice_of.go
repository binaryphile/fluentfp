package anymappable

import "github.com/binaryphile/funcTrunk/option"

// SliceOf
type SliceOf[T any] []T

// KeepIf returns the slice of elements from ts for which fn returns true.  It's the same as Filter would be.
func (ts SliceOf[T]) KeepIf(fn func(T) bool) SliceOf[T] {
	results := make([]T, 0, len(ts))

	for _, t := range ts {
		if fn(t) {
			results = append(results, t)
		}
	}

	return results
}

// Map returns the slice resulting from applying fn, whose return type is the same as the elements of ts, to each member of ts.
func (ts SliceOf[T]) Map(fn func(T) T) SliceOf[T] {
	results := make([]T, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToBool returns the slice resulting from applying fn, whose return type is bool, to each member of ts.
func (ts SliceOf[T]) MapToBool(fn func(T) bool) SliceOf[bool] {
	results := make([]bool, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToInt returns the slice resulting from applying fn, whose return type is int, to each member of ts.
func (ts SliceOf[T]) MapToInt(fn func(T) int) SliceOf[int] {
	results := make([]int, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToStr returns the slice resulting from applying fn, whose return type is string, to each member of ts.
func (ts SliceOf[T]) MapToStr(fn func(T) string) SliceOf[string] {
	results := make([]string, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToStrOption returns the slice resulting from applying fn, whose return type is string, to each member of ts.
func (ts SliceOf[T]) MapToStrOption(fn func(T) option.String) SliceOf[option.String] {
	results := make([]option.String, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToStrSlice returns the slice resulting from applying fn, whose return type is []string, to each member of ts.
func (ts SliceOf[T]) MapToStrSlice(fn func(T) []string) SliceOfStrSlice {
	results := make([][]string, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// RemoveIf returns the slice of elements from ts for which fn returns false.
// It's the negation of what Filter would be, since it's not easy to write an in-line function for negation in Go.
func (ts SliceOf[T]) RemoveIf(fn func(T) bool) SliceOf[T] {
	results := make([]T, 0, len(ts))

	for _, t := range ts {
		if !fn(t) {
			results = append(results, t)
		}
	}

	return results
}
