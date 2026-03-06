// Package base defines the core collection types and their methods.
// These types are re-exported via type aliases in the slice and kv packages.
package base

import (
	"github.com/binaryphile/fluentfp/either"
	"github.com/binaryphile/fluentfp/option"
)

// Mapper is a fluent slice usable anywhere a regular slice is, but provides additional fluent fp methods.
// Its underlying type is []T.
type Mapper[T any] []T

// FindAs returns the first element that type-asserts to R, or not-ok if none match.
// Useful for finding a specific concrete type in a slice of interfaces.
func FindAs[R, T any](ts []T) option.Option[R] {
	for _, t := range ts {
		if r, ok := any(t).(R); ok {
			return option.Of(r)
		}
	}
	return option.NotOk[R]()
}

// Convert returns the result of applying fn to each member of ts.
func (ts Mapper[T]) Convert(fn func(T) T) Mapper[T] {
	results := make([]T, len(ts))
	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// Each applies fn to each member of ts.
func (ts Mapper[T]) Each(fn func(T)) {
	for _, t := range ts {
		fn(t)
	}
}

// First returns the first element, or not-ok if the slice is empty.
func (ts Mapper[T]) First() option.Option[T] {
	if len(ts) == 0 {
		return option.NotOk[T]()
	}
	return option.Of(ts[0])
}

// Any returns true if fn returns true for any element.
func (ts Mapper[T]) Any(fn func(T) bool) bool {
	for _, t := range ts {
		if fn(t) {
			return true
		}
	}
	return false
}

// Every returns true if fn returns true for every element.
// Returns true for an empty slice (vacuous truth).
func (ts Mapper[T]) Every(fn func(T) bool) bool {
	for _, t := range ts {
		if !fn(t) {
			return false
		}
	}
	return true
}

// None returns true if fn returns false for every element.
// Returns true for an empty slice (no elements match).
func (ts Mapper[T]) None(fn func(T) bool) bool {
	return !ts.Any(fn)
}

// Clone returns a shallow copy of the slice with independent backing array.
func (ts Mapper[T]) Clone() Mapper[T] {
	if ts == nil {
		return nil
	}
	c := make([]T, len(ts))
	copy(c, ts)
	return c
}

// Single returns Right(element) if exactly one element exists,
// or Left(count) if zero or more than one.
func (ts Mapper[T]) Single() either.Either[int, T] {
	if len(ts) == 1 {
		return either.Right[int, T](ts[0])
	}
	return either.Left[int, T](len(ts))
}

// Find returns the first element matching the predicate, or not-ok if none match.
func (ts Mapper[T]) Find(fn func(T) bool) option.Option[T] {
	for _, t := range ts {
		if fn(t) {
			return option.Of(t)
		}
	}
	return option.NotOk[T]()
}

// FlatMap applies fn to each element, concatenating the resulting slices in iteration order.
// Nil slices returned by fn are treated as empty. The result is always non-nil.
func (ts Mapper[T]) FlatMap(fn func(T) []T) Mapper[T] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		results = append(results, fn(t)...)
	}
	return results
}

// IndexWhere returns the index of the first element matching the predicate, or not-ok if none match.
func (ts Mapper[T]) IndexWhere(fn func(T) bool) option.Option[int] {
	for i, t := range ts {
		if fn(t) {
			return option.Of(i)
		}
	}
	return option.NotOk[int]()
}

// KeepIf returns a new slice containing the members of ts for which fn returns true.
// It is the complement of RemoveIf.
func (ts Mapper[T]) KeepIf(fn func(T) bool) Mapper[T] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if fn(t) {
			results = append(results, t)
		}
	}

	return results
}

// Len returns the length of the slice.
func (ts Mapper[T]) Len() int {
	return len(ts)
}

// Reverse returns a new slice with elements in reverse order.
func (ts Mapper[T]) Reverse() Mapper[T] {
	results := make([]T, len(ts))
	for i, t := range ts {
		results[len(ts)-1-i] = t
	}

	return results
}

// RemoveIf returns a new slice containing members for which fn returns false.
// It is the complement of KeepIf.
func (ts Mapper[T]) RemoveIf(fn func(T) bool) Mapper[T] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if !fn(t) {
			results = append(results, t)
		}
	}

	return results
}

// Take returns the first n elements of ts.
func (ts Mapper[T]) Take(n int) Mapper[T] {
	n = max(0, n)
	if n > len(ts) {
		n = len(ts)
	}

	return ts[:n]
}

// TakeLast returns the last n elements of ts.
func (ts Mapper[T]) TakeLast(n int) Mapper[T] {
	n = max(0, n)

	return ts[max(0, len(ts)-n):]
}

// ToAny returns the result of applying fn to each member of ts.
func (ts Mapper[T]) ToAny(fn func(T) any) Mapper[any] {
	results := make([]any, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToBool returns the result of applying fn to each member of ts.
func (ts Mapper[T]) ToBool(fn func(T) bool) Mapper[bool] {
	results := make([]bool, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToByte returns the result of applying fn to each member of ts.
func (ts Mapper[T]) ToByte(fn func(T) byte) Mapper[byte] {
	results := make([]byte, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToError returns the result of applying fn to each member of ts.
func (ts Mapper[T]) ToError(fn func(T) error) Mapper[error] {
	results := make([]error, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToFloat32 returns the result of applying fn to each member of ts.
func (ts Mapper[T]) ToFloat32(fn func(T) float32) Mapper[float32] {
	results := make([]float32, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToFloat64 returns the result of applying fn to each member of ts.
func (ts Mapper[T]) ToFloat64(fn func(T) float64) Float64 {
	results := make([]float64, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToInt returns the result of applying fn to each member of ts.
func (ts Mapper[T]) ToInt(fn func(T) int) Int {
	results := make([]int, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToInt32 returns the result of applying fn to each member of ts.
func (ts Mapper[T]) ToInt32(fn func(T) int32) Mapper[int32] {
	results := make([]int32, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToInt64 returns the result of applying fn to each member of ts.
func (ts Mapper[T]) ToInt64(fn func(T) int64) Mapper[int64] {
	results := make([]int64, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToRune returns the result of applying fn to each member of ts.
func (ts Mapper[T]) ToRune(fn func(T) rune) Mapper[rune] {
	results := make([]rune, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToString returns the result of applying fn to each member of ts.
func (ts Mapper[T]) ToString(fn func(T) string) String {
	results := make([]string, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}
