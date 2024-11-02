package torunemet

import (
	"github.com/binaryphile/fluentfp/option"
)

// Boolean types
type SliceOfBoolOptions = SliceOf[option.Bool]
type SliceOfBools = SliceOf[bool]

// Byte and rune types
type SliceOfByteOptions = SliceOf[option.Byte]
type SliceOfBytes = SliceOf[byte]
type SliceOfRuneOptions = SliceOf[option.Rune]
type SliceOfRunes = SliceOf[rune]

// Complex number types
type SliceOfComplex128Options = SliceOf[option.Complex128]
type SliceOfComplex128s = SliceOf[complex128]
type SliceOfComplex64Options = SliceOf[option.Complex64]
type SliceOfComplex64s = SliceOf[complex64]

// Floating point types
type SliceOfFloat32Options = SliceOf[option.Float32]
type SliceOfFloat32s = SliceOf[float32]
type SliceOfFloat64Options = SliceOf[option.Float64]
type SliceOfFloat64s = SliceOf[float64]

// Integer types
type SliceOfInt16Options = SliceOf[option.Int16]
type SliceOfInt16s = SliceOf[int16]
type SliceOfInt32Options = SliceOf[option.Int32]
type SliceOfInt32s = SliceOf[int32]
type SliceOfInt64Options = SliceOf[option.Int64]
type SliceOfInt64s = SliceOf[int64]
type SliceOfInt8Options = SliceOf[option.Int8]
type SliceOfInt8s = SliceOf[int8]
type SliceOfIntOptions = SliceOf[option.Int]
type SliceOfInts = SliceOf[int]

// Interface types
type SliceOfAny = SliceOf[any]
type SliceOfError = SliceOf[error]

// Slice types
type SliceOfBoolSlices = RawSliceOf[[]bool]
type SliceOfByteSlices = RawSliceOf[[]byte]
type SliceOfComplex128Slices = RawSliceOf[[]complex128]
type SliceOfComplex64Slices = RawSliceOf[[]complex64]
type SliceOfFloat32Slices = RawSliceOf[[]float32]
type SliceOfFloatSlices = RawSliceOf[[]float64]
type SliceOfInt16Slices = RawSliceOf[[]int16]
type SliceOfInt32Slices = RawSliceOf[[]int32]
type SliceOfInt64Slices = RawSliceOf[[]int64]
type SliceOfInt8Slices = RawSliceOf[[]int8]
type SliceOfIntSlices = RawSliceOf[[]int]
type SliceOfRuneSlices = RawSliceOf[[]rune]
type SliceOfStringSlices = RawSliceOf[[]string]
type SliceOfUint16Slices = RawSliceOf[[]uint16]
type SliceOfUint32Slices = RawSliceOf[[]uint32]
type SliceOfUint64Slices = RawSliceOf[[]uint64]
type SliceOfUint8Slices = RawSliceOf[[]uint8]
type SliceOfUintPtrSlices = RawSliceOf[[]uintptr]
type SliceOfUintSlices = RawSliceOf[[]uint]

// String types
type SliceOfStringOptions = SliceOf[option.String]
type SliceOfStrings = SliceOf[string]

// Unsigned integer types
type SliceOfUint16Options = SliceOf[option.Uint16]
type SliceOfUint16s = SliceOf[uint16]
type SliceOfUint32Options = SliceOf[option.Uint32]
type SliceOfUint32s = SliceOf[uint32]
type SliceOfUint64Options = SliceOf[option.Uint64]
type SliceOfUint64s = SliceOf[uint64]
type SliceOfUint8Options = SliceOf[option.Uint8]
type SliceOfUint8s = SliceOf[uint8]
type SliceOfUintOptions = SliceOf[option.Uint]
type SliceOfUintPtrOptions = SliceOf[option.Uintptr]
type SliceOfUintPtrs = SliceOf[uintptr]
type SliceOfUints = SliceOf[uint]
