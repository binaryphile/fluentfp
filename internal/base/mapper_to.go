package base

import (
	"slices"

	"github.com/binaryphile/fluentfp/either"
	"github.com/binaryphile/fluentfp/option"
)

// MapperTo is a fluent slice for filter→map chains where the cross-type map comes last.
// Prefer slice.Map(ts, fn) for most cross-type mapping — it infers all types and returns
// Mapper[R] for further chaining. Use MapTo[R] only when you need to filter or transform
// before the cross-type map: slice.MapTo[R](ts).KeepIf(pred).Map(fn).
type MapperTo[R, T any] []T

// Clone returns a shallow copy of the slice with independent backing array.
func (ts MapperTo[R, T]) Clone() MapperTo[R, T] {
	if ts == nil {
		return nil
	}
	c := make([]T, len(ts))
	copy(c, ts)
	return c
}

// Convert returns the result of applying fn to each member of ts.
func (ts MapperTo[R, T]) Convert(fn func(T) T) MapperTo[R, T] {
	results := make([]T, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// First returns the first element, or not-ok if the slice is empty.
func (ts MapperTo[R, T]) First() option.Option[T] {
	if len(ts) == 0 {
		return option.NotOk[T]()
	}
	return option.Of(ts[0])
}

// Last returns the last element, or not-ok if the slice is empty.
func (ts MapperTo[R, T]) Last() option.Option[T] {
	if len(ts) == 0 {
		return option.NotOk[T]()
	}
	return option.Of(ts[len(ts)-1])
}

// Each applies fn to each member of ts.
func (ts MapperTo[R, T]) Each(fn func(T)) {
	for _, t := range ts {
		fn(t)
	}
}

// FlatMap applies fn to each element, concatenating the resulting slices in iteration order.
// Nil slices returned by fn are treated as empty. The result is always non-nil.
func (ts MapperTo[R, T]) FlatMap(fn func(T) []R) Mapper[R] {
	results := make([]R, 0, len(ts))
	for _, t := range ts {
		results = append(results, fn(t)...)
	}
	return results
}

// KeepIf returns a new slice containing the members of ts for which fn returns true.
// It is the complement of RemoveIf.
func (ts MapperTo[R, T]) KeepIf(fn func(T) bool) MapperTo[R, T] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if fn(t) {
			results = append(results, t)
		}
	}
	return results
}

// KeyByInt indexes elements by an int key derived from fn.
// If multiple elements produce the same key, the last one wins.
// For other key types, use the standalone KeyBy function.
func (ts MapperTo[R, T]) KeyByInt(fn func(T) int) map[int]T {
	return KeyBy(ts, fn)
}

// KeyByString indexes elements by a string key derived from fn.
// If multiple elements produce the same key, the last one wins.
// For other key types, use the standalone KeyBy function.
func (ts MapperTo[R, T]) KeyByString(fn func(T) string) map[string]T {
	return KeyBy(ts, fn)
}

// Len returns the length of the slice.
func (ts MapperTo[R, T]) Len() int {
	return len(ts)
}

// Single returns Right(element) if exactly one element exists,
// or Left(count) if zero or more than one.
func (ts MapperTo[R, T]) Single() either.Either[int, T] {
	if len(ts) == 1 {
		return either.Right[int, T](ts[0])
	}
	return either.Left[int, T](len(ts))
}

// Reverse returns a new slice with elements in reverse order.
func (ts MapperTo[R, T]) Reverse() MapperTo[R, T] {
	results := make([]T, len(ts))
	for i, t := range ts {
		results[len(ts)-1-i] = t
	}

	return results
}

// RemoveIf returns a new slice containing members for which fn returns false.
// It is the complement of KeepIf.
func (ts MapperTo[R, T]) RemoveIf(fn func(T) bool) MapperTo[R, T] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if !fn(t) {
			results = append(results, t)
		}
	}
	return results
}

// Partition splits ts into two slices: elements where fn returns true, and elements where it returns false.
// Single pass. Both results are independent slices.
// For use in standalone form, see the Partition function in the slice package.
func (ts MapperTo[R, T]) Partition(fn func(T) bool) (MapperTo[R, T], MapperTo[R, T]) {
	match, rest := Partition(ts, fn)
	return match, rest
}

// Take returns the first n members of ts.
func (ts MapperTo[R, T]) Take(n int) MapperTo[R, T] {
	n = max(0, n)
	if n > len(ts) {
		n = len(ts)
	}

	return ts[:n]
}

// TakeLast returns the last n members of ts.
func (ts MapperTo[R, T]) TakeLast(n int) MapperTo[R, T] {
	n = max(0, n)

	return ts[max(0, len(ts)-n):]
}

// ToAny returns the result of applying fn to each member of ts.
func (ts MapperTo[R, T]) ToAny(fn func(T) any) MapperTo[R, any] {
	results := make([]any, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToBool returns the result of applying fn to each member of ts.
func (ts MapperTo[R, T]) ToBool(fn func(T) bool) MapperTo[R, bool] {
	results := make([]bool, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToByte returns the result of applying fn to each member of ts.
func (ts MapperTo[R, T]) ToByte(fn func(T) byte) MapperTo[R, byte] {
	results := make([]byte, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToError returns the result of applying fn to each member of ts.
func (ts MapperTo[R, T]) ToError(fn func(T) error) MapperTo[R, error] {
	results := make([]error, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToFloat32 returns the result of applying fn to each member of ts.
func (ts MapperTo[R, T]) ToFloat32(fn func(T) float32) MapperTo[R, float32] {
	results := make([]float32, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToFloat64 returns the result of applying fn to each member of ts.
func (ts MapperTo[R, T]) ToFloat64(fn func(T) float64) MapperTo[R, float64] {
	results := make([]float64, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToInt returns the result of applying fn to each member of ts.
func (ts MapperTo[R, T]) ToInt(fn func(T) int) MapperTo[R, int] {
	results := make([]int, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToInt32 returns the result of applying fn to each member of ts.
func (ts MapperTo[R, T]) ToInt32(fn func(T) int32) MapperTo[R, int32] {
	results := make([]int32, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToInt64 returns the result of applying fn to each member of ts.
func (ts MapperTo[R, T]) ToInt64(fn func(T) int64) MapperTo[R, int64] {
	results := make([]int64, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// Sort returns a sorted copy using cmp (negative = a < b, zero = equal, positive = a > b).
// Build comparators from key extractors using Asc or Desc.
func (ts MapperTo[R, T]) Sort(cmp func(T, T) int) MapperTo[R, T] {
	c := make([]T, len(ts))
	copy(c, ts)
	slices.SortFunc(c, cmp)
	return c
}

// Map returns the result of applying fn to each member of ts.
func (ts MapperTo[R, T]) Map(fn func(T) R) Mapper[R] {
	results := make([]R, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToRune returns the result of applying fn to each member of ts.
func (ts MapperTo[R, T]) ToRune(fn func(T) rune) MapperTo[R, rune] {
	results := make([]rune, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToString returns the result of applying fn to each member of ts.
func (ts MapperTo[R, T]) ToString(fn func(T) string) MapperTo[R, string] {
	results := make([]string, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}
