// Package base defines the core collection types and their methods.
// These types are re-exported via type aliases in the slice and kv packages.
package base

import (
	"github.com/binaryphile/fluentfp/either"
	"github.com/binaryphile/fluentfp/option"
)

// Mapper is a defined type over []T. It preserves indexing, range, and len
// while adding chainable functional methods. The zero value is a nil slice;
// methods treat it like an empty slice.
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

// Transform returns the result of applying fn to each member of ts.
// fn must not be nil.
func (ts Mapper[T]) Transform(fn func(T) T) Mapper[T] {
	results := make([]T, len(ts))
	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// Drop returns ts without the first n elements.
// Returns empty if n >= len(ts). Negative n is treated as zero.
// The result aliases the input backing array.
func (ts Mapper[T]) Drop(n int) Mapper[T] {
	if len(ts) == 0 {
		return Mapper[T]{}
	}
	n = min(max(0, n), len(ts))

	return ts[n:]
}

// DropLast returns ts without the last n elements.
// Returns empty if n >= len(ts). Negative n is treated as zero.
// The result aliases the input backing array.
func (ts Mapper[T]) DropLast(n int) Mapper[T] {
	if len(ts) == 0 {
		return Mapper[T]{}
	}
	return ts[:max(0, len(ts)-max(0, n))]
}

// DropWhile returns the suffix of ts remaining after dropping the longest prefix
// of elements that satisfy fn. The result aliases the input backing array.
// fn must not be nil.
func (ts Mapper[T]) DropWhile(fn func(T) bool) Mapper[T] {
	if len(ts) == 0 {
		return Mapper[T]{}
	}
	for i, t := range ts {
		if !fn(t) {
			return ts[i:]
		}
	}

	return ts[len(ts):]
}

// DropLastWhile returns the prefix of ts remaining after dropping the longest suffix
// of elements that satisfy fn. The result aliases the input backing array.
// fn must not be nil.
func (ts Mapper[T]) DropLastWhile(fn func(T) bool) Mapper[T] {
	if len(ts) == 0 {
		return Mapper[T]{}
	}
	for i := len(ts) - 1; i >= 0; i-- {
		if !fn(ts[i]) {
			return ts[:i+1]
		}
	}

	return ts[:0]
}

// Each applies fn to each member of ts in index order.
// fn must not be nil.
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

// Last returns the last element, or not-ok if the slice is empty.
func (ts Mapper[T]) Last() option.Option[T] {
	if len(ts) == 0 {
		return option.NotOk[T]()
	}
	return option.Of(ts[len(ts)-1])
}

// Any reports whether fn returns true for any element.
// Scans elements in index order and stops at the first match.
// fn must not be nil.
func (ts Mapper[T]) Any(fn func(T) bool) bool {
	for _, t := range ts {
		if fn(t) {
			return true
		}
	}
	return false
}

// Every reports whether fn returns true for every element.
// Scans elements in index order and stops at the first mismatch.
// Returns true for an empty slice (vacuous truth). fn must not be nil.
func (ts Mapper[T]) Every(fn func(T) bool) bool {
	for _, t := range ts {
		if !fn(t) {
			return false
		}
	}
	return true
}

// None reports whether fn returns false for every element.
// Scans elements in index order and stops at the first match.
// Returns true for an empty slice (no elements match). fn must not be nil.
func (ts Mapper[T]) None(fn func(T) bool) bool {
	return !ts.Any(fn)
}

// Clone returns a shallow copy of the slice with independent backing array.
func (ts Mapper[T]) Clone() Mapper[T] {
	if len(ts) == 0 {
		return Mapper[T]{}
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
// Scans elements in index order and stops at the first match.
// fn must not be nil.
func (ts Mapper[T]) Find(fn func(T) bool) option.Option[T] {
	for _, t := range ts {
		if fn(t) {
			return option.Of(t)
		}
	}
	return option.NotOk[T]()
}

// FindLast returns the last element matching the predicate, or not-ok if none match.
// Scans elements in reverse index order and stops at the first match.
// fn must not be nil.
func (ts Mapper[T]) FindLast(fn func(T) bool) option.Option[T] {
	for i := len(ts) - 1; i >= 0; i-- {
		if fn(ts[i]) {
			return option.Of(ts[i])
		}
	}

	return option.NotOk[T]()
}

// FlatMap applies fn to each element, concatenating the resulting slices in index order.
// Nil slices returned by fn are treated as empty. The result is always non-nil.
// fn must not be nil.
func (ts Mapper[T]) FlatMap(fn func(T) []T) Mapper[T] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		results = append(results, fn(t)...)
	}
	return results
}

// IndexWhere returns the index of the first element matching the predicate, or not-ok if none match.
// Scans elements in index order and stops at the first match.
// fn must not be nil.
func (ts Mapper[T]) IndexWhere(fn func(T) bool) option.Option[int] {
	for i, t := range ts {
		if fn(t) {
			return option.Of(i)
		}
	}
	return option.NotOk[int]()
}

// LastIndexWhere returns the index of the last element matching the predicate, or not-ok if none match.
// Scans elements in reverse index order and stops at the first match.
// fn must not be nil.
func (ts Mapper[T]) LastIndexWhere(fn func(T) bool) option.Option[int] {
	for i := len(ts) - 1; i >= 0; i-- {
		if fn(ts[i]) {
			return option.Of(i)
		}
	}

	return option.NotOk[int]()
}

// Intersperse inserts sep between every adjacent pair of elements.
// Returns a new slice; the result does not alias the input.
func (ts Mapper[T]) Intersperse(sep T) Mapper[T] {
	if len(ts) == 0 {
		return Mapper[T]{}
	}

	results := make([]T, 2*len(ts)-1)
	for i, t := range ts {
		results[2*i] = t
		if i < len(ts)-1 {
			results[2*i+1] = sep
		}
	}

	return results
}

// KeyByInt indexes elements by an int key derived from fn.
// If multiple elements produce the same key, the last one wins.
// For other key types, use the standalone KeyBy function.
func (ts Mapper[T]) KeyByInt(fn func(T) int) map[int]T {
	return KeyBy(ts, fn)
}

// KeyByString indexes elements by a string key derived from fn.
// If multiple elements produce the same key, the last one wins.
// For other key types, use the standalone KeyBy function.
func (ts Mapper[T]) KeyByString(fn func(T) string) map[string]T {
	return KeyBy(ts, fn)
}

// KeepIf returns a new slice containing the members of ts for which fn returns true.
// It is the complement of [Mapper.RemoveIf]. fn must not be nil.
func (ts Mapper[T]) KeepIf(fn func(T) bool) Mapper[T] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if fn(t) {
			results = append(results, t)
		}
	}

	return results
}

