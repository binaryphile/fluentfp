// Package stream provides lazy, memoized, persistent sequences.
//
// A Stream is a linked list where each cell's head is eager and tail is lazy,
// evaluated at most once. Streams are value types externally with shared
// memoization via internal pointers. If a tail thunk panics, the cell
// remains unevaluated and future accesses will retry (re-invoking the thunk).
//
// This package supports pure/in-memory sources only. Effectful sources
// (channels, iterators) are deferred to a future version.
package stream

func _() {
	// Stream methods — accessors
	_ = Stream[int].IsEmpty
	_ = Stream[int].First
	_ = Stream[int].Tail

	// Stream methods — lazy operations
	_ = Stream[int].KeepIf
	_ = Stream[int].Convert
	_ = Stream[int].Take
	_ = Stream[int].TakeWhile
	_ = Stream[int].Drop
	_ = Stream[int].DropWhile

	// Stream methods — terminal operations
	_ = Stream[int].Each
	_ = Stream[int].Collect
	_ = Stream[int].Find
	_ = Stream[int].Any
	_ = Stream[int].Seq

	// Constructors
	_ = From[int]
	_ = Of[int]
	_ = Generate[int]
	_ = Repeat[int]
	_ = Unfold[int, int]
	_ = Paginate[int, int]
	_ = Prepend[int]
	_ = PrependLazy[int]

	// Standalone functions
	_ = Map[int, string]
	_ = Fold[int, string]
}
