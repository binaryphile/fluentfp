// Package fluent provides fluent slice types that can chain functional collection operations.
//
// Mapper[T] is a fluent slice that can chain operations like ToString (map), KeepIf (filter), etc.
//
// MapperTo[T, R] is a fluent slice with one additional method, ToOther, for mapping to a specified type R.
// If you don't need to map to an arbitrary type, use Mapper instead.
package slice

func _() {
	_ = Mapper[bool]{}.Each
	_ = Mapper[bool]{}.KeepIf
	_ = Mapper[bool]{}.Len
	_ = Mapper[bool]{}.RemoveIf
	_ = Mapper[bool]{}.TakeFirst
	_ = Mapper[bool]{}.ToBool
	_ = Mapper[bool]{}.ToByte
	_ = Mapper[bool]{}.ToError
	_ = Mapper[bool]{}.ToInt
	_ = Mapper[bool]{}.ToRune
	_ = Mapper[bool]{}.Convert
	_ = Mapper[bool]{}.ToString

	_ = MapperTo[bool, bool]{}.Each
	_ = MapperTo[bool, bool]{}.KeepIf
	_ = MapperTo[bool, bool]{}.Len
	_ = MapperTo[bool, bool]{}.RemoveIf
	_ = MapperTo[bool, bool]{}.TakeFirst
	_ = MapperTo[bool, bool]{}.ToBool
	_ = MapperTo[bool, bool]{}.ToByte
	_ = MapperTo[bool, bool]{}.ToError
	_ = MapperTo[bool, bool]{}.ToInt
	_ = MapperTo[bool, bool]{}.To
	_ = MapperTo[bool, bool]{}.ToRune
	_ = MapperTo[bool, bool]{}.Convert
	_ = MapperTo[bool, bool]{}.ToString
}
