package slice

// MapperTo is a fluent slice with one additional method, MapTo, for mapping to a specified type R.
// If you don't need to map to an arbitrary type, use Mapper instead.
type MapperTo[R, T any] []T

func MapTo[R, T any](ts []T) MapperTo[R, T] {
	return ts
}

// Convert returns the result of applying fn to each member of ts.
func (ts MapperTo[R, T]) Convert(fn func(T) T) MapperTo[R, T] {
	results := make([]T, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// Each applies fn to each member of ts.
func (ts MapperTo[R, T]) Each(fn func(T)) {
	for _, t := range ts {
		fn(t)
	}
}

// KeepIf returns a new slice containing the members of ts for which fn returns true.
// It is the complement of RemoveIf.
func (ts MapperTo[R, T]) KeepIf(fn func(T) bool) MapperTo[R, T] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if fn(t) {
			results = append(results, t)
		}
	}
	return results
}

// Len returns the length of the slice.
func (ts MapperTo[T, R]) Len() int {
	return len(ts)
}

// RemoveIf returns a new slice containing members for which fn returns false.
// It is the complement of KeepIf.
func (ts MapperTo[R, T]) RemoveIf(fn func(T) bool) MapperTo[R, T] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if !fn(t) {
			results = append(results, t)
		}
	}
	return results
}

// TakeFirst returns the first n members of ts.
func (ts MapperTo[R, T]) TakeFirst(n int) MapperTo[R, T] {
	if n > len(ts) {
		n = len(ts)
	}

	return ts[:n]
}

// ToAny returns the result of applying fn to each member of ts.
func (ts MapperTo[R, T]) ToAny(fn func(T) any) MapperTo[R, any] {
	results := make([]any, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToBool returns the result of applying fn to each member of ts.
func (ts MapperTo[R, T]) ToBool(fn func(T) bool) MapperTo[R, bool] {
	results := make([]bool, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToByte returns the result of applying fn to each member of ts.
func (ts MapperTo[R, T]) ToByte(fn func(T) byte) MapperTo[R, byte] {
	results := make([]byte, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToError returns the result of applying fn to each member of ts.
func (ts MapperTo[R, T]) ToError(fn func(T) error) MapperTo[R, error] {
	results := make([]error, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToFloat32 returns the result of applying fn to each member of ts.
func (ts MapperTo[R, T]) ToFloat32(fn func(T) float32) MapperTo[R, float32] {
	results := make([]float32, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToFloat64 returns the result of applying fn to each member of ts.
func (ts MapperTo[R, T]) ToFloat64(fn func(T) float64) MapperTo[R, float64] {
	results := make([]float64, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToInt returns the result of applying fn to each member of ts.
func (ts MapperTo[R, T]) ToInt(fn func(T) int) MapperTo[R, int] {
	results := make([]int, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToInt32 returns the result of applying fn to each member of ts.
func (ts MapperTo[R, T]) ToInt32(fn func(T) int32) MapperTo[R, int32] {
	results := make([]int32, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToInt64 returns the result of applying fn to each member of ts.
func (ts MapperTo[R, T]) ToInt64(fn func(T) int64) MapperTo[R, int64] {
	results := make([]int64, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// To returns the result of applying fn to each member of ts.
func (ts MapperTo[R, T]) To(fn func(T) R) Mapper[R] {
	results := make([]R, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToRune returns the result of applying fn to each member of ts.
func (ts MapperTo[R, T]) ToRune(fn func(T) rune) MapperTo[R, rune] {
	results := make([]rune, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToString returns the result of applying fn to each member of ts.
func (ts MapperTo[R, T]) ToString(fn func(T) string) MapperTo[R, string] {
	results := make([]string, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}
