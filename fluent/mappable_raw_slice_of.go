package fluent

import "github.com/binaryphile/fluentfp/option"

// MappableRawSliceOf derives from slice and has a type parameter R that is used solely to specify the return type of the Map method.
type MappableRawSliceOf[T any, R any] []T

// Convert applies fn to each element of the slice, returning a new slice of the same element type with the results.
// It is the same as Map, but without changing the element type of the slice.
func (ts MappableRawSliceOf[T, R]) Convert(fn func(T) T) MappableRawSliceOf[T, R] {
	return Map(ts, fn)
}

// KeepIf returns a new slice containing only the elements for which the provided function returns true.
func (ts MappableRawSliceOf[T, R]) KeepIf(fn func(T) bool) MappableRawSliceOf[T, R] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if fn(t) {
			results = append(results, t)
		}
	}
	return results
}

// Len returns the length of the slice.
func (ts MappableRawSliceOf[T, R]) Len() int {
	return len(ts)
}

// MapWith applies the provided function to each element of the slice, mapping it to a RawSlice of type R.
func (ts MappableRawSliceOf[T, R]) MapWith(fn func(T) R) RawSliceOf[R] {
	return Map(ts, fn)
}

// RemoveIf returns a new slice containing only the elements for which the provided function returns false.
func (ts MappableRawSliceOf[T, R]) RemoveIf(fn func(T) bool) MappableRawSliceOf[T, R] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if !fn(t) {
			results = append(results, t)
		}
	}
	return results
}

// TakeFirst returns the first n elements of ts.
func (ts MappableRawSliceOf[T, R]) TakeFirst(n int) MappableRawSliceOf[T, R] {
	if n > len(ts) {
		n = len(ts)
	}

	return ts[:n]
}

// ToAnysWith applies the provided function to each element of the slice, mapping it to a slice of `any` type.
func (ts MappableRawSliceOf[T, R]) ToAnysWith(fn func(T) any) MappableRawSliceOf[any, R] {
	return Map(ts, fn)
}

// ToBoolsWith applies the provided function to each element of the slice, mapping it to a slice of bools.
func (ts MappableRawSliceOf[T, R]) ToBoolsWith(fn func(T) bool) MappableSliceOf[bool, R] {
	return Map(ts, fn)
}

// ToBytesWith applies the provided function to each element of the slice, mapping it to a slice of bytes.
func (ts MappableRawSliceOf[T, R]) ToBytesWith(fn func(T) byte) MappableSliceOf[byte, R] {
	return Map(ts, fn)
}

// ToErrorsWith applies the provided function to each element of the slice, mapping it to a slice of errors.
func (ts MappableRawSliceOf[T, R]) ToErrorsWith(fn func(T) error) MappableRawSliceOf[error, R] {
	return Map(ts, fn)
}

// ToIntsWith applies the provided function to each element of the slice, mapping it to a slice of ints.
func (ts MappableRawSliceOf[T, R]) ToIntsWith(fn func(T) int) MappableSliceOf[int, R] {
	return Map(ts, fn)
}

// ToInt8sWith applies the provided function to each element of the slice, mapping it to a slice of int8s.
func (ts MappableRawSliceOf[T, R]) ToInt8sWith(fn func(T) int8) MappableSliceOf[int8, R] {
	return Map(ts, fn)
}

// ToInt16sWith applies the provided function to each element of the slice, mapping it to a slice of int16s.
func (ts MappableRawSliceOf[T, R]) ToInt16sWith(fn func(T) int16) MappableSliceOf[int16, R] {
	return Map(ts, fn)
}

// ToInt32sWith applies the provided function to each element of the slice, mapping it to a slice of int32s.
func (ts MappableRawSliceOf[T, R]) ToInt32sWith(fn func(T) int32) MappableSliceOf[int32, R] {
	return Map(ts, fn)
}

// ToInt64sWith applies the provided function to each element of the slice, mapping it to a slice of int64s.
func (ts MappableRawSliceOf[T, R]) ToInt64sWith(fn func(T) int64) MappableSliceOf[int64, R] {
	return Map(ts, fn)
}

// ToUintsWith applies the provided function to each element of the slice, mapping it to a slice of uints.
func (ts MappableRawSliceOf[T, R]) ToUintsWith(fn func(T) uint) MappableSliceOf[uint, R] {
	return Map(ts, fn)
}

// ToUint8sWith applies the provided function to each element of the slice, mapping it to a slice of uint8s.
func (ts MappableRawSliceOf[T, R]) ToUint8sWith(fn func(T) uint8) MappableSliceOf[uint8, R] {
	return Map(ts, fn)
}

// ToUint16sWith applies the provided function to each element of the slice, mapping it to a slice of uint16s.
func (ts MappableRawSliceOf[T, R]) ToUint16sWith(fn func(T) uint16) MappableSliceOf[uint16, R] {
	return Map(ts, fn)
}

