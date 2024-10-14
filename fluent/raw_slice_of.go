package fluent

import "github.com/binaryphile/fluentfp/option"

// RawSliceOf derives from slice.
// It is usable anywhere a slice is, but provides additional fluent fp methods.
// It is raw because it takes any type, so it is unable to provide methods like Contains,
// which relies on comparability.
type RawSliceOf[T any] []T

// Convert applies fn to each element of the slice, returning a new slice of the same element type with the results.
// It is the same as Map, but without changing the element type of the slice.
func (ts RawSliceOf[T]) Convert(fn func(T) T) RawSliceOf[T] {
	results := make([]T, len(ts))

	for i := range ts {
		results[i] = fn(ts[i])
	}

	return results
}

// Each applies fn to each member of ts.
func (ts RawSliceOf[T]) Each(fn func(T)) {
	for _, t := range ts {
		fn(t)
	}
}

// KeepIf returns the slice of elements from ts for which fn returns true.
func (ts RawSliceOf[T]) KeepIf(fn func(T) bool) RawSliceOf[T] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if fn(t) {
			results = append(results, t)
		}
	}

	return results
}

// Len returns the length of the slice.
func (ts RawSliceOf[T]) Len() int {
	return len(ts)
}

// RemoveIf returns the slice of elements from ts for which fn returns false.
func (ts RawSliceOf[T]) RemoveIf(fn func(T) bool) RawSliceOf[T] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if !fn(t) {
			results = append(results, t)
		}
	}

	return results
}

// TakeFirst returns the first n elements of ts.
func (ts RawSliceOf[T]) TakeFirst(n int) RawSliceOf[T] {
	if n > len(ts) {
		n = len(ts)
	}

	return ts[:n]
}

// ToAnysWith applies the provided function to each element of the slice, mapping it to a slice of `any` type.
func (ts RawSliceOf[T]) ToAnysWith(fn func(T) any) RawSliceOf[any] {
	return Map(ts, fn)
}

// ToBoolsWith applies the provided function to each element of the slice, mapping it to a slice of bools.
func (ts RawSliceOf[T]) ToBoolsWith(fn func(T) bool) SliceOf[bool] {
	return Map(ts, fn)
}

// ToBytesWith applies the provided function to each element of the slice, mapping it to a slice of bytes.
func (ts RawSliceOf[T]) ToBytesWith(fn func(T) byte) SliceOf[byte] {
	return Map(ts, fn)
}

// ToErrorsWith applies the provided function to each element of the slice, mapping it to a slice of errors.
func (ts RawSliceOf[T]) ToErrorsWith(fn func(T) error) RawSliceOf[error] {
	return Map(ts, fn)
}

// ToIntsWith applies the provided function to each element of the slice, mapping it to a slice of ints.
func (ts RawSliceOf[T]) ToIntsWith(fn func(T) int) SliceOf[int] {
	return Map(ts, fn)
}

// ToInt8sWith applies the provided function to each element of the slice, mapping it to a slice of int8s.
func (ts RawSliceOf[T]) ToInt8sWith(fn func(T) int8) SliceOf[int8] {
	return Map(ts, fn)
}

// ToInt16sWith applies the provided function to each element of the slice, mapping it to a slice of int16s.
func (ts RawSliceOf[T]) ToInt16sWith(fn func(T) int16) SliceOf[int16] {
	return Map(ts, fn)
}

// ToInt32sWith applies the provided function to each element of the slice, mapping it to a slice of int32s.
func (ts RawSliceOf[T]) ToInt32sWith(fn func(T) int32) SliceOf[int32] {
	return Map(ts, fn)
}

// ToInt64sWith applies the provided function to each element of the slice, mapping it to a slice of int64s.
func (ts RawSliceOf[T]) ToInt64sWith(fn func(T) int64) SliceOf[int64] {
	return Map(ts, fn)
}

// ToUintsWith applies the provided function to each element of the slice, mapping it to a slice of uints.
func (ts RawSliceOf[T]) ToUintsWith(fn func(T) uint) SliceOf[uint] {
	return Map(ts, fn)
}

// ToUint8sWith applies the provided function to each element of the slice, mapping it to a slice of uint8s.
func (ts RawSliceOf[T]) ToUint8sWith(fn func(T) uint8) SliceOf[uint8] {
	return Map(ts, fn)
}

// ToUint16sWith applies the provided function to each element of the slice, mapping it to a slice of uint16s.
func (ts RawSliceOf[T]) ToUint16sWith(fn func(T) uint16) SliceOf[uint16] {
	return Map(ts, fn)
}

