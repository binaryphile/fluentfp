package fluent

// Mapper is a fluent slice with one additional method, ToOther, for mapping to a specified type R.
// If you don't need to map to an arbitrary type, use SliceOf instead.
type Mapper[T, R any] []T

// Each applies fn to each member of ts.
func (ts Mapper[T, R]) Each(fn func(T)) {
	for _, t := range ts {
		fn(t)
	}
}

// KeepIf returns a new slice containing the members of ts for which fn returns true.
// It is the complement of RemoveIf.
func (ts Mapper[T, R]) KeepIf(fn func(T) bool) Mapper[T, R] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if fn(t) {
			results = append(results, t)
		}
	}
	return results
}

// Len returns the length of the slice.
func (ts Mapper[T, R]) Len() int {
	return len(ts)
}

// RemoveIf returns a new slice containing members for which fn returns false.
// It is the complement of KeepIf.
func (ts Mapper[T, R]) RemoveIf(fn func(T) bool) Mapper[T, R] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if !fn(t) {
			results = append(results, t)
		}
	}
	return results
}

// TakeFirst returns the first n members of ts.
func (ts Mapper[T, R]) TakeFirst(n int) Mapper[T, R] {
	if n > len(ts) {
		n = len(ts)
	}

	return ts[:n]
}

// ToBool returns the result of applying fn to each member of ts.
func (ts Mapper[T, R]) ToBool(fn func(T) bool) Mapper[bool, R] {
	results := make([]bool, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToByte returns the result of applying fn to each member of ts.
func (ts Mapper[T, R]) ToByte(fn func(T) byte) Mapper[byte, R] {
	results := make([]byte, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToError returns the result of applying fn to each member of ts.
func (ts Mapper[T, R]) ToError(fn func(T) error) Mapper[error, R] {
	results := make([]error, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToInt returns the result of applying fn to each member of ts.
func (ts Mapper[T, R]) ToInt(fn func(T) int) Mapper[int, R] {
	results := make([]int, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToOther returns the result of applying fn to each member of ts.
func (ts Mapper[T, R]) ToOther(fn func(T) R) SliceOf[R] {
	results := make([]R, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToRune returns the result of applying fn to each member of ts.
func (ts Mapper[T, R]) ToRune(fn func(T) rune) Mapper[rune, R] {
	results := make([]rune, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToSame returns the result of applying fn to each member of ts.
func (ts Mapper[T, R]) ToSame(fn func(T) T) Mapper[T, R] {
	results := make([]T, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToString returns the result of applying fn to each member of ts.
func (ts Mapper[T, R]) ToString(fn func(T) string) Mapper[string, R] {
	results := make([]string, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}
