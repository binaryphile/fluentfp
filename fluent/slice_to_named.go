package fluent

// SliceToNamed is a fluent slice, with a ToNamed method that returns a slice of type R.
type SliceToNamed[T any, R any] []T

// Convert applies fn to each element of the slice, returning a new slice of the same element type with the results.
// It is the same as Map, but without changing the element type of the slice.
func (ts SliceToNamed[T, R]) Convert(fn func(T) T) SliceToNamed[T, R] {
	results := make([]T, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

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

// Map applies the provided function to each element of the slice, mapping it to a RawSlice of type R.
func (ts SliceToNamed[T, R]) Map(fn func(T) R) SliceOf[R] {
	results := make([]R, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
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

// ToAnys applies the provided function to each element of the slice, mapping it to a slice of `any` type.
func (ts SliceToNamed[T, R]) ToAnys(fn func(T) any) SliceToNamed[any, R] {
	results := make([]any, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToAnySlices applies the provided function to each element of the slice, mapping it to a slice of `any` slices.
func (ts SliceToNamed[T, R]) ToAnySlices(fn func(T) []any) SliceToNamed[[]any, R] {
	results := make([][]any, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToBools applies the provided function to each element of the slice, mapping it to a slice of bools.
func (ts SliceToNamed[T, R]) ToBools(fn func(T) bool) SliceToNamed[bool, R] {
	results := make([]bool, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToBoolSlices applies the provided function to each element of the slice, mapping it to a slice of bool slices.
func (ts SliceToNamed[T, R]) ToBoolSlices(fn func(T) []bool) SliceToNamed[[]bool, R] {
	results := make([][]bool, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToBytes applies the provided function to each element of the slice, mapping it to a slice of bytes.
func (ts SliceToNamed[T, R]) ToBytes(fn func(T) byte) SliceToNamed[byte, R] {
	results := make([]byte, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToByteSlices applies the provided function to each element of the slice, mapping it to a slice of byte slices.
func (ts SliceToNamed[T, R]) ToByteSlices(fn func(T) []byte) SliceToNamed[[]byte, R] {
	results := make([][]byte, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToErrors applies the provided function to each element of the slice, mapping it to a slice of errors.
func (ts SliceToNamed[T, R]) ToErrors(fn func(T) error) SliceToNamed[error, R] {
	results := make([]error, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToErrorSlices applies the provided function to each element of the slice, mapping it to a slice of error slices.
func (ts SliceToNamed[T, R]) ToErrorSlices(fn func(T) []error) SliceToNamed[[]error, R] {
	results := make([][]error, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToInts applies the provided function to each element of the slice, mapping it to a slice of ints.
func (ts SliceToNamed[T, R]) ToInts(fn func(T) int) SliceToNamed[int, R] {
	results := make([]int, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToIntSlices applies the provided function to each element of the slice, mapping it to a slice of int slices.
func (ts SliceToNamed[T, R]) ToIntSlices(fn func(T) []int) SliceToNamed[[]int, R] {
	results := make([][]int, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToRunes applies the provided function to each element of the slice, mapping it to a slice of runes.
func (ts SliceToNamed[T, R]) ToRunes(fn func(T) rune) SliceToNamed[rune, R] {
	results := make([]rune, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToRuneSlices applies the provided function to each element of the slice, mapping it to a slice of rune slices.
func (ts SliceToNamed[T, R]) ToRuneSlices(fn func(T) []rune) SliceToNamed[[]rune, R] {
	results := make([][]rune, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToStrings applies the provided function to each element of the slice, mapping it to a slice of strings.
func (ts SliceToNamed[T, R]) ToStrings(fn func(T) string) SliceToNamed[string, R] {
	results := make([]string, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToStringSlices applies the provided function to each element of the slice, mapping it to a slice of string slices.
func (ts SliceToNamed[T, R]) ToStringSlices(fn func(T) []string) SliceToNamed[[]string, R] {
	results := make([][]string, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}