// ToUint32sWith applies the provided function to each element of the slice, mapping it to a slice of uint32s.
func (ts MappableRawSliceOf[T, R]) ToUint32sWith(fn func(T) uint32) MappableSliceOf[uint32, R] {
	return Map(ts, fn)
}

// ToUint64sWith applies the provided function to each element of the slice, mapping it to a slice of uint64s.
func (ts MappableRawSliceOf[T, R]) ToUint64sWith(fn func(T) uint64) MappableSliceOf[uint64, R] {
	return Map(ts, fn)
}

// ToUintptrsWith applies the provided function to each element of the slice, mapping it to a slice of uintptrs.
func (ts MappableRawSliceOf[T, R]) ToUintptrsWith(fn func(T) uintptr) MappableSliceOf[uintptr, R] {
	return Map(ts, fn)
}

// ToFloat32sWith applies the provided function to each element of the slice, mapping it to a slice of float32s.
func (ts MappableRawSliceOf[T, R]) ToFloat32sWith(fn func(T) float32) MappableSliceOf[float32, R] {
	return Map(ts, fn)
}

// ToFloat64sWith applies the provided function to each element of the slice, mapping it to a slice of float64s.
func (ts MappableRawSliceOf[T, R]) ToFloat64sWith(fn func(T) float64) MappableSliceOf[float64, R] {
	return Map(ts, fn)
}

// ToComplex64sWith applies the provided function to each element of the slice, mapping it to a slice of complex64s.
func (ts MappableRawSliceOf[T, R]) ToComplex64sWith(fn func(T) complex64) MappableSliceOf[complex64, R] {
	return Map(ts, fn)
}

// ToComplex128sWith applies the provided function to each element of the slice, mapping it to a slice of complex128s.
func (ts MappableRawSliceOf[T, R]) ToComplex128sWith(fn func(T) complex128) MappableSliceOf[complex128, R] {
	return Map(ts, fn)
}

// ToRunesWith applies the provided function to each element of the slice, mapping it to a slice of runes.
func (ts MappableRawSliceOf[T, R]) ToRunesWith(fn func(T) rune) MappableSliceOf[rune, R] {
	return Map(ts, fn)
}

// ToStringsWith applies the provided function to each element of the slice, mapping it to a slice of strings.
func (ts MappableRawSliceOf[T, R]) ToStringsWith(fn func(T) string) MappableSliceOf[string, R] {
	return Map(ts, fn)
}

// ToBoolOptionsWith applies the provided function to each element of the slice, mapping it to a slice of bool options.
func (ts MappableRawSliceOf[T, R]) ToBoolOptionsWith(fn func(T) option.Bool) MappableSliceOf[option.Bool, R] {
	return Map(ts, fn)
}

// ToByteOptionsWith applies the provided function to each element of the slice, mapping it to a slice of byte options.
func (ts MappableRawSliceOf[T, R]) ToByteOptionsWith(fn func(T) option.Byte) MappableSliceOf[option.Byte, R] {
	return Map(ts, fn)
}

// ToIntOptionsWith applies the provided function to each element of the slice, mapping it to a slice of int options.
func (ts MappableRawSliceOf[T, R]) ToIntOptionsWith(fn func(T) option.Int) MappableSliceOf[option.Int, R] {
	return Map(ts, fn)
}

// ToInt8OptionsWith applies the provided function to each element of the slice, mapping it to a slice of int8 options.
func (ts MappableRawSliceOf[T, R]) ToInt8OptionsWith(fn func(T) option.Int8) MappableSliceOf[option.Int8, R] {
	return Map(ts, fn)
}

// ToInt16OptionsWith applies the provided function to each element of the slice, mapping it to a slice of int16 options.
func (ts MappableRawSliceOf[T, R]) ToInt16OptionsWith(fn func(T) option.Int16) MappableSliceOf[option.Int16, R] {
	return Map(ts, fn)
}

// ToInt32OptionsWith applies the provided function to each element of the slice, mapping it to a slice of int32 options.
func (ts MappableRawSliceOf[T, R]) ToInt32OptionsWith(fn func(T) option.Int32) MappableSliceOf[option.Int32, R] {
	return Map(ts, fn)
}

// ToInt64OptionsWith applies the provided function to each element of the slice, mapping it to a slice of int64 options.
func (ts MappableRawSliceOf[T, R]) ToInt64OptionsWith(fn func(T) option.Int64) MappableSliceOf[option.Int64, R] {
	return Map(ts, fn)
}

// ToUintOptionsWith applies the provided function to each element of the slice, mapping it to a slice of uint options.
func (ts MappableRawSliceOf[T, R]) ToUintOptionsWith(fn func(T) option.Uint) MappableSliceOf[option.Uint, R] {
	return Map(ts, fn)
}

