package fluent

// SliceOf derives from slice.
// It is usable anywhere a slice is, but provides additional fluent fp methods.
type SliceOf[T any] []T

// Each applies fn to each member of ts.
func (ts SliceOf[T]) Each(fn func(T)) {
	for _, t := range ts {
		fn(t)
	}
}

// KeepIf returns the slice of elements from ts for which fn returns true.
func (ts SliceOf[T]) KeepIf(fn func(T) bool) SliceOf[T] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if fn(t) {
			results = append(results, t)
		}
	}

	return results
}

// Len returns the length of the slice.
func (ts SliceOf[T]) Len() int {
	return len(ts)
}

// RemoveIf returns the slice of elements from ts for which fn returns false.
func (ts SliceOf[T]) RemoveIf(fn func(T) bool) SliceOf[T] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if !fn(t) {
			results = append(results, t)
		}
	}

	return results
}

// TakeFirst returns the first n elements of ts.
func (ts SliceOf[T]) TakeFirst(n int) SliceOf[T] {
	if n > len(ts) {
		n = len(ts)
	}

	return ts[:n]
}

// ToAny returns a slice of the results of applying fn to ts.
func (ts SliceOf[T]) ToAny(fn func(T) any) SliceOf[any] {
	results := make([]any, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToAnySlice returns a slice of the results of applying fn to ts.
func (ts SliceOf[T]) ToAnySlice(fn func(T) []any) SliceOf[[]any] {
	results := make([][]any, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToBool returns a slice of the results of applying fn to ts.
func (ts SliceOf[T]) ToBool(fn func(T) bool) SliceOf[bool] {
	results := make([]bool, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToBoolSlice returns a slice of the results of applying fn to ts.
func (ts SliceOf[T]) ToBoolSlice(fn func(T) []bool) SliceOf[[]bool] {
	results := make([][]bool, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToByte returns a slice of the results of applying fn to ts.
func (ts SliceOf[T]) ToByte(fn func(T) byte) SliceOf[byte] {
	results := make([]byte, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToByteSlice returns a slice of the results of applying fn to ts.
func (ts SliceOf[T]) ToByteSlice(fn func(T) []byte) SliceOf[[]byte] {
	results := make([][]byte, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToError returns a slice of the results of applying fn to ts.
func (ts SliceOf[T]) ToError(fn func(T) error) SliceOf[error] {
	results := make([]error, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToErrorSlice returns a slice of the results of applying fn to ts.
func (ts SliceOf[T]) ToErrorSlice(fn func(T) []error) SliceOf[[]error] {
	results := make([][]error, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToInt returns a slice of the results of applying fn to ts.
func (ts SliceOf[T]) ToInt(fn func(T) int) SliceOf[int] {
	results := make([]int, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToIntSlice returns a slice of the results of applying fn to ts.
func (ts SliceOf[T]) ToIntSlice(fn func(T) []int) SliceOf[[]int] {
	results := make([][]int, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToRune returns a slice of the results of applying fn to ts.
func (ts SliceOf[T]) ToRune(fn func(T) rune) SliceOf[rune] {
	results := make([]rune, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToRuneSlice returns a slice of the results of applying fn to ts.
func (ts SliceOf[T]) ToRuneSlice(fn func(T) []rune) SliceOf[[]rune] {
	results := make([][]rune, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToSame returns a slice of the results of applying fn to ts.
func (ts SliceOf[T]) ToSame(fn func(T) T) SliceOf[T] {
	results := make([]T, len(ts))
	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// ToString returns a slice of the results of applying fn to ts.
func (ts SliceOf[T]) ToString(fn func(T) string) SliceOf[string] {
	results := make([]string, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToStringSlice returns a slice of the results of applying fn to ts.
func (ts SliceOf[T]) ToStringSlice(fn func(T) []string) SliceOf[[]string] {
	results := make([][]string, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}
