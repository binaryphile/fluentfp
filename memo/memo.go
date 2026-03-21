package memo

import "sync"

// Cell states for the memoization state machine.
const (
	pending    uint8 = iota // not yet evaluated, or panicked and retryable
	evaluating              // a goroutine is computing the result
	forced                  // successfully evaluated and memoized
)

// ofCell is the state machine for zero-arg memoization.
// Mirrors stream's cell.getTail() pattern.
type ofCell[T any] struct {
	mu     sync.Mutex
	fn     func() T
	result T
	state  uint8
	wait   chan struct{}
}

// From wraps a zero-arg function so it executes at most once on success.
// The result is cached and returned on subsequent calls. Thread-safe.
//
// If fn panics, the cell resets to pending — subsequent calls retry the function.
// This differs from sync.Once, which poisons permanently on panic.
//
// Reentrancy constraint: fn must not call the returned memoized function,
// directly or indirectly — this would deadlock on the internal mutex.
//
// Panics if fn is nil.
func From[T any](fn func() T) func() T {
	if fn == nil {
		panic("memo.From: fn must not be nil")
	}

	c := &ofCell[T]{fn: fn}
	return c.force
}

func (c *ofCell[T]) force() T {
	c.mu.Lock()

	for {
		switch c.state {
		case forced:
			c.mu.Unlock()
			return c.result

		case evaluating:
			ch := c.wait
			c.mu.Unlock()
			<-ch // wait for evaluator to finish
			c.mu.Lock()
			continue // re-check state

		case pending:
			c.state = evaluating
			c.wait = make(chan struct{})
			fn := c.fn
			ch := c.wait
			c.mu.Unlock()

			result, panicVal := safeCall(fn)

			c.mu.Lock()
			if panicVal != nil {
				c.state = pending
				close(ch) // wake waiters so one can retry
				c.mu.Unlock()
				panic(panicVal)
			}

			c.result = result
			c.fn = nil // release closure for GC
			c.state = forced
			close(ch) // wake waiters
			c.mu.Unlock()
			return result
		}
	}
}

// safeCall runs fn, catching any panic so the caller can reset state
// before re-raising.
func safeCall[T any](fn func() T) (result T, panicVal any) {
	defer func() {
		if r := recover(); r != nil {
			panicVal = r
		}
	}()

	return fn(), nil
}

// Fn wraps a single-arg function with an unbounded map cache.
// Cache-backed, not single-flight: concurrent misses for the same key
// may compute the value multiple times; the last store wins.
// Thread-safe. Panics if fn is nil.
func Fn[K comparable, V any](fn func(K) V) func(K) V {
	return FnWith(fn, NewMap[K, V]())
}

// FnWith wraps a single-arg function with a caller-provided cache.
// Cache-backed, not single-flight: concurrent misses for the same key
// may compute the value multiple times; the last store wins.
// Thread-safe (cache must handle its own synchronization).
// Panics if fn or cache is nil.
func FnWith[K comparable, V any](fn func(K) V, cache Cache[K, V]) func(K) V {
	if fn == nil {
		panic("memo.FnWith: fn must not be nil")
	}
	if cache == nil {
		panic("memo.FnWith: cache must not be nil")
	}

	return func(key K) V {
		if v, ok := cache.Load(key); ok {
			return v
		}

		v := fn(key)
		cache.Store(key, v)
		return v
	}
}

// FnErr wraps a fallible single-arg function with an unbounded map cache.
// Only successful results are cached — errors trigger retry on subsequent calls.
// Cache-backed, not single-flight: concurrent misses for the same key
// may compute the value multiple times; the last store wins.
// Thread-safe. Panics if fn is nil.
func FnErr[K comparable, V any](fn func(K) (V, error)) func(K) (V, error) {
	return FnErrWith(fn, NewMap[K, V]())
}

// FnErrWith wraps a fallible single-arg function with a caller-provided cache.
// Only successful results are cached — errors trigger retry on subsequent calls.
// Cache-backed, not single-flight: concurrent misses for the same key
// may compute the value multiple times; the last store wins.
// Thread-safe (cache must handle its own synchronization).
// Panics if fn or cache is nil.
func FnErrWith[K comparable, V any](fn func(K) (V, error), cache Cache[K, V]) func(K) (V, error) {
	if fn == nil {
		panic("memo.FnErrWith: fn must not be nil")
	}
	if cache == nil {
		panic("memo.FnErrWith: cache must not be nil")
	}

	return func(key K) (V, error) {
		if v, ok := cache.Load(key); ok {
			return v, nil
		}

		v, err := fn(key)
		if err != nil {
			return v, err
		}

		cache.Store(key, v)
		return v, nil
	}
}
