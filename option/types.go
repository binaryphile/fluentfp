package option

type (
	Any    = Option[any]
	Bool   = Option[bool]
	Byte   = Option[byte]
	Error  = Option[error]
	Int    = Option[int]
	Rune   = Option[rune]
	String = Option[string]
)

var (
	NotOkAny    = Option[any]{}
	NotOkBool   = Option[bool]{}
	NotOkByte   = Option[byte]{}
	NotOkError  = Option[error]{}
	NotOkInt    = Option[int]{}
	NotOkRune   = Option[rune]{}
	NotOkString = Option[string]{}
)
