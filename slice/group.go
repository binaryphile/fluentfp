package slice

// Group holds a grouping key and its collected items.
type Group[K comparable, T any] struct {
	Key   K
	Items []T
}
