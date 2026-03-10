package slice

// Group holds a grouping key and its collected items.
type Group[K comparable, T any] struct {
	Key   K
	Items []T
}

// GetKey returns the group's key.
func (g Group[K, T]) GetKey() K { return g.Key }

// GetItems returns the group's items.
func (g Group[K, T]) GetItems() []T { return g.Items }

// Len returns the number of items in the group.
func (g Group[K, T]) Len() int { return len(g.Items) }