// ToUint8OptionsWith applies the provided function to each element of the slice, mapping it to a slice of uint8 options.
func (ts MappableRawSliceOf[T, R]) ToUint8OptionsWith(fn func(T) option.Uint8) MappableSliceOf[option.Uint8, R] {
	return Map(ts, fn)
}

// ToUint16OptionsWith applies the provided function to each element of the slice, mapping it to a slice of uint16 options.
func (ts MappableRawSliceOf[T, R]) ToUint16OptionsWith(fn func(T) option.Uint16) MappableSliceOf[option.Uint16, R] {
	return Map(ts, fn)
}

// ToUint32OptionsWith applies the provided function to each element of the slice, mapping it to a slice of uint32 options.
func (ts MappableRawSliceOf[T, R]) ToUint32OptionsWith(fn func(T) option.Uint32) MappableSliceOf[option.Uint32, R] {
	return Map(ts, fn)
}

// ToUint64OptionsWith applies the provided function to each element of the slice, mapping it to a slice of uint64 options.
func (ts MappableRawSliceOf[T, R]) ToUint64OptionsWith(fn func(T) option.Uint64) MappableSliceOf[option.Uint64, R] {
	return Map(ts, fn)
}

// ToUintptrOptionsWith applies the provided function to each element of the slice, mapping it to a slice of uintptr options.
func (ts MappableRawSliceOf[T, R]) ToUintptrOptionsWith(fn func(T) option.Uintptr) MappableSliceOf[option.Uintptr, R] {
	return Map(ts, fn)
}

// ToFloat32OptionsWith applies the provided function to each element of the slice, mapping it to a slice of float32 options.
func (ts MappableRawSliceOf[T, R]) ToFloat32OptionsWith(fn func(T) option.Float32) MappableSliceOf[option.Float32, R] {
	return Map(ts, fn)
}

// ToFloat64OptionsWith applies the provided function to each element of the slice, mapping it to a slice of float64 options.
func (ts MappableRawSliceOf[T, R]) ToFloat64OptionsWith(fn func(T) option.Float64) MappableSliceOf[option.Float64, R] {
	return Map(ts, fn)
}

// ToComplex64OptionsWith applies the provided function to each element of the slice, mapping it to a slice of complex64 options.
func (ts MappableRawSliceOf[T, R]) ToComplex64OptionsWith(fn func(T) option.Complex64) MappableSliceOf[option.Complex64, R] {
	return Map(ts, fn)
}

// ToComplex128OptionsWith applies the provided function to each element of the slice, mapping it to a slice of complex128 options.
func (ts MappableRawSliceOf[T, R]) ToComplex128OptionsWith(fn func(T) option.Complex128) MappableSliceOf[option.Complex128, R] {
	return Map(ts, fn)
}

// ToRuneOptionsWith applies the provided function to each element of the slice, mapping it to a slice of rune options.
func (ts MappableRawSliceOf[T, R]) ToRuneOptionsWith(fn func(T) option.Rune) MappableSliceOf[option.Rune, R] {
	return Map(ts, fn)
}

// ToStringOptionsWith applies the provided function to each element of the slice, mapping it to a slice of string options.
func (ts MappableRawSliceOf[T, R]) ToStringOptionsWith(fn func(T) option.String) MappableSliceOf[option.String, R] {
	return Map(ts, fn)
}

// ToAnyOptionsWith applies the provided function to each element of the slice, mapping it to a slice of any options.
func (ts MappableRawSliceOf[T, R]) ToAnyOptionsWith(fn func(T) option.Any) MappableSliceOf[option.Any, R] {
	return Map(ts, fn)
}

// ToErrorOptionsWith applies the provided function to each element of the slice, mapping it to a slice of error options.
func (ts MappableRawSliceOf[T, R]) ToErrorOptionsWith(fn func(T) option.Error) MappableSliceOf[option.Error, R] {
	return Map(ts, fn)
}

// ToStringSlicesWith applies the provided function to each element of the slice, mapping it to a slice of string slices.
func (ts MappableRawSliceOf[T, R]) ToStringSlicesWith(fn func(T) []string) MappableRawSliceOf[[]string, R] {
	return Map(ts, fn)
}

// ToBoolSlicesWith applies the provided function to each element of the slice, mapping it to a slice of bool slices.
func (ts MappableRawSliceOf[T, R]) ToBoolSlicesWith(fn func(T) []bool) MappableRawSliceOf[[]bool, R] {
	return Map(ts, fn)
}

// ToIntSlicesWith applies the provided function to each element of the slice, mapping it to a slice of int slices.
func (ts MappableRawSliceOf[T, R]) ToIntSlicesWith(fn func(T) []int) MappableRawSliceOf[[]int, R] {
	return Map(ts, fn)
}

