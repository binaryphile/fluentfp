package fluent

import "github.com/binaryphile/fluentfp/option"

// SliceWithMap derives from slice and has a type parameter R that is used solely to specify the return type of the Map method.
type SliceWithMap[T comparable, R any] []T

// Contains checks if the slice contains the specified element.
func (ts SliceWithMap[T, R]) Contains(t T) bool {
	return ts.Index(t) != -1
}

// Convert applies fn to each element of the slice, returning a new slice of the same element type with the results.
// It is the same as Map, but without changing the element type of the slice.
func (ts SliceWithMap[T, R]) Convert(fn func(T) T) SliceWithMap[T, R] {
	return Map(ts, fn)
}

// Index returns the index of the specified element in the slice, or -1 if not found.
func (ts SliceWithMap[T, R]) Index(t T) int {
	for i, v := range ts {
		if v == t {
			return i
		}
	}
	return -1
}

// KeepIf returns a new slice containing only the elements for which the provided function returns true.
func (ts SliceWithMap[T, R]) KeepIf(fn func(T) bool) SliceWithMap[T, R] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if fn(t) {
			results = append(results, t)
		}
	}
	return results
}

// Map applies the provided function to each element of the slice, mapping it to a RawSlice of type R.
func (ts SliceWithMap[T, R]) Map(fn func(T) R) RawSliceOf[R] {
	return Map(ts, fn)
}

// ToBool applies the provided function to each element of the slice, mapping it to a slice of bools.
func (ts SliceWithMap[T, R]) ToBool(fn func(T) bool) SliceWithMap[bool, R] {
	return Map(ts, fn)
}

// ToByte applies the provided function to each element of the slice, mapping it to a slice of bytes.
func (ts SliceWithMap[T, R]) ToByte(fn func(T) byte) SliceWithMap[byte, R] {
	return Map(ts, fn)
}

// ToInt applies the provided function to each element of the slice, mapping it to a slice of ints.
func (ts SliceWithMap[T, R]) ToInt(fn func(T) int) SliceWithMap[int, R] {
	return Map(ts, fn)
}

// ToInt8 applies the provided function to each element of the slice, mapping it to a slice of int8s.
func (ts SliceWithMap[T, R]) ToInt8(fn func(T) int8) SliceWithMap[int8, R] {
	return Map(ts, fn)
}

// ToInt16 applies the provided function to each element of the slice, mapping it to a slice of int16s.
func (ts SliceWithMap[T, R]) ToInt16(fn func(T) int16) SliceWithMap[int16, R] {
	return Map(ts, fn)
}

// ToInt32 applies the provided function to each element of the slice, mapping it to a slice of int32s.
func (ts SliceWithMap[T, R]) ToInt32(fn func(T) int32) SliceWithMap[int32, R] {
	return Map(ts, fn)
}

// ToInt64 applies the provided function to each element of the slice, mapping it to a slice of int64s.
func (ts SliceWithMap[T, R]) ToInt64(fn func(T) int64) SliceWithMap[int64, R] {
	return Map(ts, fn)
}

// ToUint applies the provided function to each element of the slice, mapping it to a slice of uints.
func (ts SliceWithMap[T, R]) ToUint(fn func(T) uint) SliceWithMap[uint, R] {
	return Map(ts, fn)
}

// ToUint8 applies the provided function to each element of the slice, mapping it to a slice of uint8s.
func (ts SliceWithMap[T, R]) ToUint8(fn func(T) uint8) SliceWithMap[uint8, R] {
	return Map(ts, fn)
}

// ToUint16 applies the provided function to each element of the slice, mapping it to a slice of uint16s.
func (ts SliceWithMap[T, R]) ToUint16(fn func(T) uint16) SliceWithMap[uint16, R] {
	return Map(ts, fn)
}

// ToUint32 applies the provided function to each element of the slice, mapping it to a slice of uint32s.
func (ts SliceWithMap[T, R]) ToUint32(fn func(T) uint32) SliceWithMap[uint32, R] {
	return Map(ts, fn)
}

// ToUint64 applies the provided function to each element of the slice, mapping it to a slice of uint64s.
func (ts SliceWithMap[T, R]) ToUint64(fn func(T) uint64) SliceWithMap[uint64, R] {
	return Map(ts, fn)
}

// ToUintptr applies the provided function to each element of the slice, mapping it to a slice of uintptrs.
func (ts SliceWithMap[T, R]) ToUintptr(fn func(T) uintptr) SliceWithMap[uintptr, R] {
	return Map(ts, fn)
}