// KeepIfWhen conditionally filters: when cond is true, behaves like [Mapper.KeepIf];
// when cond is false, returns ts unchanged. This preserves method chaining
// when a filter should only apply based on an external condition (e.g., an optional query parameter).
// If cond is true, fn must not be nil.
func (ts Mapper[T]) KeepIfWhen(cond bool, fn func(T) bool) Mapper[T] {
	if !cond {
		return ts
	}
	return ts.KeepIf(fn)
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
// It is the complement of [Mapper.KeepIf]. fn must not be nil.
func (ts Mapper[T]) RemoveIf(fn func(T) bool) Mapper[T] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if !fn(t) {
			results = append(results, t)
		}
	}

	return results
}

// RemoveIfWhen conditionally filters: when cond is true, behaves like [Mapper.RemoveIf];
// when cond is false, returns ts unchanged. This preserves method chaining
// when a filter should only apply based on an external condition.
// If cond is true, fn must not be nil.
func (ts Mapper[T]) RemoveIfWhen(cond bool, fn func(T) bool) Mapper[T] {
	if !cond {
		return ts
	}
	return ts.RemoveIf(fn)
}

// Partition splits ts into two slices: elements where fn returns true, and elements where it returns false.
// Single pass. Both results are independent slices. fn must not be nil.
// For use in standalone form, see the Partition function in the slice package.
func (ts Mapper[T]) Partition(fn func(T) bool) (Mapper[T], Mapper[T]) {
	match, rest := Partition(ts, fn)
	return match, rest
}

// Take returns the first n elements of ts. Negative n is treated as zero;
// n greater than len(ts) is treated as len(ts). The result aliases the input backing array.
func (ts Mapper[T]) Take(n int) Mapper[T] {
	if len(ts) == 0 {
		return Mapper[T]{}
	}
	n = max(0, n)
	if n > len(ts) {
		n = len(ts)
	}

	return ts[:n]
}

// TakeLast returns the last n elements of ts. Negative n is treated as zero;
// n greater than len(ts) returns the entire slice. The result aliases the input backing array.
func (ts Mapper[T]) TakeLast(n int) Mapper[T] {
	if len(ts) == 0 {
		return Mapper[T]{}
	}
	n = max(0, n)

	return ts[max(0, len(ts)-n):]
}

// TakeWhile returns the longest prefix of elements that satisfy fn.
// The result aliases the input backing array. fn must not be nil.
func (ts Mapper[T]) TakeWhile(fn func(T) bool) Mapper[T] {
	if len(ts) == 0 {
		return Mapper[T]{}
	}
	for i, t := range ts {
		if !fn(t) {
			return ts[:i]
		}
	}

	return ts
}

// ToAny returns the result of applying fn to each member of ts. fn must not be nil.
func (ts Mapper[T]) ToAny(fn func(T) any) Mapper[any] {
	results := make([]any, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToBool returns the result of applying fn to each member of ts. fn must not be nil.
func (ts Mapper[T]) ToBool(fn func(T) bool) Mapper[bool] {
	results := make([]bool, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToByte returns the result of applying fn to each member of ts. fn must not be nil.
func (ts Mapper[T]) ToByte(fn func(T) byte) Mapper[byte] {
	results := make([]byte, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToError returns the result of applying fn to each member of ts. fn must not be nil.
func (ts Mapper[T]) ToError(fn func(T) error) Mapper[error] {
	results := make([]error, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToFloat32 returns the result of applying fn to each member of ts. fn must not be nil.
func (ts Mapper[T]) ToFloat32(fn func(T) float32) Mapper[float32] {
	results := make([]float32, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToFloat64 returns the result of applying fn to each member of ts. fn must not be nil.
func (ts Mapper[T]) ToFloat64(fn func(T) float64) Float64 {
	results := make([]float64, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToInt returns the result of applying fn to each member of ts. fn must not be nil.
func (ts Mapper[T]) ToInt(fn func(T) int) Int {
	results := make([]int, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToInt32 returns the result of applying fn to each member of ts. fn must not be nil.
func (ts Mapper[T]) ToInt32(fn func(T) int32) Mapper[int32] {
	results := make([]int32, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToInt64 returns the result of applying fn to each member of ts. fn must not be nil.
func (ts Mapper[T]) ToInt64(fn func(T) int64) Mapper[int64] {
	results := make([]int64, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToRune returns the result of applying fn to each member of ts. fn must not be nil.
func (ts Mapper[T]) ToRune(fn func(T) rune) Mapper[rune] {
	results := make([]rune, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToString returns the result of applying fn to each member of ts. fn must not be nil.
func (ts Mapper[T]) ToString(fn func(T) string) String {
	results := make([]string, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}
