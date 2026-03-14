package slice

import "github.com/binaryphile/fluentfp/internal/base"

// Type aliases for types defined in internal/base.
// All methods defined on the base types are available through these aliases.
type Mapper[T any] = base.Mapper[T]
type Entries[K comparable, V any] = base.Entries[K, V]
type Float64 = base.Float64
type Int = base.Int
type String = base.String

// Convenience aliases for common Mapper instantiations.
type Any = Mapper[any]
type Bool = Mapper[bool]
type Byte = Mapper[byte]
type Error = Mapper[error]
type Float32 = Mapper[float32]
type Rune = Mapper[rune]
