// Package slice provides fluent slice types that can chain functional collection operations.
//
// Mapper[T] is a fluent slice that can chain operations like ToString (map), KeepIf (filter), etc.
//
// MapperTo[T, R] is a fluent slice with one additional method, MapTo, for mapping to a specified type R.
// If you don't need to map to an arbitrary type, use Mapper instead.
package slice

func _() {
	_ = From[int]

	_ = Mapper[int].Any
	_ = Mapper[int].Clone
	_ = Mapper[int].Convert
	_ = Mapper[int].Each
	_ = Mapper[int].Every
	_ = Mapper[int].Find
	_ = Mapper[int].None
	_ = Mapper[int].First
	_ = Mapper[int].FlatMap
	_ = Mapper[int].IndexWhere
	_ = Mapper[int].KeepIf
	_ = Mapper[int].Len
	_ = Mapper[int].ParallelEach
	_ = Mapper[int].ParallelKeepIf
	_ = Mapper[int].RemoveIf
	_ = Mapper[int].Reverse
	_ = Mapper[int].Single
	_ = Mapper[int].Take
	_ = Mapper[int].TakeLast
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
	_ = MapperTo[int, int].FlatMap
	_ = MapperTo[int, int].KeepIf
	_ = MapperTo[int, int].Len
	_ = MapperTo[int, int].Map
	_ = MapperTo[int, int].ParallelEach
	_ = MapperTo[int, int].ParallelKeepIf
	_ = MapperTo[int, int].ParallelMap
	_ = MapperTo[int, int].RemoveIf
	_ = MapperTo[int, int].Reverse
	_ = MapperTo[int, int].Single
	_ = MapperTo[int, int].Take
	_ = MapperTo[int, int].TakeLast
	_ = MapperTo[int, int].ToAny
	_ = MapperTo[int, int].ToBool
	_ = MapperTo[int, int].ToByte
	_ = MapperTo[int, int].ToError
	_ = MapperTo[int, int].ToInt
	_ = MapperTo[int, int].ToRune
	_ = MapperTo[int, int].ToString

	_ = String.ToSet

	_ = Float64.Max
	_ = Float64.Min

	_ = Int.Max
	_ = Int.Min
	_ = Int.Sum

	_ = Chunk[int]
	_ = Compact[int]
	_ = Contains[int]
	_ = FromMap[int, int]
	_ = FromSet[int]
	_ = GroupBy[int, int]
	_ = SortBy[int, int]
	_ = SortByDesc[int, int]

	_ = FindAs[int, int]
	_ = Fold[int, int]
	_ = Map[int, int]
	_ = MapAccum[int, int, int]
	_ = ParallelMap[int, int]
	_ = ToSet[int]
	_ = ToSetBy[int, int]
	_ = UniqueBy[int, int]
	_ = Unzip2[int, int, int]
	_ = Unzip3[int, int, int, int]
	_ = Unzip4[int, int, int, int, int]
}
