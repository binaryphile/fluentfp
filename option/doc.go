// Package option provides types and functions to work with optional values.
//
// Option[T] holds a value plus an ok flag. The zero value is not-ok,
// so it works without initialization. Construct via Of, New, When, WhenFunc,
// NonZero, NonEmpty, NonNil, or Env.
package option

func _() {
	// methods
	_ = Option[int].Convert
	_ = Option[int].FlatMap
	_ = Option[int].Get
	_ = Option[int].IfNotOk
	_ = Option[int].IfOk
	_ = Option[int].IsOk
	_ = Option[int].KeepIf
	_ = Option[int].MustGet
	_ = Option[int].Or
	_ = Option[int].OrCall
	_ = Option[int].OrElse
	_ = Option[int].OrEmpty
	_ = Option[int].OrZero
	_ = Option[int].RemoveIf
	_ = Option[int].ToAny
	_ = Option[int].ToBool
	_ = Option[int].ToByte
	_ = Option[int].ToError
	_ = Option[int].ToInt
	_ = Option[int].ToOpt
	_ = Option[int].ToRune
	_ = Option[int].ToString

	// type aliases
	type _ = Any
	type _ = Bool
	type _ = Byte
	type _ = Error
	type _ = Int
	type _ = Rune
	type _ = String

	// standalone functions
	_ = Env("")
	_ = FlatMap(Option[int]{}, func(int) Option[int] { return Option[int]{} })
	_ = Lift(func(int) {})
	_ = Lookup(map[int]int{}, 0)
	_ = Map(Option[int]{}, func(int) int { return 0 })
	_ = New[int]
	_ = NonEmpty
	_ = NonEmptyCall("", func(string) int { return 0 })
	_ = NonErr[int]
	_ = NonNil[int]
	_ = NonNilCall[int, int](nil, func(int) int { return 0 })
	_ = NonZero[int]
	_ = NonZeroCall[int, int](0, func(int) int { return 0 })
	_ = NotOk[int]
	_ = Of[int]
	_ = OrFalse
	_ = When[int](false, 0)
	_ = WhenFunc[int](false, func() int { return 0 })

	// pre-declared not-ok values
	_ = NotOkAny
	_ = NotOkBool
	_ = NotOkByte
	_ = NotOkError
	_ = NotOkInt
	_ = NotOkRune
	_ = NotOkString
}
