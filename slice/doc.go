// Package slice provides fluent slice types that can chain functional collection operations.
//
// Mapper[T] is a fluent slice that can chain operations like ToString (map), KeepIf (filter), etc.
//
// MapperTo[T, R] is a fluent slice with one additional method, MapTo, for mapping to a specified type R.
// If you don't need to map to an arbitrary type, use Mapper instead.
package slice

func _() {
	_ = From[int]

	_ = Mapper[int].Clone
	_ = Mapper[int].Convert
	_ = Mapper[int].Each
	_ = Mapper[int].KeepIf
	_ = Mapper[int].Len
	_ = Mapper[int].RemoveIf
	_ = Mapper[int].Single
	_ = Mapper[int].TakeFirst
	_ = Mapper[int].ToAny
	_ = Mapper[int].ToBool
	_ = Mapper[int].ToByte
	_ = Mapper[int].ToError
	_ = Mapper[int].ToInt
	_ = Mapper[int].ToRune
	_ = Mapper[int].ToString

	_ = MapperTo[int, int].Clone
	_ = MapperTo[int, int].Convert
	_ = MapperTo[int, int].Each
	_ = MapperTo[int, int].KeepIf
	_ = MapperTo[int, int].Len
	_ = MapperTo[int, int].Map
	_ = MapperTo[int, int].RemoveIf
	_ = MapperTo[int, int].Single
	_ = MapperTo[int, int].TakeFirst
	_ = MapperTo[int, int].ToAny
	_ = MapperTo[int, int].ToBool
	_ = MapperTo[int, int].ToByte
	_ = MapperTo[int, int].ToError
	_ = MapperTo[int, int].ToInt
	_ = MapperTo[int, int].ToRune
	_ = MapperTo[int, int].ToString

	_ = String.ToSet

	_ = SortBy[int, int]
	_ = SortByDesc[int, int]

	_ = Fold[int, int]
	_ = Unzip2[int, int, int]
	_ = Unzip3[int, int, int, int]
	_ = Unzip4[int, int, int, int, int]
}
