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
	_ = Entries[int, int].Keys
	_ = Entries[int, int].ToRune
	_ = Entries[int, int].ToString
	_ = Entries[int, int].Values
	_ = Entries[int, int].KeepIf
	_ = Entries[int, int].RemoveIf

	_ = Map[int, int, int]
	_ = MapKeys[int, int, int]
	_ = MapValues[int, int, int]
	_ = Values[int, int]
	_ = Keys[int, int]

	_ = ToPairs[int, int]
	_ = FromPairs[int, int]

	_ = Invert[int, int]
	_ = Merge[int, int]
	_ = PickByKeys[int, int]
	_ = OmitByKeys[int, int]
}