// ToInt8SlicesWith applies the provided function to each element of the slice, mapping it to a slice of int8 slices.
func (ts MappableRawSliceOf[T, R]) ToInt8SlicesWith(fn func(T) []int8) MappableRawSliceOf[[]int8, R] {
	return Map(ts, fn)
}

// ToInt16SlicesWith applies the provided function to each element of the slice, mapping it to a slice of int16 slices.
func (ts MappableRawSliceOf[T, R]) ToInt16SlicesWith(fn func(T) []int16) MappableRawSliceOf[[]int16, R] {
	return Map(ts, fn)
}

// ToInt32SlicesWith applies the provided function to each element of the slice, mapping it to a slice of int32 slices.
func (ts MappableRawSliceOf[T, R]) ToInt32SlicesWith(fn func(T) []int32) MappableRawSliceOf[[]int32, R] {
	return Map(ts, fn)
}

// ToInt64SlicesWith applies the provided function to each element of the slice, mapping it to a slice of int64 slices.
func (ts MappableRawSliceOf[T, R]) ToInt64SlicesWith(fn func(T) []int64) MappableRawSliceOf[[]int64, R] {
	return Map(ts, fn)
}

// ToUintSlicesWith applies the provided function to each element of the slice, mapping it to a slice of uint slices.
func (ts MappableRawSliceOf[T, R]) ToUintSlicesWith(fn func(T) []uint) MappableRawSliceOf[[]uint, R] {
	return Map(ts, fn)
}

// ToUint8SlicesWith applies the provided function to each element of the slice, mapping it to a slice of uint8 slices.
func (ts MappableRawSliceOf[T, R]) ToUint8SlicesWith(fn func(T) []uint8) MappableRawSliceOf[[]uint8, R] {
	return Map(ts, fn)
}

// ToUint16SlicesWith applies the provided function to each element of the slice, mapping it to a slice of uint16 slices.
func (ts MappableRawSliceOf[T, R]) ToUint16SlicesWith(fn func(T) []uint16) MappableRawSliceOf[[]uint16, R] {
	return Map(ts, fn)
}

// ToUint32SlicesWith applies the provided function to each element of the slice, mapping it to a slice of uint32 slices.
func (ts MappableRawSliceOf[T, R]) ToUint32SlicesWith(fn func(T) []uint32) MappableRawSliceOf[[]uint32, R] {
	return Map(ts, fn)
}

// ToUint64SlicesWith applies the provided function to each element of the slice, mapping it to a slice of uint64 slices.
func (ts MappableRawSliceOf[T, R]) ToUint64SlicesWith(fn func(T) []uint64) MappableRawSliceOf[[]uint64, R] {
	return Map(ts, fn)
}

// ToUintptrSlicesWith applies the provided function to each element of the slice, mapping it to a slice of uintptr slices.
func (ts MappableRawSliceOf[T, R]) ToUintptrSlicesWith(fn func(T) []uintptr) MappableRawSliceOf[[]uintptr, R] {
	return Map(ts, fn)
}

// ToFloat32SlicesWith applies the provided function to each element of the slice, mapping it to a slice of float32 slices.
func (ts MappableRawSliceOf[T, R]) ToFloat32SlicesWith(fn func(T) []float32) MappableRawSliceOf[[]float32, R] {
	return Map(ts, fn)
}

// ToFloat64SlicesWith applies the provided function to each element of the slice, mapping it to a slice of float64 slices.
func (ts MappableRawSliceOf[T, R]) ToFloat64SlicesWith(fn func(T) []float64) MappableRawSliceOf[[]float64, R] {
	return Map(ts, fn)
}

// ToComplex64SlicesWith applies the provided function to each element of the slice, mapping it to a slice of complex64 slices.
func (ts MappableRawSliceOf[T, R]) ToComplex64SlicesWith(fn func(T) []complex64) MappableRawSliceOf[[]complex64, R] {
	return Map(ts, fn)
}

// ToComplex128SlicesWith applies the provided function to each element of the slice, mapping it to a slice of complex128 slices.
func (ts MappableRawSliceOf[T, R]) ToComplex128SlicesWith(fn func(T) []complex128) MappableRawSliceOf[[]complex128, R] {
	return Map(ts, fn)
}

// ToByteSlicesWith applies the provided function to each element of the slice, mapping it to a slice of byte slices.
func (ts MappableRawSliceOf[T, R]) ToByteSlicesWith(fn func(T) []byte) MappableRawSliceOf[[]byte, R] {
	return Map(ts, fn)
}

// ToRuneSlicesWith applies the provided function to each element of the slice, mapping it to a slice of rune slices.
func (ts MappableRawSliceOf[T, R]) ToRuneSlicesWith(fn func(T) []rune) MappableRawSliceOf[[]rune, R] {
	return Map(ts, fn)
}
