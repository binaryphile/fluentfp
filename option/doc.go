//go:build ignore
// +build ignore

package option

func _() {
	_ = (Basic[bool]{}).Get
	_ = (Basic[bool]{}).IsOk
	_ = (Basic[bool]{}).MustGet
	_ = (Basic[bool]{}).Or
	_ = (Basic[bool]{}).OrCall
	_ = (Basic[bool]{}).OrEmpty
	_ = (Basic[bool]{}).OrFalse
	_ = (Basic[bool]{}).OrZero
	_ = (Basic[bool]{}).ToPointer
	_ = Getenv("")
	_ = New[bool]
	_ = OfPointee[bool]
	_ = IfProvided[bool]
	_ = Of[bool]
}

var (
	_ = AnyOf
	_ = NewAny
	_ = NotOkAny
	_ = AnyIfProvided

	_ = BoolOf
	_ = NewBool
	_ = NotOkBool
	_ = BoolIfProvided

	_ = NewString
	_ = NotOkString
	_ = StringOf
	_ = StringIfProvided

	_ = IntOf
	_ = NewInt
	_ = NotOkInt
	_ = IntIfProvided

	_ = Int8Of
	_ = NewInt8
	_ = NotOkInt8
	_ = Int8IfProvided

	_ = Int16Of
	_ = NewInt16
	_ = NotOkInt16
	_ = Int16IfProvided

	_ = Int32Of
	_ = NewInt32
	_ = NotOkInt32
	_ = Int32IfProvided

	_ = Int64Of
	_ = NewInt64
	_ = NotOkInt64
	_ = Int64IfProvided

	_ = UintOf
	_ = NewUint
	_ = NotOkUint
	_ = UintIfProvided

	_ = Uint8Of
	_ = NewUint8
	_ = NotOkUint8
	_ = Uint8IfProvided

	_ = Uint16Of
	_ = NewUint16
	_ = NotOkUint16
	_ = Uint16IfProvided

	_ = Uint32Of
	_ = NewUint32
	_ = NotOkUint32
	_ = Uint32IfProvided

	_ = Uint64Of
	_ = NewUint64
	_ = NotOkUint64
	_ = Uint64IfProvided

	_ = UintptrOf
	_ = NewUintptr
	_ = NotOkUintptr
	_ = UintptrIfProvided

	_ = Float32Of
	_ = NewFloat32
	_ = NotOkFloat32
	_ = Float32IfProvided

	_ = Float64Of
	_ = NewFloat64
	_ = NotOkFloat64
	_ = Float64IfProvided

	_ = Complex64Of
	_ = NewComplex64
	_ = NotOkComplex64
	_ = Complex64IfProvided

	_ = Complex128Of
	_ = NewComplex128
	_ = NotOkComplex128
	_ = Complex128IfProvided

	_ = ByteOf
	_ = NewByte
	_ = NotOkByte
	_ = ByteIfProvided

	_ = RuneOf
	_ = NewRune
	_ = NotOkRune
	_ = RuneIfProvided

	_ = ErrorOf
	_ = NewError
	_ = NotOkError
	_ = ErrorIfProvided
)