// ToFloat32 applies the provided function to each element of the slice, mapping it to a slice of float32s.
func (ts SliceWithMap[T, R]) ToFloat32(fn func(T) float32) SliceWithMap[float32, R] {
	return Map(ts, fn)
}

// ToFloat64 applies the provided function to each element of the slice, mapping it to a slice of float64s.
func (ts SliceWithMap[T, R]) ToFloat64(fn func(T) float64) SliceWithMap[float64, R] {
	return Map(ts, fn)
}

// ToComplex64 applies the provided function to each element of the slice, mapping it to a slice of complex64s.
func (ts SliceWithMap[T, R]) ToComplex64(fn func(T) complex64) SliceWithMap[complex64, R] {
	return Map(ts, fn)
}

// ToComplex128 applies the provided function to each element of the slice, mapping it to a slice of complex128s.
func (ts SliceWithMap[T, R]) ToComplex128(fn func(T) complex128) SliceWithMap[complex128, R] {
	return Map(ts, fn)
}

// ToRune applies the provided function to each element of the slice, mapping it to a slice of runes.
func (ts SliceWithMap[T, R]) ToRune(fn func(T) rune) SliceWithMap[rune, R] {
	return Map(ts, fn)
}

// ToString applies the provided function to each element of the slice, mapping it to a slice of strings.
func (ts SliceWithMap[T, R]) ToString(fn func(T) string) SliceWithMap[string, R] {
	return Map(ts, fn)
}

// ToAny applies the provided function to each element of the slice, mapping it to a slice of `any` type.
func (ts SliceWithMap[T, R]) ToAny(fn func(T) any) RawSliceWithMap[any, R] {
	return Map(ts, fn)
}

// ToError applies the provided function to each element of the slice, mapping it to a slice of errors.
func (ts SliceWithMap[T, R]) ToError(fn func(T) error) RawSliceWithMap[error, R] {
	return Map(ts, fn)
}

// ToBoolOption applies the provided function to each element of the slice, mapping it to a slice of bool options.
func (ts SliceWithMap[T, R]) ToBoolOption(fn func(T) option.Bool) SliceWithMap[option.Bool, R] {
	return Map(ts, fn)
}

// ToByteOption applies the provided function to each element of the slice, mapping it to a slice of byte options.
func (ts SliceWithMap[T, R]) ToByteOption(fn func(T) option.Byte) SliceWithMap[option.Byte, R] {
	return Map(ts, fn)
}

// ToIntOption applies the provided function to each element of the slice, mapping it to a slice of int options.
func (ts SliceWithMap[T, R]) ToIntOption(fn func(T) option.Int) SliceWithMap[option.Int, R] {
	return Map(ts, fn)
}

// ToInt8Option applies the provided function to each element of the slice, mapping it to a slice of int8 options.
func (ts SliceWithMap[T, R]) ToInt8Option(fn func(T) option.Int8) SliceWithMap[option.Int8, R] {
	return Map(ts, fn)
}

// ToInt16Option applies the provided function to each element of the slice, mapping it to a slice of int16 options.
func (ts SliceWithMap[T, R]) ToInt16Option(fn func(T) option.Int16) SliceWithMap[option.Int16, R] {
	return Map(ts, fn)
}

// ToInt32Option applies the provided function to each element of the slice, mapping it to a slice of int32 options.
func (ts SliceWithMap[T, R]) ToInt32Option(fn func(T) option.Int32) SliceWithMap[option.Int32, R] {
	return Map(ts, fn)
}

// ToInt64Option applies the provided function to each element of the slice, mapping it to a slice of int64 options.
func (ts SliceWithMap[T, R]) ToInt64Option(fn func(T) option.Int64) SliceWithMap[option.Int64, R] {
	return Map(ts, fn)
}

// ToUintOption applies the provided function to each element of the slice, mapping it to a slice of uint options.
func (ts SliceWithMap[T, R]) ToUintOption(fn func(T) option.Uint) SliceWithMap[option.Uint, R] {
	return Map(ts, fn)
}

// ToUint8Option applies the provided function to each element of the slice, mapping it to a slice of uint8 options.
func (ts SliceWithMap[T, R]) ToUint8Option(fn func(T) option.Uint8) SliceWithMap[option.Uint8, R] {
	return Map(ts, fn)
}

