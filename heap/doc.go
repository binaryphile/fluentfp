// Package heap provides a persistent (immutable) priority queue backed by a
// pairing heap, parameterized by a comparator function.
//
// Based on Stone's Algorithms for Functional Programming (Ch 4). The pairing
// merge strategy (Stone's heap-list-merger) gives O(1) insert and merge with
// O(log n) amortized delete-min.
//
// Create heaps with [New] or [From]. The zero value is a valid empty heap for
// queries (IsEmpty, Min, Len) but panics on Insert, DeleteMin, and Merge
// because no comparator is available.
//
// Use slice.Asc and slice.Desc from the slice package to build comparators:
//
//	h := heap.New(slice.Asc(Widget.Priority))   // min-heap by priority
//	h := heap.New(slice.Desc(Widget.Priority))  // max-heap by priority
package heap
