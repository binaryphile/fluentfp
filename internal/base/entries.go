package base

// Entries is a defined type over map[K]V.
// Indexing, ranging, and len all work as with a plain map.
// The zero value is a nil map — safe for reads (len, range) but panics on write.
// From does not copy; the Entries and the original map share the same backing data.
type Entries[K comparable, V any] map[K]V

// Values extracts the values as a Mapper for further transformation.
// Order is not guaranteed (map iteration order).
func (e Entries[K, V]) Values() Mapper[V] {
	result := make([]V, 0, len(e))
	for _, v := range e {
		result = append(result, v)
	}

	return result
}

// Keys extracts the keys as a Mapper for further transformation.
// Order is not guaranteed (map iteration order).
func (e Entries[K, V]) Keys() Mapper[K] {
	result := make([]K, 0, len(e))
	for k := range e {
		result = append(result, k)
	}

	return result
}

// KeepIf returns a new Entries containing only the key-value pairs where fn returns true.
func (e Entries[K, V]) KeepIf(fn func(K, V) bool) Entries[K, V] {
	result := make(map[K]V, len(e))
	for k, v := range e {
		if fn(k, v) {
			result[k] = v
		}
	}

	return result
}

// RemoveIf returns a new Entries containing only the key-value pairs where fn returns false.
func (e Entries[K, V]) RemoveIf(fn func(K, V) bool) Entries[K, V] {
	result := make(map[K]V, len(e))
	for k, v := range e {
		if !fn(k, v) {
			result[k] = v
		}
	}

	return result
}

// ToAny returns the result of applying fn to each key-value pair.
func (e Entries[K, V]) ToAny(fn func(K, V) any) Mapper[any] {
	result := make([]any, 0, len(e))
	for k, v := range e {
		result = append(result, fn(k, v))
	}

	return result
}

// ToBool returns the result of applying fn to each key-value pair.
func (e Entries[K, V]) ToBool(fn func(K, V) bool) Mapper[bool] {
	result := make([]bool, 0, len(e))
	for k, v := range e {
		result = append(result, fn(k, v))
	}

	return result
}

// ToByte returns the result of applying fn to each key-value pair.
func (e Entries[K, V]) ToByte(fn func(K, V) byte) Mapper[byte] {
	result := make([]byte, 0, len(e))
	for k, v := range e {
		result = append(result, fn(k, v))
	}

	return result
}

// ToError returns the result of applying fn to each key-value pair.
func (e Entries[K, V]) ToError(fn func(K, V) error) Mapper[error] {
	result := make([]error, 0, len(e))
	for k, v := range e {
		result = append(result, fn(k, v))
	}

	return result
}

// ToFloat32 returns the result of applying fn to each key-value pair.
func (e Entries[K, V]) ToFloat32(fn func(K, V) float32) Mapper[float32] {
	result := make([]float32, 0, len(e))
	for k, v := range e {
		result = append(result, fn(k, v))
	}

	return result
}

// ToFloat64 returns the result of applying fn to each key-value pair.
func (e Entries[K, V]) ToFloat64(fn func(K, V) float64) Float64 {
	result := make([]float64, 0, len(e))
	for k, v := range e {
		result = append(result, fn(k, v))
	}

	return result
}

// ToInt returns the result of applying fn to each key-value pair.
func (e Entries[K, V]) ToInt(fn func(K, V) int) Int {
	result := make([]int, 0, len(e))
	for k, v := range e {
		result = append(result, fn(k, v))
	}

	return result
}

// ToInt32 returns the result of applying fn to each key-value pair.
func (e Entries[K, V]) ToInt32(fn func(K, V) int32) Mapper[int32] {
	result := make([]int32, 0, len(e))
	for k, v := range e {
		result = append(result, fn(k, v))
	}

	return result
}

// ToInt64 returns the result of applying fn to each key-value pair.
func (e Entries[K, V]) ToInt64(fn func(K, V) int64) Mapper[int64] {
	result := make([]int64, 0, len(e))
	for k, v := range e {
		result = append(result, fn(k, v))
	}

	return result
}

// ToRune returns the result of applying fn to each key-value pair.
func (e Entries[K, V]) ToRune(fn func(K, V) rune) Mapper[rune] {
	result := make([]rune, 0, len(e))
	for k, v := range e {
		result = append(result, fn(k, v))
	}

	return result
}

// ToString returns the result of applying fn to each key-value pair.
func (e Entries[K, V]) ToString(fn func(K, V) string) String {
	result := make([]string, 0, len(e))
	for k, v := range e {
		result = append(result, fn(k, v))
	}

	return result
}