// ToUint32sWith applies the provided function to each element of the slice, mapping it to a slice of uint32s.
func (ts RawSliceOf[T]) ToUint32sWith(fn func(T) uint32) SliceOf[uint32] {
	return Map(ts, fn)
}

// ToUint64sWith applies the provided function to each element of the slice, mapping it to a slice of uint64s.
func (ts RawSliceOf[T]) ToUint64sWith(fn func(T) uint64) SliceOf[uint64] {
	return Map(ts, fn)
}

// ToUintptrsWith applies the provided function to each element of the slice, mapping it to a slice of uintptrs.
func (ts RawSliceOf[T]) ToUintptrsWith(fn func(T) uintptr) SliceOf[uintptr] {
	return Map(ts, fn)
}

// ToFloat32sWith applies the provided function to each element of the slice, mapping it to a slice of float32s.
func (ts RawSliceOf[T]) ToFloat32sWith(fn func(T) float32) SliceOf[float32] {
	return Map(ts, fn)
}

// ToFloat64sWith applies the provided function to each element of the slice, mapping it to a slice of float64s.
func (ts RawSliceOf[T]) ToFloat64sWith(fn func(T) float64) SliceOf[float64] {
	return Map(ts, fn)
}

// ToComplex64sWith applies the provided function to each element of the slice, mapping it to a slice of complex64s.
func (ts RawSliceOf[T]) ToComplex64sWith(fn func(T) complex64) SliceOf[complex64] {
	return Map(ts, fn)
}

// ToComplex128sWith applies the provided function to each element of the slice, mapping it to a slice of complex128s.
func (ts RawSliceOf[T]) ToComplex128sWith(fn func(T) complex128) SliceOf[complex128] {
	return Map(ts, fn)
}

// ToRunesWith applies the provided function to each element of the slice, mapping it to a slice of runes.
func (ts RawSliceOf[T]) ToRunesWith(fn func(T) rune) SliceOf[rune] {
	return Map(ts, fn)
}

// ToStringsWith applies the provided function to each element of the slice, mapping it to a slice of strings.
func (ts RawSliceOf[T]) ToStringsWith(fn func(T) string) SliceOf[string] {
	return Map(ts, fn)
}

// ToBoolOptionsWith applies the provided function to each element of the slice, mapping it to a slice of bool options.
func (ts RawSliceOf[T]) ToBoolOptionsWith(fn func(T) option.Bool) SliceOf[option.Bool] {
	return Map(ts, fn)
}

// ToByteOptionsWith applies the provided function to each element of the slice, mapping it to a slice of byte options.
func (ts RawSliceOf[T]) ToByteOptionsWith(fn func(T) option.Byte) SliceOf[option.Byte] {
	return Map(ts, fn)
}

// ToIntOptionsWith applies the provided function to each element of the slice, mapping it to a slice of int options.
func (ts RawSliceOf[T]) ToIntOptionsWith(fn func(T) option.Int) SliceOf[option.Int] {
	return Map(ts, fn)
}

// ToInt8OptionsWith applies the provided function to each element of the slice, mapping it to a slice of int8 options.
func (ts RawSliceOf[T]) ToInt8OptionsWith(fn func(T) option.Int8) SliceOf[option.Int8] {
	return Map(ts, fn)
}

// ToInt16OptionsWith applies the provided function to each element of the slice, mapping it to a slice of int16 options.
func (ts RawSliceOf[T]) ToInt16OptionsWith(fn func(T) option.Int16) SliceOf[option.Int16] {
	return Map(ts, fn)
}

// ToInt32OptionsWith applies the provided function to each element of the slice, mapping it to a slice of int32 options.
func (ts RawSliceOf[T]) ToInt32OptionsWith(fn func(T) option.Int32) SliceOf[option.Int32] {
	return Map(ts, fn)
}

// ToInt64OptionsWith applies the provided function to each element of the slice, mapping it to a slice of int64 options.
func (ts RawSliceOf[T]) ToInt64OptionsWith(fn func(T) option.Int64) SliceOf[option.Int64] {
	return Map(ts, fn)
}

// ToUintOptionsWith applies the provided function to each element of the slice, mapping it to a slice of uint options.
func (ts RawSliceOf[T]) ToUintOptionsWith(fn func(T) option.Uint) SliceOf[option.Uint] {
	return Map(ts, fn)
}

// ToUint8OptionsWith applies the provided function to each element of the slice, mapping it to a slice of uint8 options.
func (ts RawSliceOf[T]) ToUint8OptionsWith(fn func(T) option.Uint8) SliceOf[option.Uint8] {
	return Map(ts, fn)
}

