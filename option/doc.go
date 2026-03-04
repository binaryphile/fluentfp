// Package option provides types and functions to work with optional values.
package option

func _() {
	_ = Basic[int].Get
	_ = Basic[int].IsOk
	_ = Basic[int].MustGet
	_ = Basic[int].Or
	_ = Basic[int].OrCall
	_ = Basic[int].OrEmpty
	_ = Basic[int].OrFalse
	_ = Basic[int].OrZero
	_ = Basic[int].ToOpt

	type _ = Any
	type _ = Bool
	type _ = Byte
	type _ = Error
	type _ = Int
	type _ = Rune
	type _ = String

	_ = Getenv("")
	_ = IfNonEmpty
	_ = IfNonNil[int]
	_ = IfNonZero[int]
	_ = Map(Basic[int]{}, func(int) int { return 0 })
	_ = NonEmptyMap("", func(string) int { return 0 })
	_ = NonNilMap[int, int](nil, func(int) int { return 0 })
	_ = NonZeroMap[int, int](0, func(int) int { return 0 })
	_ = New[int]
	_ = Of[int]

	_ = NotOkAny
	_ = NotOkBool
	_ = NotOkByte
	_ = NotOkError
	_ = NotOkInt
	_ = NotOkRune
	_ = NotOkString
}
