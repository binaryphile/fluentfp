package option

type (
	Any    = Basic[any]
	Bool   = Basic[bool]
	Byte   = Basic[byte]
	Error  = Basic[error]
	Int    = Basic[int]
	Rune   = Basic[rune]
	String = Basic[string]
)

var (
	NotOkAny    = Basic[any]{}
	NotOkBool   = Basic[bool]{}
	NotOkByte   = Basic[byte]{}
	NotOkError  = Basic[error]{}
	NotOkInt    = Basic[int]{}
	NotOkRune   = Basic[rune]{}
	NotOkString = Basic[string]{}
)