// ToUint16OptionsWith applies the provided function to each element of the slice, mapping it to a slice of uint16 options.
func (ts RawSliceOf[T]) ToUint16OptionsWith(fn func(T) option.Uint16) SliceOf[option.Uint16] {
	return Map(ts, fn)
}

// ToUint32OptionsWith applies the provided function to each element of the slice, mapping it to a slice of uint32 options.
func (ts RawSliceOf[T]) ToUint32OptionsWith(fn func(T) option.Uint32) SliceOf[option.Uint32] {
	return Map(ts, fn)
}

// ToUint64OptionsWith applies the provided function to each element of the slice, mapping it to a slice of uint64 options.
func (ts RawSliceOf[T]) ToUint64OptionsWith(fn func(T) option.Uint64) SliceOf[option.Uint64] {
	return Map(ts, fn)
}

// ToUintptrOptionsWith applies the provided function to each element of the slice, mapping it to a slice of uintptr options.
func (ts RawSliceOf[T]) ToUintptrOptionsWith(fn func(T) option.Uintptr) SliceOf[option.Uintptr] {
	return Map(ts, fn)
}

// ToFloat32OptionsWith applies the provided function to each element of the slice, mapping it to a slice of float32 options.
func (ts RawSliceOf[T]) ToFloat32OptionsWith(fn func(T) option.Float32) SliceOf[option.Float32] {
	return Map(ts, fn)
}

// ToFloat64OptionsWith applies the provided function to each element of the slice, mapping it to a slice of float64 options.
func (ts RawSliceOf[T]) ToFloat64OptionsWith(fn func(T) option.Float64) SliceOf[option.Float64] {
	return Map(ts, fn)
}

// ToComplex64OptionsWith applies the provided function to each element of the slice, mapping it to a slice of complex64 options.
func (ts RawSliceOf[T]) ToComplex64OptionsWith(fn func(T) option.Complex64) SliceOf[option.Complex64] {
	return Map(ts, fn)
}

// ToComplex128OptionsWith applies the provided function to each element of the slice, mapping it to a slice of complex128 options.
func (ts RawSliceOf[T]) ToComplex128OptionsWith(fn func(T) option.Complex128) SliceOf[option.Complex128] {
	return Map(ts, fn)
}

// ToRuneOptionsWith applies the provided function to each element of the slice, mapping it to a slice of rune options.
func (ts RawSliceOf[T]) ToRuneOptionsWith(fn func(T) option.Rune) SliceOf[option.Rune] {
	return Map(ts, fn)
}

// ToStringOptionsWith applies the provided function to each element of the slice, mapping it to a slice of string options.
func (ts RawSliceOf[T]) ToStringOptionsWith(fn func(T) option.String) SliceOf[option.String] {
	return Map(ts, fn)
}

// ToAnyOptionsWith applies the provided function to each element of the slice, mapping it to a slice of any options.
func (ts RawSliceOf[T]) ToAnyOptionsWith(fn func(T) option.Any) SliceOf[option.Any] {
	return Map(ts, fn)
}

// ToErrorOptionsWith applies the provided function to each element of the slice, mapping it to a slice of error options.
func (ts RawSliceOf[T]) ToErrorOptionsWith(fn func(T) option.Error) SliceOf[option.Error] {
	return Map(ts, fn)
}

// ToBoolSlicesWith applies the provided function to each element of the slice, mapping it to a slice of bool slices.
func (ts RawSliceOf[T]) ToBoolSlicesWith(fn func(T) []bool) RawSliceOf[[]bool] {
	return Map(ts, fn)
}

// ToByteSlicesWith applies the provided function to each element of the slice, mapping it to a slice of byte slices.
func (ts RawSliceOf[T]) ToByteSlicesWith(fn func(T) []byte) RawSliceOf[[]byte] {
	return Map(ts, fn)
}

// ToIntSlicesWith applies the provided function to each element of the slice, mapping it to a slice of int slices.
func (ts RawSliceOf[T]) ToIntSlicesWith(fn func(T) []int) RawSliceOf[[]int] {
	return Map(ts, fn)
}

// ToInt8SlicesWith applies the provided function to each element of the slice, mapping it to a slice of int8 slices.
func (ts RawSliceOf[T]) ToInt8SlicesWith(fn func(T) []int8) RawSliceOf[[]int8] {
	return Map(ts, fn)
}

