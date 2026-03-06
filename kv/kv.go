// Package kv provides fluent operations on Go maps (key-value collections).
package kv

import "github.com/binaryphile/fluentfp/internal/base"

// Entries is a defined type over map[K]V.
// Indexing, ranging, and len all work as with a plain map.
// The zero value is a nil map — safe for reads (len, range) but panics on write.
// From does not copy; the Entries and the original map share the same backing data.
type Entries[K comparable, V any] = base.Entries[K, V]

// MapperTo wraps a map for cross-type transformation.
type MapperTo[T any, K comparable, V any] = base.EntryMapper[T, K, V]

// From converts a map to Entries for fluent operations.
func From[K comparable, V any](m map[K]V) Entries[K, V] {
	return Entries[K, V](m)
}

// Values extracts the values of m as a Mapper for further transformation.
// Shortcut for From(m).Values().
func Values[K comparable, V any](m map[K]V) base.Mapper[V] {
	return From(m).Values()
}

// Keys extracts the keys of m as a Mapper for further transformation.
// Shortcut for From(m).Keys().
func Keys[K comparable, V any](m map[K]V) base.Mapper[K] {
	return From(m).Keys()
}

// Map transforms each key-value pair in m using fn and returns the results
// as a Mapper. All type parameters are inferred from the arguments.
// Use MapTo[T](m).Map(fn) when explicit type specification is needed.
// Order is not guaranteed (map iteration order).
func Map[K comparable, V, T any](m map[K]V, fn func(K, V) T) base.Mapper[T] {
	result := make([]T, 0, len(m))
	for k, v := range m {
		result = append(result, fn(k, v))
	}

	return result
}

// MapTo wraps a map for transformation to type T.
// Usage: kv.MapTo[TargetType](m).Map(fn)
func MapTo[T any, K comparable, V any](m map[K]V) MapperTo[T, K, V] {
	return base.NewEntryMapper[T](m)
}
