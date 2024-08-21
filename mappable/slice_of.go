package mappable

import (
	"bitbucket.org/accelecon/charybdis/tools/avwob2drm/pkg/option"
)

type SliceOf[T comparable] []T

func (ts SliceOf[T]) Contains(t T) bool {
	return ts.Index(t) != -1
}

func (ts SliceOf[T]) Index(t T) int {
	for i := range ts {
		if t == ts[i] {
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

func (ts SliceOf[T]) Map(fn func(T) T) SliceOf[T] {
	results := make([]T, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToBool returns the slice resulting from applying fn, whose return type is bool, to each member of ts.
func (ts SliceOf[T]) MapToBool(fn func(T) bool) SliceOfBools {
	results := make([]bool, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToInt returns the slice resulting from applying fn, whose return type is int, to each member of ts.
func (ts SliceOf[T]) MapToInt(fn func(T) int) SliceOfInts {
	results := make([]int, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToStr returns the slice resulting from applying fn, whose return type is string, to each member of ts.
func (ts SliceOf[T]) MapToStr(fn func(T) string) SliceOfStrings {
	results := make([]string, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToStrOption returns the slice resulting from applying fn, whose return type is option.String, to each member of ts.
func (ts SliceOf[T]) MapToStrOption(fn func(T) option.String) SliceOf[option.String] {
	results := make([]option.String, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToStrSlice returns the slice resulting from applying fn, whose return type is []string, to each member of ts.
func (ts SliceOf[T]) MapToStrSlice(fn func(T) []string) SliceOfStrSlices {
	results := make([][]string, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
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
