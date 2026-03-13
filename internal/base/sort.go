package base

import "slices"

// Sort returns a sorted copy using an unstable sort; elements that compare
// equal may be reordered. cmp must define a strict weak ordering: negative
// means a < b, zero means equal, positive means a > b. It must be consistent
// and transitive.
// Build comparators from key extractors using Asc or Desc.
func (ts Mapper[T]) Sort(cmp func(T, T) int) Mapper[T] {
	c := make([]T, len(ts))
	copy(c, ts)
	slices.SortFunc(c, cmp)
	return c
}

// IsSorted reports whether ts is sorted according to cmp.
func (ts Mapper[T]) IsSorted(cmp func(T, T) int) bool {
	return slices.IsSortedFunc(ts, cmp)
}
