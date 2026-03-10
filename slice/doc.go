// Package slice provides fluent slice types that can chain functional collection operations.
//
// Types are defined in internal/base and aliased here. All methods are available through the aliases.
//
// Mapper[T] is a fluent slice that can chain operations like ToString (map), KeepIf (filter), etc.
//
// MapperTo[T, R] is a fluent slice with one additional method, MapTo, for mapping to a specified type R.
// If you don't need to map to an arbitrary type, use Mapper instead.
//
// Entries[K, V] is a fluent map type for chaining map operations.
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
	_ = Mapper[int].KeyByInt
	_ = Mapper[int].KeyByString
	_ = Mapper[int].KeepIf
	_ = Mapper[int].Last
	_ = Mapper[int].Len
	_ = Mapper[int].ParallelEach
	_ = Mapper[int].ParallelKeepIf
	_ = Mapper[int].Partition
	_ = Mapper[int].RemoveIf
	_ = Mapper[int].Reverse
	_ = Mapper[int].Single
	_ = Mapper[int].Sort
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
	_ = MapperTo[int, int].KeyByInt
	_ = MapperTo[int, int].KeyByString
	_ = MapperTo[int, int].Last
	_ = MapperTo[int, int].Len
	_ = MapperTo[int, int].Map
	_ = MapperTo[int, int].ParallelEach
	_ = MapperTo[int, int].ParallelKeepIf
	_ = MapperTo[int, int].ParallelMap
	_ = MapperTo[int, int].Partition
	_ = MapperTo[int, int].RemoveIf
	_ = MapperTo[int, int].Reverse
	_ = MapperTo[int, int].Single
	_ = MapperTo[int, int].Sort
	_ = MapperTo[int, int].Take
	_ = MapperTo[int, int].TakeLast
	_ = MapperTo[int, int].ToAny
	_ = MapperTo[int, int].ToBool
	_ = MapperTo[int, int].ToByte
	_ = MapperTo[int, int].ToError
	_ = MapperTo[int, int].ToInt
	_ = MapperTo[int, int].ToRune
	_ = MapperTo[int, int].ToString

	_ = Entries[int, int].Values
	_ = Entries[int, int].Keys
	_ = Entries[int, int].ToAny
	_ = Entries[int, int].ToBool
	_ = Entries[int, int].ToByte
	_ = Entries[int, int].ToError
	_ = Entries[int, int].ToFloat32
	_ = Entries[int, int].ToFloat64
	_ = Entries[int, int].ToInt
	_ = Entries[int, int].ToInt32
	_ = Entries[int, int].ToInt64
	_ = Entries[int, int].ToRune
	_ = Entries[int, int].ToString

	_ = String.NonEmpty
	_ = String.ToSet
	_ = Float64.Max
	_ = Float64.Min
	_ = Float64.Sum
	_ = Int.Max
	_ = Int.Min
	_ = Int.Sum

	_ = Chunk[int]
	_ = NonEmpty
	_ = NonZero[int]
	_ = Contains[int]
	_ = Difference[int]
	_ = Flatten[int]
	_ = FromSet[int]
	_ = Group[int, int]{}
	_ = GroupBy[int, int]
	_ = Intersect[int]
	_ = KeyBy[int, int]
	_ = Partition[int]
	_ = SortBy[int, int]
	_ = SortByDesc[int, int]
	_ = Union[int]

	_ = Asc[int, int]
	_ = Desc[int, int]

	_ = FanOut[int, int]
	_ = FanOutEach[int]
	_ = FanOutWeighted[int, int]
	_ = FanOutEachWeighted[int]
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
