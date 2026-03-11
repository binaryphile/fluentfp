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

import "github.com/binaryphile/fluentfp/option"

// node is a bush node (Stone §3.11): an element and zero or more children.
type node[T any] struct {
	elem     T
	children []*node[T]
}

// Heap is a persistent priority queue. Operations return new heaps; the
// original is unchanged. The comparator determines ordering: negative means
// the first argument has higher priority.
type Heap[T any] struct {
	root *node[T]
	cmp  func(T, T) int
	size int
}

// internal

// merge combines two nodes. The winner (smaller by cmp) becomes the root;
// the loser is prepended to the winner's children. Stone's merge-heaps.
func merge[T any](a, b *node[T], cmp func(T, T) int) *node[T] {
	if a == nil {
		return b
	}

	if b == nil {
		return a
	}

	if cmp(a.elem, b.elem) <= 0 {
		children := make([]*node[T], len(a.children)+1)
		children[0] = b
		copy(children[1:], a.children)

		return &node[T]{elem: a.elem, children: children}
	}

	children := make([]*node[T], len(b.children)+1)
	children[0] = a
	copy(children[1:], b.children)

	return &node[T]{elem: b.elem, children: children}
}

// mergePairs implements the pairing algorithm (Stone's heap-list-merger).
// Pass 1: pair adjacent children left-to-right.
// Pass 2: merge paired results right-to-left.
func mergePairs[T any](nodes []*node[T], cmp func(T, T) int) *node[T] {
	if len(nodes) == 0 {
		return nil
	}

	if len(nodes) == 1 {
		return nodes[0]
	}

	// Pass 1: left-to-right pairing
	paired := make([]*node[T], 0, (len(nodes)+1)/2)
	for i := 0; i+1 < len(nodes); i += 2 {
		paired = append(paired, merge(nodes[i], nodes[i+1], cmp))
	}

	if len(nodes)%2 == 1 {
		paired = append(paired, nodes[len(nodes)-1])
	}

	// Pass 2: right-to-left merge
	result := paired[len(paired)-1]
	for i := len(paired) - 2; i >= 0; i-- {
		result = merge(paired[i], result, cmp)
	}

	return result
}

// requireCmp panics if the heap has no comparator.
func (h Heap[T]) requireCmp() {
	if h.cmp == nil {
		panic("heap: no comparator (use heap.New or heap.From)")
	}
}

// constructors

// New returns an empty heap ordered by cmp.
// Panics if cmp is nil.
func New[T any](cmp func(T, T) int) Heap[T] {
	if cmp == nil {
		panic("heap.New: comparator must not be nil")
	}

	return Heap[T]{cmp: cmp}
}

// From builds a heap from a slice, ordered by cmp.
// Panics if cmp is nil.
func From[T any](ts []T, cmp func(T, T) int) Heap[T] {
	h := New[T](cmp)
	for _, t := range ts {
		h = h.Insert(t)
	}

	return h
}

// core operations

// Insert returns a new heap with t added. O(1).
// Stone's heap-adjoiner: create a singleton bush, merge with root.
func (h Heap[T]) Insert(t T) Heap[T] {
	h.requireCmp()

	singleton := &node[T]{elem: t}

	return Heap[T]{
		root: merge(h.root, singleton, h.cmp),
		cmp:  h.cmp,
		size: h.size + 1,
	}
}

// DeleteMin returns a new heap with the minimum element removed.
// Returns an empty heap (with comparator) if called on an empty heap.
// O(log n) amortized.
func (h Heap[T]) DeleteMin() Heap[T] {
	h.requireCmp()

	if h.root == nil {
		return Heap[T]{cmp: h.cmp}
	}

	return Heap[T]{
		root: mergePairs(h.root.children, h.cmp),
		cmp:  h.cmp,
		size: h.size - 1,
	}
}

// Merge returns a new heap combining h and other. O(1).
// Uses h's comparator. Both heaps must use the same ordering for correct results.
// Stone's merge-heaps.
func (h Heap[T]) Merge(other Heap[T]) Heap[T] {
	h.requireCmp()

	return Heap[T]{
		root: merge(h.root, other.root, h.cmp),
		cmp:  h.cmp,
		size: h.size + other.size,
	}
}

// queries

// Min returns the minimum element, or a not-ok option if empty. O(1).
func (h Heap[T]) Min() option.Option[T] {
	if h.root == nil {
		return option.Option[T]{}
	}

	return option.Of(h.root.elem)
}

// Pop returns the minimum element, the remaining heap, and true.
// Returns zero T, empty heap, and false if the heap is empty.
// Stone's heap-extractor.
func (h Heap[T]) Pop() (_ T, _ Heap[T], _ bool) {
	if h.root == nil {
		return
	}

	return h.root.elem, h.DeleteMin(), true
}

// IsEmpty reports whether the heap has no elements.
func (h Heap[T]) IsEmpty() bool {
	return h.root == nil
}

// Len returns the number of elements.
func (h Heap[T]) Len() int {
	return h.size
}

// consumption

// Collect returns all elements in sorted order. O(n log n).
// Returns nil for an empty heap. Stone's fold-heap applied with list + prepend.
func (h Heap[T]) Collect() []T {
	if h.root == nil {
		return nil
	}

	result := make([]T, 0, h.size)
	for !h.IsEmpty() {
		result = append(result, h.root.elem)
		h = h.DeleteMin()
	}

	return result
}
