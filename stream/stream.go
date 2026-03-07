package stream

import (
	"sync"

	"github.com/binaryphile/fluentfp/option"
)

// Cell states for the tail evaluation state machine.
const (
	cellPending    uint8 = iota // not yet evaluated, or panicked and retryable
	cellEvaluating              // a goroutine is computing the tail
	cellForced                  // successfully evaluated and memoized
)

// Stream is a lazy, memoized, persistent sequence. The zero value is an empty stream.
//
// Head-eager, tail-lazy: when a cell exists, its head is known. Only the tail is
// deferred and evaluated at most once. Operations like KeepIf eagerly scan to the
// first match; Map/Convert/TakeWhile eagerly transform the current head. Tail
// computation is always deferred into a thunk.
//
// Concurrent tail forcing is synchronized and memoized. Multiple goroutines can
// safely traverse the same stream; each cell's tail thunk executes at most once
// (on success). Thunk execution occurs outside internal locks — other goroutines
// block on internal state, not on user callback execution.
//
// Streams are persistent and memoized: multiple references to the same stream
// share forced cells. Holding a reference to an early cell pins all forced suffix
// cells reachable from it. From([]T) closures capture subslice views, which can
// pin the original backing array.
//
// Retry-on-panic: if a tail thunk panics, the cell stays unevaluated and future
// Tail() calls re-invoke the thunk. Head computation (at construction) is eager
// and not retryable. Callback purity is assumed for deterministic retry behavior.
//
// Reentrancy constraint: callbacks must not force the same cell being evaluated.
// This includes indirect paths (e.g., a Map callback that forces the Map result
// stream). This constraint is inherent to memoized lazy evaluation.
//
// Stream is a value type externally (like Option, Either, Result). Internal pointer
// enables shared memoization across multiple references.
type Stream[T any] struct {
	cell *cell[T]
}

// cell is the internal cons-cell. Head is always known; tail is lazy and memoized.
//
// Valid states:
//   - Terminal eager cell: state=pending, tail == nil, next == nil
//   - Pre-forced eager cell: state=forced, next pre-set (e.g., Repeat)
//   - Unevaluated lazy cell: state=pending, tail != nil
//   - Evaluating lazy cell: state=evaluating, wait != nil
//   - Evaluated lazy cell: state=forced, tail == nil, next is set
//   - Panicked lazy cell: state=pending, tail != nil (retryable)
type cell[T any] struct {
	head  T
	mu    sync.Mutex
	tail  func() *cell[T] // thunk; nil after successful evaluation or for eager cells
	next  *cell[T]         // memoized tail result
	state uint8            // cellPending, cellEvaluating, cellForced
	wait  chan struct{}     // closed when evaluation completes; nil when not evaluating
}

// getTail forces and memoizes the tail using a state machine that evaluates
// the thunk outside the internal mutex.
//
// Flow: pending → evaluating → forced (success) or → pending (panic, retryable).
// Waiters block on a channel, not on the mutex during thunk execution.
// If the thunk panics, the cell resets to pending and the panic is re-raised.
// Re-raising preserves the panic value but not the original stack trace.
func (c *cell[T]) getTail() *cell[T] {
	c.mu.Lock()

	for {
		switch c.state {
		case cellForced:
			c.mu.Unlock()
			return c.next

		case cellEvaluating:
			ch := c.wait
			c.mu.Unlock()
			<-ch // wait for evaluator to finish
			c.mu.Lock()
			continue // re-check state

		case cellPending:
			if c.tail == nil {
				// Eager/terminal cell — no thunk to evaluate.
				c.mu.Unlock()
				return c.next
			}

			// Become the evaluator.
			c.state = cellEvaluating
			c.wait = make(chan struct{})
			thunk := c.tail
			ch := c.wait
			c.mu.Unlock()

			// Execute thunk outside the lock.
			result, panicVal := evalThunk(thunk)

			c.mu.Lock()
			if panicVal != nil {
				c.state = cellPending
				close(ch) // wake waiters so one can retry
				c.mu.Unlock()
				panic(panicVal)
			}

			c.next = result
			c.tail = nil // release closure for GC
			c.state = cellForced
			close(ch) // wake waiters
			c.mu.Unlock()
			return result
		}
	}
}

// evalThunk runs a tail thunk, catching any panic so the caller can reset
// cell state before re-raising. The panic value is preserved; the original
// stack trace is not (inherent cost of catch-and-rethrow).
func evalThunk[T any](thunk func() *cell[T]) (result *cell[T], panicVal any) {
	defer func() {
		if r := recover(); r != nil {
			panicVal = r
		}
	}()

	return thunk(), nil
}

// IsEmpty returns true if the stream has no elements.
func (s Stream[T]) IsEmpty() bool {
	return s.cell == nil
}

// First returns the head element, or a not-ok option if empty.
func (s Stream[T]) First() option.Option[T] {
	if s.cell == nil {
		return option.Option[T]{}
	}

	return option.Of(s.cell.head)
}

// Tail returns the rest of the stream after the first element.
// Forces one cell's tail thunk (mutex-guarded, memoized). Returns empty stream if empty.
func (s Stream[T]) Tail() Stream[T] {
	if s.cell == nil {
		return Stream[T]{}
	}

	return Stream[T]{cell: s.cell.getTail()}
}
