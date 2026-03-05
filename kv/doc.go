package kv

func _() {
	_ = From[int, int]
	_ = Entries[int, int].ToValues
	_ = Entries[int, int].ToKeys

	_ = Map[int, int, int]
	_ = MapTo[int, int, int]
	_ = MapperTo[int, int, int].Map

	_ = Values[int, int]
	_ = Keys[int, int]
}
