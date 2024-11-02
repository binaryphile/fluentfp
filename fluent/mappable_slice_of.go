package fluent

// MappableSliceOf derives from slice and has a type parameter R that is used solely to specify the return type of the Map method.
type MappableSliceOf[T any, R any] []T

// Convert applies fn to each element of the slice, returning a new slice of the same element type with the results.
// It is the same as Map, but without changing the element type of the slice.
func (ts MappableSliceOf[T, R]) Convert(fn func(T) T) MappableSliceOf[T, R] {
	results := make([]T, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// Each applies fn to each member of ts.
func (ts MappableSliceOf[T, R]) Each(fn func(T)) {
	for _, t := range ts {
		fn(t)
	}
}

// KeepIf returns a new slice containing only the elements for which the provided function returns true.
func (ts MappableSliceOf[T, R]) KeepIf(fn func(T) bool) MappableSliceOf[T, R] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if fn(t) {
			results = append(results, t)
		}
	}
	return results
}

// Len returns the length of the slice.
func (ts MappableSliceOf[T, R]) Len() int {
	return len(ts)
}

// Map applies the provided function to each element of the slice, mapping it to a RawSlice of type R.
func (ts MappableSliceOf[T, R]) Map(fn func(T) R) SliceOf[R] {
	results := make([]R, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// RemoveIf returns a new slice containing only the elements for which the provided function returns false.
func (ts MappableSliceOf[T, R]) RemoveIf(fn func(T) bool) MappableSliceOf[T, R] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if !fn(t) {
			results = append(results, t)
		}
	}
	return results
}

// TakeFirst returns the first n elements of ts.
func (ts MappableSliceOf[T, R]) TakeFirst(n int) MappableSliceOf[T, R] {
	if n > len(ts) {
		n = len(ts)
	}

	return ts[:n]
}

// ToAnys applies the provided function to each element of the slice, mapping it to a slice of `any` type.
func (ts MappableSliceOf[T, R]) ToAnys(fn func(T) any) MappableSliceOf[any, R] {
	results := make([]any, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToAnySlices applies the provided function to each element of the slice, mapping it to a slice of `any` slices.
func (ts MappableSliceOf[T, R]) ToAnySlices(fn func(T) []any) MappableSliceOf[[]any, R] {
	results := make([][]any, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToBools applies the provided function to each element of the slice, mapping it to a slice of bools.
func (ts MappableSliceOf[T, R]) ToBools(fn func(T) bool) MappableSliceOf[bool, R] {
	results := make([]bool, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToBoolSlices applies the provided function to each element of the slice, mapping it to a slice of bool slices.
func (ts MappableSliceOf[T, R]) ToBoolSlices(fn func(T) []bool) MappableSliceOf[[]bool, R] {
	results := make([][]bool, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToBytes applies the provided function to each element of the slice, mapping it to a slice of bytes.
func (ts MappableSliceOf[T, R]) ToBytes(fn func(T) byte) MappableSliceOf[byte, R] {
	results := make([]byte, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToByteSlices applies the provided function to each element of the slice, mapping it to a slice of byte slices.
func (ts MappableSliceOf[T, R]) ToByteSlices(fn func(T) []byte) MappableSliceOf[[]byte, R] {
	results := make([][]byte, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToErrors applies the provided function to each element of the slice, mapping it to a slice of errors.
func (ts MappableSliceOf[T, R]) ToErrors(fn func(T) error) MappableSliceOf[error, R] {
	results := make([]error, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToErrorSlices applies the provided function to each element of the slice, mapping it to a slice of error slices.
func (ts MappableSliceOf[T, R]) ToErrorSlices(fn func(T) []error) MappableSliceOf[[]error, R] {
	results := make([][]error, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToInts applies the provided function to each element of the slice, mapping it to a slice of ints.
func (ts MappableSliceOf[T, R]) ToInts(fn func(T) int) MappableSliceOf[int, R] {
	results := make([]int, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToIntSlices applies the provided function to each element of the slice, mapping it to a slice of int slices.
func (ts MappableSliceOf[T, R]) ToIntSlices(fn func(T) []int) MappableSliceOf[[]int, R] {
	results := make([][]int, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToRunes applies the provided function to each element of the slice, mapping it to a slice of runes.
func (ts MappableSliceOf[T, R]) ToRunes(fn func(T) rune) MappableSliceOf[rune, R] {
	results := make([]rune, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToRuneSlices applies the provided function to each element of the slice, mapping it to a slice of rune slices.
func (ts MappableSliceOf[T, R]) ToRuneSlices(fn func(T) []rune) MappableSliceOf[[]rune, R] {
	results := make([][]rune, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToStrings applies the provided function to each element of the slice, mapping it to a slice of strings.
func (ts MappableSliceOf[T, R]) ToStrings(fn func(T) string) MappableSliceOf[string, R] {
	results := make([]string, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}

// ToStringSlices applies the provided function to each element of the slice, mapping it to a slice of string slices.
func (ts MappableSliceOf[T, R]) ToStringSlices(fn func(T) []string) MappableSliceOf[[]string, R] {
	results := make([][]string, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}
