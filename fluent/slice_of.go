package fluent

// SliceOf derives from slice.
// It is usable anywhere a slice is, but provides additional fluent fp methods.
type SliceOf[T any] []T

// Convert applies fn to each element of the slice, returning a new slice of the same element type with the results.
// It is the same as Map, but without changing the element type of the slice.
func (ts SliceOf[T]) Convert(fn func(T) T) SliceOf[T] {
	results := make([]T, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

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

// ToAnys applies the provided function to each element of the slice, mapping it to a slice of `any` type.
func (ts SliceOf[T]) ToAnys(fn func(T) any) SliceOf[any] {
	results := make([]any, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToAnySlices applies the provided function to each element of the slice, mapping it to a slice of `any` slices.
func (ts SliceOf[T]) ToAnySlices(fn func(T) []any) SliceOf[[]any] {
	results := make([][]any, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToBools applies the provided function to each element of the slice, mapping it to a slice of bools.
func (ts SliceOf[T]) ToBools(fn func(T) bool) SliceOf[bool] {
	results := make([]bool, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToBoolSlices applies the provided function to each element of the slice, mapping it to a slice of bool slices.
func (ts SliceOf[T]) ToBoolSlices(fn func(T) []bool) SliceOf[[]bool] {
	results := make([][]bool, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToBytes applies the provided function to each element of the slice, mapping it to a slice of bytes.
func (ts SliceOf[T]) ToBytes(fn func(T) byte) SliceOf[byte] {
	results := make([]byte, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToByteSlices applies the provided function to each element of the slice, mapping it to a slice of byte slices.
func (ts SliceOf[T]) ToByteSlices(fn func(T) []byte) SliceOf[[]byte] {
	results := make([][]byte, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToErrors applies the provided function to each element of the slice, mapping it to a slice of errors.
func (ts SliceOf[T]) ToErrors(fn func(T) error) SliceOf[error] {
	results := make([]error, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToErrorSlices applies the provided function to each element of the slice, mapping it to a slice of error slices.
func (ts SliceOf[T]) ToErrorSlices(fn func(T) []error) SliceOf[[]error] {
	results := make([][]error, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToInts applies the provided function to each element of the slice, mapping it to a slice of ints.
func (ts SliceOf[T]) ToInts(fn func(T) int) SliceOf[int] {
	results := make([]int, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToIntSlices applies the provided function to each element of the slice, mapping it to a slice of int slices.
func (ts SliceOf[T]) ToIntSlices(fn func(T) []int) SliceOf[[]int] {
	results := make([][]int, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToRunes applies the provided function to each element of the slice, mapping it to a slice of runes.
func (ts SliceOf[T]) ToRunes(fn func(T) rune) SliceOf[rune] {
	results := make([]rune, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToRuneSlices applies the provided function to each element of the slice, mapping it to a slice of rune slices.
func (ts SliceOf[T]) ToRuneSlices(fn func(T) []rune) SliceOf[[]rune] {
	results := make([][]rune, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToStrings applies the provided function to each element of the slice, mapping it to a slice of strings.
func (ts SliceOf[T]) ToStrings(fn func(T) string) SliceOf[string] {
	results := make([]string, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToStringSlices applies the provided function to each element of the slice, mapping it to a slice of string slices.
func (ts SliceOf[T]) ToStringSlices(fn func(T) []string) SliceOf[[]string] {
	results := make([][]string, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}
