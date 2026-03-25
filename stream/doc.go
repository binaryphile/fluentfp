// Package stream provides lazy, memoized, persistent sequences.
//
// A Stream is a linked list where each cell's head is eager and tail is lazy,
// evaluated at most once. Streams are value types externally with shared
// memoization via internal pointers. If a tail thunk panics, the cell
// remains unevaluated and future accesses will retry (re-invoking the thunk).
//
// This package supports pure/in-memory sources only. Effectful sources
// (channels, iterators) are deferred to a future version.
//
// For non-memoized lazy sequences that re-evaluate on each iteration,
// see [seq.Seq].
package stream
