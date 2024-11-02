package option

// Other types

type Bool = Basic[bool]
type Byte = Basic[byte]
type Rune = Basic[rune]
type String = Basic[string]

// Interface types

type Any = Basic[any]
type Error = Basic[error]

// Integer types

type Int = Basic[int]
type Int8 = Basic[int8]
type Int16 = Basic[int16]
type Int32 = Basic[int32]
type Int64 = Basic[int64]

// Unsigned integer types

type Uint = Basic[uint]
type Uint8 = Basic[uint8]
type Uint16 = Basic[uint16]
type Uint32 = Basic[uint32]
type Uint64 = Basic[uint64]
type Uintptr = Basic[uintptr]

// Floating point types

type Float32 = Basic[float32]
type Float64 = Basic[float64]

// Complex number types

type Complex64 = Basic[complex64]
type Complex128 = Basic[complex128]
