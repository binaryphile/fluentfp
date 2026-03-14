// Package slice provides fluent slice types that can chain functional collection operations.
//
// Mapper[T] is a fluent slice that can chain operations like ToString (map), KeepIf (filter), etc.
//
// Entries[K, V] is a fluent map type for chaining map operations.
package slice

func _() {
	_ = From[int]

	_ = Mapper[int].Any
	_ = Mapper[int].Clone
	_ = Mapper[int].Convert
	_ = Mapper[int].Drop
	_ = Mapper[int].DropLast
	_ = Mapper[int].DropLastWhile
	_ = Mapper[int].DropWhile
	_ = Mapper[int].Each
	_ = Mapper[int].Every
	_ = Mapper[int].Find
	_ = Mapper[int].FindLast
	_ = Mapper[int].First
	_ = Mapper[int].FlatMap
	_ = Mapper[int].IndexWhere
	_ = Mapper[int].Intersperse
	_ = Mapper[int].IsSorted
	_ = Mapper[int].KeyByInt
	_ = Mapper[int].KeyByString
	_ = Mapper[int].KeepIf
	_ = Mapper[int].Last
	_ = Mapper[int].LastIndexWhere
	_ = Mapper[int].Len
	_ = Mapper[int].None
	_ = Mapper[int].PEach
	_ = Mapper[int].PKeepIf
	_ = Mapper[int].Partition
	_ = Mapper[int].RemoveIf
	_ = Mapper[int].Reverse
	_ = Mapper[int].Sample
	_ = Mapper[int].Samples
	_ = Mapper[int].Shuffle
	_ = Mapper[int].Single
	_ = Mapper[int].Sort
	_ = Mapper[int].Take
	_ = Mapper[int].TakeLast
	_ = Mapper[int].TakeWhile
	_ = Mapper[int].ToAny
	_ = Mapper[int].ToBool
	_ = Mapper[int].ToByte
	_ = Mapper[int].ToError
	_ = Mapper[int].ToFloat32
	_ = Mapper[int].ToFloat64
	_ = Mapper[int].ToInt
	_ = Mapper[int].ToInt32
	_ = Mapper[int].ToInt64
	_ = Mapper[int].ToRune
	_ = Mapper[int].ToString

	_ = Entries[int, int].Values
	_ = Entries[int, int].Keys
	_ = Entries[int, int].KeepIf
	_ = Entries[int, int].RemoveIf
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

	_ = String.Contains
	_ = String.ContainsAny
	_ = String.Each
	_ = String.Join
	_ = String.Len
	_ = String.NonEmpty
	_ = String.ToSet
	_ = String.Unique
	_ = Float64.Max
	_ = Float64.Min
	_ = Float64.Sum
	_ = Int.Max
	_ = Int.Min
	_ = Int.Sum

	_ = Associate[int, int, int]
	_ = Chunk[int]
	_ = NonEmpty
	_ = NonZero[int]
	_ = Contains[int]
	_ = Difference[int]
	_ = Enumerate[int]
	_ = FilterMap[int, int]
	_ = IndexOf[int]
	_ = FlatMap[int, int]
	_ = Flatten[int]
	_ = FromSet[int]
	_ = Group[int, int]{}
	_ = GroupBy[int, int]
	_ = Intersect[int]
	_ = KeyBy[int, int]
	_ = LastIndexOf[int]
	_ = Partition[int]
	_ = SortBy[int, int]
	_ = SortByDesc[int, int]
	_ = GroupSame[int]
	_ = Union[int]

	_ = Asc[int, int]
	_ = Desc[int, int]
	_ = IsSortedBy[int, int]

	_ = FanOutAll[int, int]
	_ = FanOutWeightedAll[int, int]
	_ = FanOut[int, int]
	_ = FanOutEach[int]
	_ = FanOutWeighted[int, int]
	_ = FanOutEachWeighted[int]
	_ = FindAs[int, int]
	_ = Scan[int, int]
	_ = Zip[int, int]
	_ = ZipWith[int, int, int]
	_ = Range
	_ = RangeFrom
	_ = RangeStep
	_ = RepeatN[int]
	_ = Window[int]
	_ = Fold[int, int]
	_ = Map[int, int]
	_ = MapAccum[int, int, int]
	_ = MaxBy[int, int]
	_ = MinBy[int, int]
	_ = PFlatMap[int, int]
	_ = PMap[int, int]
	_ = Reduce[int]
	_ = ToSet[int]
	_ = ToSetBy[int, int]
	_ = Unique[int]
	_ = UniqueBy[int, int]
	_ = Unzip2[int, int, int]
	_ = Unzip3[int, int, int, int]
	_ = Unzip4[int, int, int, int, int]
}
