package option

type Bool = Basic[bool]

var (
	BoolOf    = BasicOf[bool]
	NewBool   = NewBasic[bool]
	NotOkBool = Bool{}
)

type String = Basic[string]

var (
	NewString       = NewBasic[string]
	NotOkString     = String{}
	StringOf        = BasicOf[string]
	StringOfNonZero = BasicOfNonZero[string]
)
