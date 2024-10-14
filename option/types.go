package option

type Any = Basic[any]

var (
	AnyOf         = Of[any]
	NewAny        = NewBasic[any]
	NotOkAny      = Any{}
	AnyIfProvided = IfProvided[any]
)

type Bool = Basic[bool]

var (
	BoolOf         = Of[bool]
	NewBool        = NewBasic[bool]
	NotOkBool      = Bool{}
	BoolIfProvided = IfProvided[bool]
)

type String = Basic[string]

var (
	NewString        = NewBasic[string]
	NotOkString      = String{}
	StringOf         = Of[string]
	StringIfProvided = IfProvided[string]
)

// Integer types

type Int = Basic[int]
type Int8 = Basic[int8]
type Int16 = Basic[int16]
type Int32 = Basic[int32]
type Int64 = Basic[int64]

var (
	IntOf         = Of[int]
	NewInt        = NewBasic[int]
	NotOkInt      = Int{}
	IntIfProvided = IfProvided[int]

	Int8Of         = Of[int8]
	NewInt8        = NewBasic[int8]
	NotOkInt8      = Int8{}
	Int8IfProvided = IfProvided[int8]

	Int16Of         = Of[int16]
	NewInt16        = NewBasic[int16]
	NotOkInt16      = Int16{}
	Int16IfProvided = IfProvided[int16]

	Int32Of         = Of[int32]
	NewInt32        = NewBasic[int32]
	NotOkInt32      = Int32{}
	Int32IfProvided = IfProvided[int32]

	Int64Of         = Of[int64]
	NewInt64        = NewBasic[int64]
	NotOkInt64      = Int64{}
	Int64IfProvided = IfProvided[int64]
)

// Unsigned integer types

type Uint = Basic[uint]
type Uint8 = Basic[uint8]
type Uint16 = Basic[uint16]
type Uint32 = Basic[uint32]
type Uint64 = Basic[uint64]
type Uintptr = Basic[uintptr]

var (
	UintOf         = Of[uint]
	NewUint        = NewBasic[uint]
	NotOkUint      = Uint{}
	UintIfProvided = IfProvided[uint]

	Uint8Of         = Of[uint8]
	NewUint8        = NewBasic[uint8]
	NotOkUint8      = Uint8{}
	Uint8IfProvided = IfProvided[uint8]

	Uint16Of         = Of[uint16]
	NewUint16        = NewBasic[uint16]
	NotOkUint16      = Uint16{}
	Uint16IfProvided = IfProvided[uint16]

	Uint32Of         = Of[uint32]
	NewUint32        = NewBasic[uint32]
	NotOkUint32      = Uint32{}
	Uint32IfProvided = IfProvided[uint32]

	Uint64Of         = Of[uint64]
	NewUint64        = NewBasic[uint64]
	NotOkUint64      = Uint64{}
	Uint64IfProvided = IfProvided[uint64]

	UintptrOf         = Of[uintptr]
	NewUintptr        = NewBasic[uintptr]
	NotOkUintptr      = Uintptr{}
	UintptrIfProvided = IfProvided[uintptr]
)

// Floating point types

type Float32 = Basic[float32]
type Float64 = Basic[float64]

var (
	Float32Of         = Of[float32]
	NewFloat32        = NewBasic[float32]
	NotOkFloat32      = Float32{}
	Float32IfProvided = IfProvided[float32]

	Float64Of         = Of[float64]
	NewFloat64        = NewBasic[float64]
	NotOkFloat64      = Float64{}
	Float64IfProvided = IfProvided[float64]
)

// Complex number types

type Complex64 = Basic[complex64]
type Complex128 = Basic[complex128]

var (
	Complex64Of         = Of[complex64]
	NewComplex64        = NewBasic[complex64]
	NotOkComplex64      = Complex64{}
	Complex64IfProvided = IfProvided[complex64]

	Complex128Of         = Of[complex128]
	NewComplex128        = NewBasic[complex128]
	NotOkComplex128      = Complex128{}
	Complex128IfProvided = IfProvided[complex128]
)

// Byte and rune types

type Byte = Basic[byte]
type Rune = Basic[rune]

var (
	ByteOf         = Of[byte]
	NewByte        = NewBasic[byte]
	NotOkByte      = Byte{}
	ByteIfProvided = IfProvided[byte]

	RuneOf         = Of[rune]
	NewRune        = NewBasic[rune]
	NotOkRune      = Rune{}
	RuneIfProvided = IfProvided[rune]
)

// Error type

type Error = Basic[error]

var (
	ErrorOf         = Of[error]
	NewError        = NewBasic[error]
	NotOkError      = Error{}
	ErrorIfProvided = IfProvided[error]
)
