package mappable

import (
	"bitbucket.org/accelecon/charybdis/tools/avwob2drm/pkg/option"
)

type SliceWithMapTo[T comparable, R any] []T

// KeepIf returns the slice of elements from ts for which fn returns true.  It's the same as Filter would be.
func (ts SliceWithMapTo[T, R]) KeepIf(fn func(T) bool) SliceWithMapTo[T, R] {
	results := make([]T, 0, len(ts))

	for _, t := range ts {
		if fn(t) {
			results = append(results, t)
		}
	}

	return results
}

func (ts SliceWithMapTo[T, R]) Map(fn func(T) T) SliceWithMapTo[T, R] {
	results := make([]T, len(ts))

	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

func (ts SliceWithMapTo[T, R]) MapTo(fn func(T) R) RawSliceOf[R] {
	results := make([]R, len(ts))

	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// MapToBool returns the slice resulting from applying fn, whose return type is bool, to each member of ts.
func (ts SliceWithMapTo[T, R]) MapToBool(fn func(T) bool) SliceWithMapTo[bool, R] {
	results := make([]bool, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToInt returns the slice resulting from applying fn, whose return type is int, to each member of ts.
func (ts SliceWithMapTo[T, R]) MapToInt(fn func(T) int) SliceWithMapTo[int, R] {
	results := make([]int, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToStr returns the slice resulting from applying fn, whose return type is string, to each member of ts.
func (ts SliceWithMapTo[T, R]) MapToStr(fn func(T) string) SliceWithMapTo[string, R] {
	results := make([]string, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToStrOption returns the slice resulting from applying fn, whose return type is option.String, to each member of ts.
func (ts SliceWithMapTo[T, R]) MapToStrOption(fn func(T) option.String) SliceWithMapTo[option.String, R] {
	results := make([]option.String, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToStrSlice returns the slice resulting from applying fn, whose return type is []string, to each member of ts.
func (ts SliceWithMapTo[T, R]) MapToStrSlice(fn func(T) []string) RawSliceTo[[]string, R] {
	results := make([][]string, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// RemoveIf returns the slice of elements from ts for which fn returns false.
// It's the negation of what Filter would be, since it's not easy to write an in-line function for negation in Go.
func (ts SliceWithMapTo[T, R]) RemoveIf(fn func(T) bool) SliceWithMapTo[T, R] {
	results := make([]T, 0, len(ts))

	for _, t := range ts {
		if !fn(t) {
			results = append(results, t)
		}
	}

	return results
}
