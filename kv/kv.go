// Package kv provides fluent operations on Go maps (key-value collections).
package kv

import (
	"github.com/binaryphile/fluentfp/internal/base"
	"github.com/binaryphile/fluentfp/tuple/pair"
)

// Entries is a defined type over map[K]V.
// Indexing, ranging, and len all work as with a plain map.
// The zero value is a nil map — safe for reads (len, range) but panics on write.
// From does not copy; the Entries and the original map share the same backing data.
type Entries[K comparable, V any] = base.Entries[K, V]

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
// Order is not guaranteed (map iteration order).
// Panics if fn is nil.
func Map[K comparable, V, T any](m map[K]V, fn func(K, V) T) base.Mapper[T] {
	if fn == nil {
		panic("kv.Map: fn must not be nil")
	}

	result := make([]T, 0, len(m))
	for k, v := range m {
		result = append(result, fn(k, v))
	}

	return result
}

// MapValues transforms each value in m using fn, preserving keys.
// Returns Entries for chaining (e.g., MapValues(m, fn).KeepIf(pred).Values()).
// Panics if fn is nil.
func MapValues[K comparable, V, V2 any](m map[K]V, fn func(V) V2) base.Entries[K, V2] {
	if fn == nil {
		panic("kv.MapValues: fn must not be nil")
	}

	result := make(map[K]V2, len(m))
	for k, v := range m {
		result[k] = fn(v)
	}

	return result
}

// MapKeys transforms each key in m using fn, preserving values.
// Returns Entries for chaining (e.g., MapKeys(m, fn).KeepIf(pred).Values()).
// If fn maps multiple keys to the same K2, last-wins (map iteration order).
// Panics if fn is nil.
func MapKeys[K comparable, V any, K2 comparable](m map[K]V, fn func(K) K2) base.Entries[K2, V] {
	if fn == nil {
		panic("kv.MapKeys: fn must not be nil")
	}

	result := make(map[K2]V, len(m))
	for k, v := range m {
		result[fn(k)] = v
	}

	return result
}

// ToPairs converts a map to a slice of key-value pairs.
// Order is not guaranteed (map iteration order).
func ToPairs[K comparable, V any](m map[K]V) base.Mapper[pair.Pair[K, V]] {
	result := make([]pair.Pair[K, V], 0, len(m))
	for k, v := range m {
		result = append(result, pair.Of(k, v))
	}

	return result
}

// FromPairs converts a slice of key-value pairs to Entries.
// If duplicate keys exist, the last pair wins.
func FromPairs[K comparable, V any](pairs []pair.Pair[K, V]) Entries[K, V] {
	result := make(map[K]V, len(pairs))
	for _, p := range pairs {
		result[p.First] = p.Second
	}

	return result
}

// Invert swaps keys and values. If multiple keys map to the same value,
// an arbitrary one wins (map iteration order). Both K and V must be comparable.
func Invert[K, V comparable](m map[K]V) map[V]K {
	result := make(map[V]K, len(m))
	for k, v := range m {
		result[v] = k
	}
	return result
}

// Merge combines multiple maps. Later maps override earlier keys.
// Returns a new map; inputs are not modified.
func Merge[K comparable, V any](maps ...map[K]V) map[K]V {
	size := 0
	for _, m := range maps {
		size += len(m)
	}
	result := make(map[K]V, size)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// PickByKeys returns a new map containing only entries whose keys appear in keys.
// Keys not present in m are silently ignored.
func PickByKeys[K comparable, V any](m map[K]V, keys []K) map[K]V {
	result := make(map[K]V, len(keys))
	for _, k := range keys {
		if v, ok := m[k]; ok {
			result[k] = v
		}
	}
	return result
}

// OmitByKeys returns a new map containing entries whose keys do NOT appear in keys.
func OmitByKeys[K comparable, V any](m map[K]V, keys []K) map[K]V {
	exclude := make(map[K]bool, len(keys))
	for _, k := range keys {
		exclude[k] = true
	}
	result := make(map[K]V, len(m))
	for k, v := range m {
		if !exclude[k] {
			result[k] = v
		}
	}
	return result
}
