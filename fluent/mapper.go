package fluent

// Mapper is a fluent slice, with an additional ToOther method that returns a slice of type R.
type Mapper[T any, R any] []T

// Each applies fn to each member of ts.
func (ts Mapper[T, R]) Each(fn func(T)) {
	for _, t := range ts {
		fn(t)
	}
}

// KeepIf returns a new slice containing only the elements for which the provided function returns true.
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

// RemoveIf returns a new slice containing only the elements for which the provided function returns false.
func (ts Mapper[T, R]) RemoveIf(fn func(T) bool) Mapper[T, R] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if !fn(t) {
			results = append(results, t)
		}
	}
	return results
}

// TakeFirst returns the first n elements of ts.
func (ts Mapper[T, R]) TakeFirst(n int) Mapper[T, R] {
	if n > len(ts) {
		n = len(ts)
	}

	return ts[:n]
}

// ToBool returns a slice of the results of applying fn to ts.
func (ts Mapper[T, R]) ToBool(fn func(T) bool) Mapper[bool, R] {
	results := make([]bool, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToByte returns a slice of the results of applying fn to ts.
func (ts Mapper[T, R]) ToByte(fn func(T) byte) Mapper[byte, R] {
	results := make([]byte, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToInt returns a slice of the results of applying fn to ts.
func (ts Mapper[T, R]) ToInt(fn func(T) int) Mapper[int, R] {
	results := make([]int, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToOther returns a slice of the results of applying fn to ts.
func (ts Mapper[T, R]) ToOther(fn func(T) R) SliceOf[R] {
	results := make([]R, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToRune returns a slice of the results of applying fn to ts.
func (ts Mapper[T, R]) ToRune(fn func(T) rune) Mapper[rune, R] {
	results := make([]rune, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToSame returns a slice of the results of applying fn to ts.
func (ts Mapper[T, R]) ToSame(fn func(T) T) Mapper[T, R] {
	results := make([]T, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToString applies the provided function to each element of the slice, mapping it to a slice of strings.
func (ts Mapper[T, R]) ToString(fn func(T) string) Mapper[string, R] {
	results := make([]string, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}
