package slice

// Mapper is a fluent slice usable anywhere a regular slice is, but provides additional fluent fp methods.
// Its underlying type is []T.
type Mapper[T any] []T

func Of[T any](ts []T) Mapper[T] {
	return ts
}

// Convert returns the result of applying fn to each member of ts.
func (ts Mapper[T]) Convert(fn func(T) T) Mapper[T] {
	results := make([]T, len(ts))
	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// Each applies fn to each member of ts.
func (ts Mapper[T]) Each(fn func(T)) {
	for _, t := range ts {
		fn(t)
	}
}

// KeepIf returns a new slice containing the members of ts for which fn returns true.
// It is the complement of RemoveIf.
func (ts Mapper[T]) KeepIf(fn func(T) bool) Mapper[T] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if fn(t) {
			results = append(results, t)
		}
	}

	return results
}

// Len returns the length of the slice.
func (ts Mapper[T]) Len() int {
	return len(ts)
}

// RemoveIf returns a new slice containing members for which fn returns false.
// It is the complement of KeepIf.
func (ts Mapper[T]) RemoveIf(fn func(T) bool) Mapper[T] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if !fn(t) {
			results = append(results, t)
		}
	}

	return results
}

// TakeFirst returns the first n elements of ts.
func (ts Mapper[T]) TakeFirst(n int) Mapper[T] {
	if n > len(ts) {
		n = len(ts)
	}

	return ts[:n]
}

// ToAny returns the result of applying fn to each member of ts.
func (ts Mapper[T]) ToAny(fn func(T) any) Mapper[any] {
	results := make([]any, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToBool returns the result of applying fn to each member of ts.
func (ts Mapper[T]) ToBool(fn func(T) bool) Mapper[bool] {
	results := make([]bool, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToByte returns the result of applying fn to each member of ts.
func (ts Mapper[T]) ToByte(fn func(T) byte) Mapper[byte] {
	results := make([]byte, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToError returns the result of applying fn to each member of ts.
func (ts Mapper[T]) ToError(fn func(T) error) Mapper[error] {
	results := make([]error, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToInt returns the result of applying fn to each member of ts.
func (ts Mapper[T]) ToInt(fn func(T) int) Mapper[int] {
	results := make([]int, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToRune returns the result of applying fn to each member of ts.
func (ts Mapper[T]) ToRune(fn func(T) rune) Mapper[rune] {
	results := make([]rune, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToString returns the result of applying fn to each member of ts.
func (ts Mapper[T]) ToString(fn func(T) string) Mapper[string] {
	results := make([]string, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}
