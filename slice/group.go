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

// CollectGroups converts a slice of groups (typically from [GroupBy]) into a map.
// Values for each key are the group's Items. Duplicate keys are merged
// via append, not overwritten.
// Returns an empty writable map for nil or empty input.
func CollectGroups[K comparable, T any](groups []Group[K, T]) map[K][]T {
	result := make(map[K][]T, len(groups))

	for _, g := range groups {
		result[g.Key] = append(result[g.Key], g.Items...)
	}

	return result
}
