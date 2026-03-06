package kv

func _() {
	_ = From[int, int]
	_ = Entries[int, int].ToAny
	_ = Entries[int, int].ToBool
	_ = Entries[int, int].ToByte
	_ = Entries[int, int].ToError
	_ = Entries[int, int].ToFloat32
	_ = Entries[int, int].ToFloat64
	_ = Entries[int, int].ToInt
	_ = Entries[int, int].ToInt32
	_ = Entries[int, int].ToInt64
	_ = Entries[int, int].ToKeys
	_ = Entries[int, int].ToRune
	_ = Entries[int, int].ToString
	_ = Entries[int, int].ToValues

	_ = Map[int, int, int]
	_ = MapTo[int, int, int]
	_ = MapperTo[int, int, int].Map

	_ = Values[int, int]
	_ = Keys[int, int]
}
