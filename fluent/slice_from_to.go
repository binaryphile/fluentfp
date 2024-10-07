package fluent

import (
	"github.com/binaryphile/fluentfp/option"
)

type SliceFromTo[T comparable, R any] []T

func (ts SliceFromTo[T, R]) Contains(t T) bool {
	return ts.Index(t) != -1
}

// ForEach applies fn to each member of ts.
func (ts SliceFromTo[T, R]) ForEach(fn func(T)) {
	for _, t := range ts {
		fn(t)
	}
}

func (ts SliceFromTo[T, R]) Index(t T) int {
	for i := range ts {
		if t == ts[i] {
			return i
		}
	}

	return -1
}

// KeepIf returns the slice of elements from ts for which fn returns true.  It's the same as Filter would be.
func (ts SliceFromTo[T, R]) KeepIf(fn func(T) bool) SliceFromTo[T, R] {
	results := make([]T, 0, len(ts))

	for _, t := range ts {
		if fn(t) {
			results = append(results, t)
		}
	}

	return results
}

func (ts SliceFromTo[T, R]) Map(fn func(T) T) SliceFromTo[T, R] {
	results := make([]T, len(ts))

	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

func (ts SliceFromTo[T, R]) MapTo(fn func(T) R) RawSliceOf[R] {
	results := make([]R, len(ts))

	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// MapToBool returns the slice resulting from applying fn, whose return type is bool, to each member of ts.
func (ts SliceFromTo[T, R]) MapToBool(fn func(T) bool) SliceFromTo[bool, R] {
	results := make([]bool, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToInt returns the slice resulting from applying fn, whose return type is int, to each member of ts.
func (ts SliceFromTo[T, R]) MapToInt(fn func(T) int) SliceFromTo[int, R] {
	results := make([]int, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToString returns the slice resulting from applying fn, whose return type is string, to each member of ts.
func (ts SliceFromTo[T, R]) MapToString(fn func(T) string) SliceFromTo[string, R] {
	results := make([]string, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToStringOption returns the slice resulting from applying fn, whose return type is option.String, to each member of ts.
func (ts SliceFromTo[T, R]) MapToStringOption(fn func(T) option.String) SliceFromTo[option.String, R] {
	results := make([]option.String, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToStringSlice returns the slice resulting from applying fn, whose return type is []string, to each member of ts.
func (ts SliceFromTo[T, R]) MapToStringSlice(fn func(T) []string) RawSliceFromTo[[]string, R] {
	results := make([][]string, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// RemoveIf returns the slice of elements from ts for which fn returns false.
// It's the negation of what Filter would be, since it's not easy to write an in-line function for negation in Go.
func (ts SliceFromTo[T, R]) RemoveIf(fn func(T) bool) SliceFromTo[T, R] {
	results := make([]T, 0, len(ts))

	for _, t := range ts {
		if !fn(t) {
			results = append(results, t)
		}
	}

	return results
}
