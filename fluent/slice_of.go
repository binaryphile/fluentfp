package fluent

import (
	"github.com/binaryphile/fluentfp/option"
)

type SliceOf[T comparable] []T

func (ts SliceOf[T]) Contains(t T) bool {
	return ts.IndexOf(t) != -1
}

// Each applies fn to each member of ts.
func (ts SliceOf[T]) Each(fn func(T)) {
	for _, t := range ts {
		fn(t)
	}
}

func (ts SliceOf[T]) IndexOf(t T) int {
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

func (ts SliceOf[T]) Convert(fn func(T) T) SliceOf[T] {
	results := make([]T, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

func (ts SliceOf[T]) MapToAny(fn func(T) any) SliceOf[any] {
	results := make([]any, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToError returns the slice resulting from applying fn, whose return type is error, to each member of ts.
func (ts SliceOf[T]) MapToError(fn func(T) error) SliceOf[error] {
	results := make([]error, len(ts))

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

// ConvertToString returns the slice resulting from applying fn, whose return type is string, to each member of ts.
func (ts SliceOf[T]) ConvertToString(fn func(T) string) SliceOfStrings {
	results := make([]string, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToStringOption returns the slice resulting from applying fn, whose return type is option.String, to each member of ts.
func (ts SliceOf[T]) MapToStringOption(fn func(T) option.String) SliceOfStringOptions {
	results := make([]option.String, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// MapToSliceOfStrings returns the slice resulting from applying fn, whose return type is []string, to each member of ts.
func (ts SliceOf[T]) MapToSliceOfStrings(fn func(T) []string) SliceOfStringSlices {
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

// Take returns the first n elements of ts.
func (ts SliceOf[T]) Take(n int) SliceOf[T] {
	if n > len(ts) {
		n = len(ts)
	}

	return ts[:n]
}
