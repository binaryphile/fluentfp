//go:build ignore

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
	_ = (Basic[bool]{}).ToOpt
	_ = FromOpt[bool]
	_ = Getenv("")
	_ = IfProvided[bool]
	_ = Map(Basic[bool]{}, func(bool) bool { return true })
	_ = New[bool]
	_ = NotOkAny
	_ = NotOkBool
	_ = NotOkByte
	_ = NotOkError
	_ = NotOkInt
	_ = NotOkRune
	_ = NotOkString
	_ = Of[bool]
}
