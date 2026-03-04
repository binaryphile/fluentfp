// Package option provides types and functions to work with optional values.
package option

func _() {
	_ = Option[int].Get
	_ = Option[int].IsOk
	_ = Option[int].MustGet
	_ = Option[int].Or
	_ = Option[int].OrCall
	_ = Option[int].OrEmpty
	_ = Option[int].OrFalse
	_ = Option[int].OrZero
	_ = Option[int].ToOpt

	type _ = Any
	type _ = Bool
	type _ = Byte
	type _ = Error
	type _ = Int
	type _ = Rune
	type _ = String

	_ = Getenv("")
	_ = NonEmpty
	_ = NonNil[int]
	_ = NonZero[int]
	_ = Map(Option[int]{}, func(int) int { return 0 })
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
