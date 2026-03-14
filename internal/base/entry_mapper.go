package base

// EntryMapper wraps a map for cross-type transformation.
// T is first so K and V are inferred from the map argument.
type EntryMapper[T any, K comparable, V any] struct {
	m map[K]V
}

// NewEntryMapper creates an EntryMapper for transformation to type T.
// Accepts any map type with underlying type map[K]V, including defined types like Entries.
func NewEntryMapper[T any, K comparable, V any, M ~map[K]V](m M) EntryMapper[T, K, V] {
	return EntryMapper[T, K, V]{m: map[K]V(m)}
}

// Map transforms each key-value pair using fn and returns the results
// as a Mapper for further transformation.
// Order is not guaranteed (map iteration order).
func (mt EntryMapper[T, K, V]) Map(fn func(K, V) T) Mapper[T] {
	result := make([]T, 0, len(mt.m))
	for k, v := range mt.m {
		result = append(result, fn(k, v))
	}

	return result
}
