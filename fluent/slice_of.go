package fluent

import "github.com/binaryphile/fluentfp/option"

// SliceOf derives from slice.
// It is usable anywhere a slice is, but provides additional fluent fp methods.
type SliceOf[T comparable] []T

// Contains returns whether the slice contains the specified element.
func (ts SliceOf[T]) Contains(t T) bool {
	return ts.IndexOf(t) != -1
}

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

// IndexOf returns the index of the specified element in the slice, or -1 if not found.
func (ts SliceOf[T]) IndexOf(t T) int {
	for i, v := range ts {
		if t == v {
			return i
		}
	}

	return -1
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

// ToBoolWith applies the provided function to each element of the slice, mapping it to a slice of bools.
func (ts SliceOf[T]) ToBoolWith(fn func(T) bool) SliceOf[bool] {
	return Map(ts, fn)
}

// ToByteWith applies the provided function to each element of the slice, mapping it to a slice of bytes.
func (ts SliceOf[T]) ToByteWith(fn func(T) byte) SliceOf[byte] {
	return Map(ts, fn)
}

// ToIntWith applies the provided function to each element of the slice, mapping it to a slice of ints.
func (ts SliceOf[T]) ToIntWith(fn func(T) int) SliceOf[int] {
	return Map(ts, fn)
}

// ToInt8With applies the provided function to each element of the slice, mapping it to a slice of int8s.
func (ts SliceOf[T]) ToInt8With(fn func(T) int8) SliceOf[int8] {
	return Map(ts, fn)
}

// ToInt16With applies the provided function to each element of the slice, mapping it to a slice of int16s.
func (ts SliceOf[T]) ToInt16With(fn func(T) int16) SliceOf[int16] {
	return Map(ts, fn)
}

// ToInt32With applies the provided function to each element of the slice, mapping it to a slice of int32s.
func (ts SliceOf[T]) ToInt32With(fn func(T) int32) SliceOf[int32] {
	return Map(ts, fn)
}

// ToInt64With applies the provided function to each element of the slice, mapping it to a slice of int64s.
func (ts SliceOf[T]) ToInt64With(fn func(T) int64) SliceOf[int64] {
	return Map(ts, fn)
}

// ToUintWith applies the provided function to each element of the slice, mapping it to a slice of uints.
func (ts SliceOf[T]) ToUintWith(fn func(T) uint) SliceOf[uint] {
	return Map(ts, fn)
}

// ToUint8With applies the provided function to each element of the slice, mapping it to a slice of uint8s.
func (ts SliceOf[T]) ToUint8With(fn func(T) uint8) SliceOf[uint8] {
	return Map(ts, fn)
}

// ToUint16With applies the provided function to each element of the slice, mapping it to a slice of uint16s.
func (ts SliceOf[T]) ToUint16With(fn func(T) uint16) SliceOf[uint16] {
	return Map(ts, fn)
}

// ToUint32With applies the provided function to each element of the slice, mapping it to a slice of uint32s.
func (ts SliceOf[T]) ToUint32With(fn func(T) uint32) SliceOf[uint32] {
	return Map(ts, fn)
}

// ToUint64With applies the provided function to each element of the slice, mapping it to a slice of uint64s.
func (ts SliceOf[T]) ToUint64With(fn func(T) uint64) SliceOf[uint64] {
	return Map(ts, fn)
}

// ToUintptrWith applies the provided function to each element of the slice, mapping it to a slice of uintptrs.
func (ts SliceOf[T]) ToUintptrWith(fn func(T) uintptr) SliceOf[uintptr] {
	return Map(ts, fn)
}

// ToFloat32With applies the provided function to each element of the slice, mapping it to a slice of float32s.
func (ts SliceOf[T]) ToFloat32With(fn func(T) float32) SliceOf[float32] {
	return Map(ts, fn)
}

// ToFloat64With applies the provided function to each element of the slice, mapping it to a slice of float64s.
func (ts SliceOf[T]) ToFloat64With(fn func(T) float64) SliceOf[float64] {
	return Map(ts, fn)
}

// ToComplex64With applies the provided function to each element of the slice, mapping it to a slice of complex64s.
func (ts SliceOf[T]) ToComplex64With(fn func(T) complex64) SliceOf[complex64] {
	return Map(ts, fn)
}

// ToComplex128With applies the provided function to each element of the slice, mapping it to a slice of complex128s.
func (ts SliceOf[T]) ToComplex128With(fn func(T) complex128) SliceOf[complex128] {
	return Map(ts, fn)
}

// ToRuneWith applies the provided function to each element of the slice, mapping it to a slice of runes.
func (ts SliceOf[T]) ToRuneWith(fn func(T) rune) SliceOf[rune] {
	return Map(ts, fn)
}

// ToStringWith applies the provided function to each element of the slice, mapping it to a slice of strings.
func (ts SliceOf[T]) ToStringWith(fn func(T) string) SliceOf[string] {
	return Map(ts, fn)
}

// ToAnyWith applies the provided function to each element of the slice, mapping it to a slice of `any` type.
func (ts SliceOf[T]) ToAnyWith(fn func(T) any) RawSliceOf[any] {
	return Map(ts, fn)
}

// ToErrorWith applies the provided function to each element of the slice, mapping it to a slice of errors.
func (ts SliceOf[T]) ToErrorWith(fn func(T) error) RawSliceOf[error] {
	return Map(ts, fn)
}

// ToBoolOptionWith applies the provided function to each element of the slice, mapping it to a slice of bool options.
func (ts SliceOf[T]) ToBoolOptionWith(fn func(T) option.Bool) SliceOf[option.Bool] {
	return Map(ts, fn)
}

// ToByteOptionWith applies the provided function to each element of the slice, mapping it to a slice of byte options.
func (ts SliceOf[T]) ToByteOptionWith(fn func(T) option.Byte) SliceOf[option.Byte] {
	return Map(ts, fn)
}

// ToIntOptionWith applies the provided function to each element of the slice, mapping it to a slice of int options.
func (ts SliceOf[T]) ToIntOptionWith(fn func(T) option.Int) SliceOf[option.Int] {
	return Map(ts, fn)
}

// ToInt8OptionWith applies the provided function to each element of the slice, mapping it to a slice of int8 options.
func (ts SliceOf[T]) ToInt8OptionWith(fn func(T) option.Int8) SliceOf[option.Int8] {
	return Map(ts, fn)
}

// ToInt16OptionWith applies the provided function to each element of the slice, mapping it to a slice of int16 options.
func (ts SliceOf[T]) ToInt16OptionWith(fn func(T) option.Int16) SliceOf[option.Int16] {
	return Map(ts, fn)
}

// ToInt32OptionWith applies the provided function to each element of the slice, mapping it to a slice of int32 options.
func (ts SliceOf[T]) ToInt32OptionWith(fn func(T) option.Int32) SliceOf[option.Int32] {
	return Map(ts, fn)
}

// ToInt64OptionWith applies the provided function to each element of the slice, mapping it to a slice of int64 options.
func (ts SliceOf[T]) ToInt64OptionWith(fn func(T) option.Int64) SliceOf[option.Int64] {
	return Map(ts, fn)
}

// ToUintOptionWith applies the provided function to each element of the slice, mapping it to a slice of uint options.
func (ts SliceOf[T]) ToUintOptionWith(fn func(T) option.Uint) SliceOf[option.Uint] {
	return Map(ts, fn)
}

// ToUint8OptionWith applies the provided function to each element of the slice, mapping it to a slice of uint8 options.
func (ts SliceOf[T]) ToUint8OptionWith(fn func(T) option.Uint8) SliceOf[option.Uint8] {
	return Map(ts, fn)
}

// ToUint16OptionWith applies the provided function to each element of the slice, mapping it to a slice of uint16 options.
func (ts SliceOf[T]) ToUint16OptionWith(fn func(T) option.Uint16) SliceOf[option.Uint16] {
	return Map(ts, fn)
}

// ToUint32OptionWith applies the provided function to each element of the slice, mapping it to a slice of uint32 options.
func (ts SliceOf[T]) ToUint32OptionWith(fn func(T) option.Uint32) SliceOf[option.Uint32] {
	return Map(ts, fn)
}

// ToUint64OptionWith applies the provided function to each element of the slice, mapping it to a slice of uint64 options.
func (ts SliceOf[T]) ToUint64OptionWith(fn func(T) option.Uint64) SliceOf[option.Uint64] {
	return Map(ts, fn)
}

// ToUintptrOptionWith applies the provided function to each element of the slice, mapping it to a slice of uintptr options.
func (ts SliceOf[T]) ToUintptrOptionWith(fn func(T) option.Uintptr) SliceOf[option.Uintptr] {
	return Map(ts, fn)
}

// ToFloat32OptionWith applies the provided function to each element of the slice, mapping it to a slice of float32 options.
func (ts SliceOf[T]) ToFloat32OptionWith(fn func(T) option.Float32) SliceOf[option.Float32] {
	return Map(ts, fn)
}

// ToFloat64OptionWith applies the provided function to each element of the slice, mapping it to a slice of float64 options.
func (ts SliceOf[T]) ToFloat64OptionWith(fn func(T) option.Float64) SliceOf[option.Float64] {
	return Map(ts, fn)
}

// ToComplex64OptionWith applies the provided function to each element of the slice, mapping it to a slice of complex64 options.
func (ts SliceOf[T]) ToComplex64OptionWith(fn func(T) option.Complex64) SliceOf[option.Complex64] {
	return Map(ts, fn)
}

// ToComplex128OptionWith applies the provided function to each element of the slice, mapping it to a slice of complex128 options.
func (ts SliceOf[T]) ToComplex128OptionWith(fn func(T) option.Complex128) SliceOf[option.Complex128] {
	return Map(ts, fn)
}

// ToRuneOptionWith applies the provided function to each element of the slice, mapping it to a slice of rune options.
func (ts SliceOf[T]) ToRuneOptionWith(fn func(T) option.Rune) SliceOf[option.Rune] {
	return Map(ts, fn)
}

// ToStringOptionWith applies the provided function to each element of the slice, mapping it to a slice of string options.
func (ts SliceOf[T]) ToStringOptionWith(fn func(T) option.String) SliceOf[option.String] {
	return Map(ts, fn)
}

// ToAnyOptionWith applies the provided function to each element of the slice, mapping it to a slice of any options.
func (ts SliceOf[T]) ToAnyOptionWith(fn func(T) option.Any) SliceOf[option.Any] {
	return Map(ts, fn)
}

// ToErrorOptionWith applies the provided function to each element of the slice, mapping it to a slice of error options.
func (ts SliceOf[T]) ToErrorOptionWith(fn func(T) option.Error) SliceOf[option.Error] {
	return Map(ts, fn)
}

// ToBoolSliceWith applies the provided function to each element of the slice, mapping it to a slice of bool slices.
func (ts SliceOf[T]) ToBoolSliceWith(fn func(T) []bool) RawSliceOf[[]bool] {
	return Map(ts, fn)
}

// ToByteSliceWith applies the provided function to each element of the slice, mapping it to a slice of byte slices.
func (ts SliceOf[T]) ToByteSliceWith(fn func(T) []byte) RawSliceOf[[]byte] {
	return Map(ts, fn)
}

// ToIntSliceWith applies the provided function to each element of the slice, mapping it to a slice of int slices.
func (ts SliceOf[T]) ToIntSliceWith(fn func(T) []int) RawSliceOf[[]int] {
	return Map(ts, fn)
}

// ToInt8SliceWith applies the provided function to each element of the slice, mapping it to a slice of int8 slices.
func (ts SliceOf[T]) ToInt8SliceWith(fn func(T) []int8) RawSliceOf[[]int8] {
	return Map(ts, fn)
}

// ToInt16SliceWith applies the provided function to each element of the slice, mapping it to a slice of int16 slices.
func (ts SliceOf[T]) ToInt16SliceWith(fn func(T) []int16) RawSliceOf[[]int16] {
	return Map(ts, fn)
}

// ToInt32SliceWith applies the provided function to each element of the slice, mapping it to a slice of int32 slices.
func (ts SliceOf[T]) ToInt32SliceWith(fn func(T) []int32) RawSliceOf[[]int32] {
	return Map(ts, fn)
}

// ToInt64SliceWith applies the provided function to each element of the slice, mapping it to a slice of int64 slices.
func (ts SliceOf[T]) ToInt64SliceWith(fn func(T) []int64) RawSliceOf[[]int64] {
	return Map(ts, fn)
}

// ToUintSliceWith applies the provided function to each element of the slice, mapping it to a slice of uint slices.
func (ts SliceOf[T]) ToUintSliceWith(fn func(T) []uint) RawSliceOf[[]uint] {
	return Map(ts, fn)
}

// ToUint8SliceWith applies the provided function to each element of the slice, mapping it to a slice of uint8 slices.
func (ts SliceOf[T]) ToUint8SliceWith(fn func(T) []uint8) RawSliceOf[[]uint8] {
	return Map(ts, fn)
}

// ToUint16SliceWith applies the provided function to each element of the slice, mapping it to a slice of uint16 slices.
func (ts SliceOf[T]) ToUint16SliceWith(fn func(T) []uint16) RawSliceOf[[]uint16] {
	return Map(ts, fn)
}

// ToUint32SliceWith applies the provided function to each element of the slice, mapping it to a slice of uint32 slices.
func (ts SliceOf[T]) ToUint32SliceWith(fn func(T) []uint32) RawSliceOf[[]uint32] {
	return Map(ts, fn)
}

// ToUint64SliceWith applies the provided function to each element of the slice, mapping it to a slice of uint64 slices.
func (ts SliceOf[T]) ToUint64SliceWith(fn func(T) []uint64) RawSliceOf[[]uint64] {
	return Map(ts, fn)
}

// ToUintptrSliceWith applies the provided function to each element of the slice, mapping it to a slice of uintptr slices.
func (ts SliceOf[T]) ToUintptrSliceWith(fn func(T) []uintptr) RawSliceOf[[]uintptr] {
	return Map(ts, fn)
}

// ToFloat32SliceWith applies the provided function to each element of the slice, mapping it to a slice of float32 slices.
func (ts SliceOf[T]) ToFloat32SliceWith(fn func(T) []float32) RawSliceOf[[]float32] {
	return Map(ts, fn)
}

// ToFloat64SliceWith applies the provided function to each element of the slice, mapping it to a slice of float64 slices.
func (ts SliceOf[T]) ToFloat64SliceWith(fn func(T) []float64) RawSliceOf[[]float64] {
	return Map(ts, fn)
}

// ToComplex64SliceWith applies the provided function to each element of the slice, mapping it to a slice of complex64 slices.
func (ts SliceOf[T]) ToComplex64SliceWith(fn func(T) []complex64) RawSliceOf[[]complex64] {
	return Map(ts, fn)
}

// ToComplex128SliceWith applies the provided function to each element of the slice, mapping it to a slice of complex128 slices.
func (ts SliceOf[T]) ToComplex128SliceWith(fn func(T) []complex128) RawSliceOf[[]complex128] {
	return Map(ts, fn)
}

// ToRuneSliceWith applies the provided function to each element of the slice, mapping it to a slice of rune slices.
func (ts SliceOf[T]) ToRuneSliceWith(fn func(T) []rune) RawSliceOf[[]rune] {
	return Map(ts, fn)
}

// ToStringSliceWith applies the provided function to each element of the slice, mapping it to a slice of string slices.
func (ts SliceOf[T]) ToStringSliceWith(fn func(T) []string) RawSliceOf[[]string] {
	return Map(ts, fn)
}

// ToAnySliceWith applies the provided function to each element of the slice, mapping it to a slice of `any` interfaces.
func (ts SliceOf[T]) ToAnySliceWith(fn func(T) []any) RawSliceOf[[]any] {
	return Map(ts, fn)
}

// ToErrorSliceWith applies the provided function to each element of the slice, mapping it to a slice of errors.
func (ts SliceOf[T]) ToErrorSliceWith(fn func(T) []error) RawSliceOf[[]error] {
	return Map(ts, fn)
}
