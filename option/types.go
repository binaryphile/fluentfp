package option

type Bool = Basic[bool]

var (
	BoolOf    = BasicOf[bool]
	NewBool   = NewBasic[bool]
	NotOkBool = Bool{}
)

type String = Basic[string]

var (
	NewString   = NewBasic[string]
	NotOkString = String{}
	StringOf    = BasicOf[string]
)

// Integer types

type Int = Basic[int]
type Int8 = Basic[int8]
type Int16 = Basic[int16]
type Int32 = Basic[int32]
type Int64 = Basic[int64]

var (
	IntOf    = BasicOf[int]
	NewInt   = NewBasic[int]
	NotOkInt = Int{}

	Int8Of    = BasicOf[int8]
	NewInt8   = NewBasic[int8]
	NotOkInt8 = Int8{}

	Int16Of    = BasicOf[int16]
	NewInt16   = NewBasic[int16]
	NotOkInt16 = Int16{}

	Int32Of    = BasicOf[int32]
	NewInt32   = NewBasic[int32]
	NotOkInt32 = Int32{}

	Int64Of    = BasicOf[int64]
	NewInt64   = NewBasic[int64]
	NotOkInt64 = Int64{}
)

// Unsigned integer types

type Uint = Basic[uint]
type Uint8 = Basic[uint8]
type Uint16 = Basic[uint16]
type Uint32 = Basic[uint32]
type Uint64 = Basic[uint64]
type Uintptr = Basic[uintptr]

var (
	UintOf    = BasicOf[uint]
	NewUint   = NewBasic[uint]
	NotOkUint = Uint{}

	Uint8Of    = BasicOf[uint8]
	NewUint8   = NewBasic[uint8]
	NotOkUint8 = Uint8{}

	Uint16Of    = BasicOf[uint16]
	NewUint16   = NewBasic[uint16]
	NotOkUint16 = Uint16{}

	Uint32Of    = BasicOf[uint32]
	NewUint32   = NewBasic[uint32]
	NotOkUint32 = Uint32{}

	Uint64Of    = BasicOf[uint64]
	NewUint64   = NewBasic[uint64]
	NotOkUint64 = Uint64{}

	UintptrOf    = BasicOf[uintptr]
	NewUintptr   = NewBasic[uintptr]
	NotOkUintptr = Uintptr{}
)

// Floating point types

type Float32 = Basic[float32]
type Float64 = Basic[float64]

var (
	Float32Of    = BasicOf[float32]
	NewFloat32   = NewBasic[float32]
	NotOkFloat32 = Float32{}

	Float64Of    = BasicOf[float64]
	NewFloat64   = NewBasic[float64]
	NotOkFloat64 = Float64{}
)

// Complex number types

type Complex64 = Basic[complex64]
type Complex128 = Basic[complex128]

var (
	Complex64Of    = BasicOf[complex64]
	NewComplex64   = NewBasic[complex64]
	NotOkComplex64 = Complex64{}

	Complex128Of    = BasicOf[complex128]
	NewComplex128   = NewBasic[complex128]
	NotOkComplex128 = Complex128{}
)

// Byte and rune types

type Byte = Basic[byte]
type Rune = Basic[rune]

var (
	ByteOf    = BasicOf[byte]
	NewByte   = NewBasic[byte]
	NotOkByte = Byte{}

	RuneOf    = BasicOf[rune]
	NewRune   = NewBasic[rune]
	NotOkRune = Rune{}
)
