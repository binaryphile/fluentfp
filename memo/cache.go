package memo

import (
	"container/list"
	"sync"
)

// Cache is a thread-safe key-value store for memoized results.
// Implementations must handle their own synchronization.
type Cache[K comparable, V any] interface {
	Load(key K) (V, bool)
	Store(key K, value V)
}

// mapCache is an unbounded concurrent-safe cache backed by sync.RWMutex + map.
type mapCache[K comparable, V any] struct {
	mu sync.RWMutex
	m  map[K]V
}

// NewMap returns an unbounded, concurrent-safe cache.
func NewMap[K comparable, V any]() Cache[K, V] {
	return &mapCache[K, V]{m: make(map[K]V)}
}

func (c *mapCache[K, V]) Load(key K) (V, bool) {
	c.mu.RLock()
	v, ok := c.m[key]
	c.mu.RUnlock()
	return v, ok
}

func (c *mapCache[K, V]) Store(key K, value V) {
	c.mu.Lock()
	c.m[key] = value
	c.mu.Unlock()
}

// kv holds a key-value pair for LRU list elements.
type kv[K comparable, V any] struct {
	key   K
	value V
}

// lruCache is a concurrent-safe LRU cache. Load mutates access order
// (move-to-front), so all operations take a full mutex.
type lruCache[K comparable, V any] struct {
	mu       sync.Mutex
	m        map[K]*list.Element
	order    *list.List
	capacity int
}

// NewLRU returns a concurrent-safe LRU cache that evicts the least recently
// used entry when capacity is exceeded. Panics if capacity <= 0.
func NewLRU[K comparable, V any](capacity int) Cache[K, V] {
	if capacity <= 0 {
		panic("memo.NewLRU: capacity must be positive")
	}

	return &lruCache[K, V]{
		m:        make(map[K]*list.Element),
		order:    list.New(),
		capacity: capacity,
	}
}

func (c *lruCache[K, V]) Load(key K) (V, bool) {
	c.mu.Lock()
	el, ok := c.m[key]
	if !ok {
		c.mu.Unlock()
		var zero V
		return zero, false
	}

	c.order.MoveToFront(el)
	v := el.Value.(kv[K, V]).value
	c.mu.Unlock()
	return v, true
}

func (c *lruCache[K, V]) Store(key K, value V) {
	c.mu.Lock()

	if el, ok := c.m[key]; ok {
		el.Value = kv[K, V]{key: key, value: value}
		c.order.MoveToFront(el)
		c.mu.Unlock()
		return
	}

	if c.order.Len() >= c.capacity {
		back := c.order.Back()
		c.order.Remove(back)
		delete(c.m, back.Value.(kv[K, V]).key)
	}

	el := c.order.PushFront(kv[K, V]{key: key, value: value})
	c.m[key] = el
	c.mu.Unlock()
}
