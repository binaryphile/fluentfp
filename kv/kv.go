// Package kv provides fluent operations on Go maps (key-value collections).
package kv

import "github.com/binaryphile/fluentfp/slice"

// Entries is a defined type over map[K]V.
// Indexing, ranging, and len all work as with a plain map.
type Entries[K comparable, V any] map[K]V

// From converts a map to Entries for fluent operations.
func From[K comparable, V any](m map[K]V) Entries[K, V] {
	return Entries[K, V](m)
}

// ToValues extracts the values as a Mapper for further transformation.
// Order is not guaranteed (map iteration order).
func (e Entries[K, V]) ToValues() slice.Mapper[V] {
	result := make([]V, 0, len(e))
	for _, v := range e {
		result = append(result, v)
	}

	return result
}

// ToKeys extracts the keys as a Mapper for further transformation.
// Order is not guaranteed (map iteration order).
func (e Entries[K, V]) ToKeys() slice.Mapper[K] {
	result := make([]K, 0, len(e))
	for k := range e {
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

// ToAny returns the result of applying fn to each key-value pair.
func (e Entries[K, V]) ToAny(fn func(K, V) any) slice.Mapper[any] {
	result := make([]any, 0, len(e))
	for k, v := range e {
		result = append(result, fn(k, v))
	}

	return result
}

// ToBool returns the result of applying fn to each key-value pair.
func (e Entries[K, V]) ToBool(fn func(K, V) bool) slice.Mapper[bool] {
	result := make([]bool, 0, len(e))
	for k, v := range e {
		result = append(result, fn(k, v))
	}

	return result
}

// ToByte returns the result of applying fn to each key-value pair.
func (e Entries[K, V]) ToByte(fn func(K, V) byte) slice.Mapper[byte] {
	result := make([]byte, 0, len(e))
	for k, v := range e {
		result = append(result, fn(k, v))
	}

	return result
}

// ToError returns the result of applying fn to each key-value pair.
func (e Entries[K, V]) ToError(fn func(K, V) error) slice.Mapper[error] {
	result := make([]error, 0, len(e))
	for k, v := range e {
		result = append(result, fn(k, v))
	}

	return result
}

// ToFloat32 returns the result of applying fn to each key-value pair.
func (e Entries[K, V]) ToFloat32(fn func(K, V) float32) slice.Mapper[float32] {
	result := make([]float32, 0, len(e))
	for k, v := range e {
		result = append(result, fn(k, v))
	}

	return result
}

// ToFloat64 returns the result of applying fn to each key-value pair.
func (e Entries[K, V]) ToFloat64(fn func(K, V) float64) slice.Float64 {
	result := make([]float64, 0, len(e))
	for k, v := range e {
		result = append(result, fn(k, v))
	}

	return result
}

// ToInt returns the result of applying fn to each key-value pair.
func (e Entries[K, V]) ToInt(fn func(K, V) int) slice.Int {
	result := make([]int, 0, len(e))
	for k, v := range e {
		result = append(result, fn(k, v))
	}

	return result
}

// ToInt32 returns the result of applying fn to each key-value pair.
func (e Entries[K, V]) ToInt32(fn func(K, V) int32) slice.Mapper[int32] {
	result := make([]int32, 0, len(e))
	for k, v := range e {
		result = append(result, fn(k, v))
	}

	return result
}

// ToInt64 returns the result of applying fn to each key-value pair.
func (e Entries[K, V]) ToInt64(fn func(K, V) int64) slice.Mapper[int64] {
	result := make([]int64, 0, len(e))
	for k, v := range e {
		result = append(result, fn(k, v))
	}

	return result
}

// ToRune returns the result of applying fn to each key-value pair.
func (e Entries[K, V]) ToRune(fn func(K, V) rune) slice.Mapper[rune] {
	result := make([]rune, 0, len(e))
	for k, v := range e {
		result = append(result, fn(k, v))
	}

	return result
}

// ToString returns the result of applying fn to each key-value pair.
func (e Entries[K, V]) ToString(fn func(K, V) string) slice.String {
	result := make([]string, 0, len(e))
	for k, v := range e {
		result = append(result, fn(k, v))
	}

	return result
}

// Map transforms each key-value pair in m using fn and returns the results
// as a Mapper. All type parameters are inferred from the arguments.
// Use MapTo[T](m).Map(fn) when explicit type specification is needed.
// Order is not guaranteed (map iteration order).
func Map[K comparable, V, T any](m map[K]V, fn func(K, V) T) slice.Mapper[T] {
	result := make([]T, 0, len(m))
	for k, v := range m {
		result = append(result, fn(k, v))
	}

	return result
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