// ToUint16Option applies the provided function to each element of the slice, mapping it to a slice of uint16 options.
func (ts SliceWithMap[T, R]) ToUint16Option(fn func(T) option.Uint16) SliceWithMap[option.Uint16, R] {
	return Map(ts, fn)
}

// ToUint32Option applies the provided function to each element of the slice, mapping it to a slice of uint32 options.
func (ts SliceWithMap[T, R]) ToUint32Option(fn func(T) option.Uint32) SliceWithMap[option.Uint32, R] {
	return Map(ts, fn)
}

// ToUint64Option applies the provided function to each element of the slice, mapping it to a slice of uint64 options.
func (ts SliceWithMap[T, R]) ToUint64Option(fn func(T) option.Uint64) SliceWithMap[option.Uint64, R] {
	return Map(ts, fn)
}

// ToUintptrOption applies the provided function to each element of the slice, mapping it to a slice of uintptr options.
func (ts SliceWithMap[T, R]) ToUintptrOption(fn func(T) option.Uintptr) SliceWithMap[option.Uintptr, R] {
	return Map(ts, fn)
}

// ToFloat32Option applies the provided function to each element of the slice, mapping it to a slice of float32 options.
func (ts SliceWithMap[T, R]) ToFloat32Option(fn func(T) option.Float32) SliceWithMap[option.Float32, R] {
	return Map(ts, fn)
}

// ToFloat64Option applies the provided function to each element of the slice, mapping it to a slice of float64 options.
func (ts SliceWithMap[T, R]) ToFloat64Option(fn func(T) option.Float64) SliceWithMap[option.Float64, R] {
	return Map(ts, fn)
}

// ToComplex64Option applies the provided function to each element of the slice, mapping it to a slice of complex64 options.
func (ts SliceWithMap[T, R]) ToComplex64Option(fn func(T) option.Complex64) SliceWithMap[option.Complex64, R] {
	return Map(ts, fn)
}

// ToComplex128Option applies the provided function to each element of the slice, mapping it to a slice of complex128 options.
func (ts SliceWithMap[T, R]) ToComplex128Option(fn func(T) option.Complex128) SliceWithMap[option.Complex128, R] {
	return Map(ts, fn)
}

// ToRuneOption applies the provided function to each element of the slice, mapping it to a slice of rune options.
func (ts SliceWithMap[T, R]) ToRuneOption(fn func(T) option.Rune) SliceWithMap[option.Rune, R] {
	return Map(ts, fn)
}

// ToStringOption applies the provided function to each element of the slice, mapping it to a slice of string options.
func (ts SliceWithMap[T, R]) ToStringOption(fn func(T) option.String) SliceWithMap[option.String, R] {
	return Map(ts, fn)
}

// ToAnyOption applies the provided function to each element of the slice, mapping it to a slice of any options.
func (ts SliceWithMap[T, R]) ToAnyOption(fn func(T) option.Any) SliceWithMap[option.Any, R] {
	return Map(ts, fn)
}

// ToErrorOption applies the provided function to each element of the slice, mapping it to a slice of error options.
func (ts SliceWithMap[T, R]) ToErrorOption(fn func(T) option.Error) SliceWithMap[option.Error, R] {
	return Map(ts, fn)
}

// ToStringSlice applies the provided function to each element of the slice, mapping it to a slice of string slices.
func (ts SliceWithMap[T, R]) ToStringSlice(fn func(T) []string) RawSliceWithMap[[]string, R] {
	return Map(ts, fn)
}

// ToBoolSlice applies the provided function to each element of the slice, mapping it to a slice of bool slices.
func (ts SliceWithMap[T, R]) ToBoolSlice(fn func(T) []bool) RawSliceWithMap[[]bool, R] {
	return Map(ts, fn)
}

// ToIntSlice applies the provided function to each element of the slice, mapping it to a slice of int slices.
func (ts SliceWithMap[T, R]) ToIntSlice(fn func(T) []int) RawSliceWithMap[[]int, R] {
	return Map(ts, fn)
}

// ToInt8Slice applies the provided function to each element of the slice, mapping it to a slice of int8 slices.
func (ts SliceWithMap[T, R]) ToInt8Slice(fn func(T) []int8) RawSliceWithMap[[]int8, R] {
	return Map(ts, fn)
}

// ToInt16Slice applies the provided function to each element of the slice, mapping it to a slice of int16 slices.
func (ts SliceWithMap[T, R]) ToInt16Slice(fn func(T) []int16) RawSliceWithMap[[]int16, R] {
	return Map(ts, fn)
}

