package fluent

// SliceToNamed is a fluent slice, with an additional ToNamed method that returns a slice of type R.
type SliceToNamed[T any, R any] []T

// Each applies fn to each member of ts.
func (ts SliceToNamed[T, R]) Each(fn func(T)) {
	for _, t := range ts {
		fn(t)
	}
}

// KeepIf returns a new slice containing only the elements for which the provided function returns true.
func (ts SliceToNamed[T, R]) KeepIf(fn func(T) bool) SliceToNamed[T, R] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if fn(t) {
			results = append(results, t)
		}
	}
	return results
}

// Len returns the length of the slice.
func (ts SliceToNamed[T, R]) Len() int {
	return len(ts)
}

// RemoveIf returns a new slice containing only the elements for which the provided function returns false.
func (ts SliceToNamed[T, R]) RemoveIf(fn func(T) bool) SliceToNamed[T, R] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if !fn(t) {
			results = append(results, t)
		}
	}
	return results
}

// TakeFirst returns the first n elements of ts.
func (ts SliceToNamed[T, R]) TakeFirst(n int) SliceToNamed[T, R] {
	if n > len(ts) {
		n = len(ts)
	}

	return ts[:n]
}

// ToAny returns a slice of the results of applying fn to ts.
func (ts SliceToNamed[T, R]) ToAny(fn func(T) any) SliceToNamed[any, R] {
	results := make([]any, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToAnySlice returns a slice of the results of applying fn to ts.
func (ts SliceToNamed[T, R]) ToAnySlice(fn func(T) []any) SliceToNamed[[]any, R] {
	results := make([][]any, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToBool returns a slice of the results of applying fn to ts.
func (ts SliceToNamed[T, R]) ToBool(fn func(T) bool) SliceToNamed[bool, R] {
	results := make([]bool, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToBoolSlice returns a slice of the results of applying fn to ts.
func (ts SliceToNamed[T, R]) ToBoolSlice(fn func(T) []bool) SliceToNamed[[]bool, R] {
	results := make([][]bool, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToByte returns a slice of the results of applying fn to ts.
func (ts SliceToNamed[T, R]) ToByte(fn func(T) byte) SliceToNamed[byte, R] {
	results := make([]byte, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToByteSlice returns a slice of the results of applying fn to ts.
func (ts SliceToNamed[T, R]) ToByteSlice(fn func(T) []byte) SliceToNamed[[]byte, R] {
	results := make([][]byte, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToError returns a slice of the results of applying fn to ts.
func (ts SliceToNamed[T, R]) ToError(fn func(T) error) SliceToNamed[error, R] {
	results := make([]error, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToErrorSlice returns a slice of the results of applying fn to ts.
func (ts SliceToNamed[T, R]) ToErrorSlice(fn func(T) []error) SliceToNamed[[]error, R] {
	results := make([][]error, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToInt returns a slice of the results of applying fn to ts.
func (ts SliceToNamed[T, R]) ToInt(fn func(T) int) SliceToNamed[int, R] {
	results := make([]int, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToIntSlice returns a slice of the results of applying fn to ts.
func (ts SliceToNamed[T, R]) ToIntSlice(fn func(T) []int) SliceToNamed[[]int, R] {
	results := make([][]int, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToNamed returns a slice of the results of applying fn to ts.
func (ts SliceToNamed[T, R]) ToNamed(fn func(T) R) SliceOf[R] {
	results := make([]R, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToRune returns a slice of the results of applying fn to ts.
func (ts SliceToNamed[T, R]) ToRune(fn func(T) rune) SliceToNamed[rune, R] {
	results := make([]rune, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToRuneSlice returns a slice of the results of applying fn to ts.
func (ts SliceToNamed[T, R]) ToRuneSlice(fn func(T) []rune) SliceToNamed[[]rune, R] {
	results := make([][]rune, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToSame returns a slice of the results of applying fn to ts.
func (ts SliceToNamed[T, R]) ToSame(fn func(T) T) SliceToNamed[T, R] {
	results := make([]T, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToString applies the provided function to each element of the slice, mapping it to a slice of strings.
func (ts SliceToNamed[T, R]) ToString(fn func(T) string) SliceToNamed[string, R] {
	results := make([]string, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToStringSlice applies the provided function to each element of the slice, mapping it to a slice of string slices.
func (ts SliceToNamed[T, R]) ToStringSlice(fn func(T) []string) SliceToNamed[[]string, R] {
	results := make([][]string, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}
