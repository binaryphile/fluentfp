// Package kv provides fluent operations on Go maps (key-value collections).
package kv

import "github.com/binaryphile/fluentfp/slice"

// Entries wraps a map for key/value extraction.
type Entries[K comparable, V any] struct {
	m map[K]V
}

// From wraps a map for fluent operations.
func From[K comparable, V any](m map[K]V) Entries[K, V] {
	return Entries[K, V]{m: m}
}

// ToValues extracts the values as a Mapper for further transformation.
// Order is not guaranteed (map iteration order).
func (e Entries[K, V]) ToValues() slice.Mapper[V] {
	result := make([]V, 0, len(e.m))
	for _, v := range e.m {
		result = append(result, v)
	}

	return result
}

// ToKeys extracts the keys as a Mapper for further transformation.
// Order is not guaranteed (map iteration order).
func (e Entries[K, V]) ToKeys() slice.Mapper[K] {
	result := make([]K, 0, len(e.m))
	for k := range e.m {
		result = append(result, k)
	}

	return result
}

// Values extracts the values of m as a Mapper for further transformation.
// Shortcut for From(m).ToValues().
func Values[K comparable, V any](m map[K]V) slice.Mapper[V] {
	return From(m).ToValues()
}

// Keys extracts the keys of m as a Mapper for further transformation.
// Shortcut for From(m).ToKeys().
func Keys[K comparable, V any](m map[K]V) slice.Mapper[K] {
	return From(m).ToKeys()
}

// MapperTo wraps a map for cross-type transformation.
// T is first so K and V are inferred from the map argument.
type MapperTo[T any, K comparable, V any] struct {
	m map[K]V
}

// MapTo wraps a map for transformation to type T.
// Usage: kv.MapTo[TargetType](m).Map(fn)
func MapTo[T any, K comparable, V any](m map[K]V) MapperTo[T, K, V] {
	return MapperTo[T, K, V]{m: m}
}

// Map transforms each key-value pair using fn and returns the results
// as a Mapper for further transformation.
// Order is not guaranteed (map iteration order).
func (mt MapperTo[T, K, V]) Map(fn func(K, V) T) slice.Mapper[T] {
	result := make([]T, 0, len(mt.m))
	for k, v := range mt.m {
		result = append(result, fn(k, v))
	}

	return result
}