// ToInt32Slice applies the provided function to each element of the slice, mapping it to a slice of int32 slices.
func (ts SliceWithMap[T, R]) ToInt32Slice(fn func(T) []int32) RawSliceWithMap[[]int32, R] {
	return Map(ts, fn)
}

// ToInt64Slice applies the provided function to each element of the slice, mapping it to a slice of int64 slices.
func (ts SliceWithMap[T, R]) ToInt64Slice(fn func(T) []int64) RawSliceWithMap[[]int64, R] {
	return Map(ts, fn)
}

// ToUintSlice applies the provided function to each element of the slice, mapping it to a slice of uint slices.
func (ts SliceWithMap[T, R]) ToUintSlice(fn func(T) []uint) RawSliceWithMap[[]uint, R] {
	return Map(ts, fn)
}

// ToUint8Slice applies the provided function to each element of the slice, mapping it to a slice of uint8 slices.
func (ts SliceWithMap[T, R]) ToUint8Slice(fn func(T) []uint8) RawSliceWithMap[[]uint8, R] {
	return Map(ts, fn)
}

// ToUint16Slice applies the provided function to each element of the slice, mapping it to a slice of uint16 slices.
func (ts SliceWithMap[T, R]) ToUint16Slice(fn func(T) []uint16) RawSliceWithMap[[]uint16, R] {
	return Map(ts, fn)
}

// ToUint32Slice applies the provided function to each element of the slice, mapping it to a slice of uint32 slices.
func (ts SliceWithMap[T, R]) ToUint32Slice(fn func(T) []uint32) RawSliceWithMap[[]uint32, R] {
	return Map(ts, fn)
}

// ToUint64Slice applies the provided function to each element of the slice, mapping it to a slice of uint64 slices.
func (ts SliceWithMap[T, R]) ToUint64Slice(fn func(T) []uint64) RawSliceWithMap[[]uint64, R] {
	return Map(ts, fn)
}

// ToUintptrSlice applies the provided function to each element of the slice, mapping it to a slice of uintptr slices.
func (ts SliceWithMap[T, R]) ToUintptrSlice(fn func(T) []uintptr) RawSliceWithMap[[]uintptr, R] {
	return Map(ts, fn)
}

// ToFloat32Slice applies the provided function to each element of the slice, mapping it to a slice of float32 slices.
func (ts SliceWithMap[T, R]) ToFloat32Slice(fn func(T) []float32) RawSliceWithMap[[]float32, R] {
	return Map(ts, fn)
}

// ToFloat64Slice applies the provided function to each element of the slice, mapping it to a slice of float64 slices.
func (ts SliceWithMap[T, R]) ToFloat64Slice(fn func(T) []float64) RawSliceWithMap[[]float64, R] {
	return Map(ts, fn)
}

// ToComplex64Slice applies the provided function to each element of the slice, mapping it to a slice of complex64 slices.
func (ts SliceWithMap[T, R]) ToComplex64Slice(fn func(T) []complex64) RawSliceWithMap[[]complex64, R] {
	return Map(ts, fn)
}

// ToComplex128Slice applies the provided function to each element of the slice, mapping it to a slice of complex128 slices.
func (ts SliceWithMap[T, R]) ToComplex128Slice(fn func(T) []complex128) RawSliceWithMap[[]complex128, R] {
	return Map(ts, fn)
}

// ToByteSlice applies the provided function to each element of the slice, mapping it to a slice of byte slices.
func (ts SliceWithMap[T, R]) ToByteSlice(fn func(T) []byte) RawSliceWithMap[[]byte, R] {
	return Map(ts, fn)
}

// ToRuneSlice applies the provided function to each element of the slice, mapping it to a slice of rune slices.
func (ts SliceWithMap[T, R]) ToRuneSlice(fn func(T) []rune) RawSliceWithMap[[]rune, R] {
	return Map(ts, fn)
}

// RemoveIf returns a new slice containing only the elements for which the provided function returns false.
func (ts SliceWithMap[T, R]) RemoveIf(fn func(T) bool) SliceWithMap[T, R] {
	results := make([]T, 0, len(ts))
	for _, t := range ts {
		if !fn(t) {
			results = append(results, t)
		}
	}
	return results
}

// TakeFirst returns the first n elements of ts.
func (ts SliceWithMap[T, R]) TakeFirst(n int) SliceWithMap[T, R] {
	if n > len(ts) {
		n = len(ts)
	}

	return ts[:n]
}
