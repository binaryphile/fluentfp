package base

import "slices"

// Sort returns a sorted copy using cmp (negative = a < b, zero = equal, positive = a > b).
// Build comparators from key extractors using Asc or Desc.
func (ts Mapper[T]) Sort(cmp func(T, T) int) Mapper[T] {
	c := make([]T, len(ts))
	copy(c, ts)
	slices.SortFunc(c, cmp)
	return c
}
