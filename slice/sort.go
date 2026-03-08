package slice

import (
	"cmp"
	"slices"
)

// SortBy returns a sorted copy of ts, ordered ascending by the key extracted via fn.
func SortBy[T any, K cmp.Ordered](ts Mapper[T], fn func(T) K) Mapper[T] {
	c := make([]T, len(ts))
	copy(c, ts)
	slices.SortFunc(c, func(a, b T) int { return cmp.Compare(fn(a), fn(b)) })
	return c
}

// SortByDesc returns a sorted copy of ts, ordered descending by the key extracted via fn.
func SortByDesc[T any, K cmp.Ordered](ts Mapper[T], fn func(T) K) Mapper[T] {
	c := make([]T, len(ts))
	copy(c, ts)
	slices.SortFunc(c, func(a, b T) int { return cmp.Compare(fn(b), fn(a)) })
	return c
}

// Asc builds an ascending comparator from a key extractor.
func Asc[T any, S cmp.Ordered](key func(T) S) func(T, T) int {
	return func(a, b T) int { return cmp.Compare(key(a), key(b)) }
}

// Desc builds a descending comparator from a key extractor.
func Desc[T any, S cmp.Ordered](key func(T) S) func(T, T) int {
	return func(a, b T) int { return cmp.Compare(key(b), key(a)) }
}