// ToInt16SlicesWith applies the provided function to each element of the slice, mapping it to a slice of int16 slices.
func (ts RawSliceOf[T]) ToInt16SlicesWith(fn func(T) []int16) RawSliceOf[[]int16] {
	return Map(ts, fn)
}

// ToInt32SlicesWith applies the provided function to each element of the slice, mapping it to a slice of int32 slices.
func (ts RawSliceOf[T]) ToInt32SlicesWith(fn func(T) []int32) RawSliceOf[[]int32] {
	return Map(ts, fn)
}

// ToInt64SlicesWith applies the provided function to each element of the slice, mapping it to a slice of int64 slices.
func (ts RawSliceOf[T]) ToInt64SlicesWith(fn func(T) []int64) RawSliceOf[[]int64] {
	return Map(ts, fn)
}

// ToUintSlicesWith applies the provided function to each element of the slice, mapping it to a slice of uint slices.
func (ts RawSliceOf[T]) ToUintSlicesWith(fn func(T) []uint) RawSliceOf[[]uint] {
	return Map(ts, fn)
}

// ToUint8SlicesWith applies the provided function to each element of the slice, mapping it to a slice of uint8 slices.
func (ts RawSliceOf[T]) ToUint8SlicesWith(fn func(T) []uint8) RawSliceOf[[]uint8] {
	return Map(ts, fn)
}

// ToUint16SlicesWith applies the provided function to each element of the slice, mapping it to a slice of uint16 slices.
func (ts RawSliceOf[T]) ToUint16SlicesWith(fn func(T) []uint16) RawSliceOf[[]uint16] {
	return Map(ts, fn)
}

// ToUint32SlicesWith applies the provided function to each element of the slice, mapping it to a slice of uint32 slices.
func (ts RawSliceOf[T]) ToUint32SlicesWith(fn func(T) []uint32) RawSliceOf[[]uint32] {
	return Map(ts, fn)
}

// ToUint64SlicesWith applies the provided function to each element of the slice, mapping it to a slice of uint64 slices.
func (ts RawSliceOf[T]) ToUint64SlicesWith(fn func(T) []uint64) RawSliceOf[[]uint64] {
	return Map(ts, fn)
}

// ToUintptrSlicesWith applies the provided function to each element of the slice, mapping it to a slice of uintptr slices.
func (ts RawSliceOf[T]) ToUintptrSlicesWith(fn func(T) []uintptr) RawSliceOf[[]uintptr] {
	return Map(ts, fn)
}

// ToFloat32SlicesWith applies the provided function to each element of the slice, mapping it to a slice of float32 slices.
func (ts RawSliceOf[T]) ToFloat32SlicesWith(fn func(T) []float32) RawSliceOf[[]float32] {
	return Map(ts, fn)
}

// ToFloat64SlicesWith applies the provided function to each element of the slice, mapping it to a slice of float64 slices.
func (ts RawSliceOf[T]) ToFloat64SlicesWith(fn func(T) []float64) RawSliceOf[[]float64] {
	return Map(ts, fn)
}

// ToComplex64SlicesWith applies the provided function to each element of the slice, mapping it to a slice of complex64 slices.
func (ts RawSliceOf[T]) ToComplex64SlicesWith(fn func(T) []complex64) RawSliceOf[[]complex64] {
	return Map(ts, fn)
}

// ToComplex128SlicesWith applies the provided function to each element of the slice, mapping it to a slice of complex128 slices.
func (ts RawSliceOf[T]) ToComplex128SlicesWith(fn func(T) []complex128) RawSliceOf[[]complex128] {
	return Map(ts, fn)
}

// ToRuneSlicesWith applies the provided function to each element of the slice, mapping it to a slice of rune slices.
func (ts RawSliceOf[T]) ToRuneSlicesWith(fn func(T) []rune) RawSliceOf[[]rune] {
	return Map(ts, fn)
}

// ToStringSlicesWith applies the provided function to each element of the slice, mapping it to a slice of string slices.
func (ts RawSliceOf[T]) ToStringSlicesWith(fn func(T) []string) RawSliceOf[[]string] {
	return Map(ts, fn)
}

// ToAnySlicesWith applies the provided function to each element of the slice, mapping it to a slice of `any` interfaces.
func (ts RawSliceOf[T]) ToAnySlicesWith(fn func(T) []any) RawSliceOf[[]any] {
	return Map(ts, fn)
}

// ToErrorSlicesWith applies the provided function to each element of the slice, mapping it to a slice of errors.
func (ts RawSliceOf[T]) ToErrorSlicesWith(fn func(T) []error) RawSliceOf[[]error] {
	return Map(ts, fn)
}
