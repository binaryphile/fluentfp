// Package fluent provides fluent slice types that can chain functional collection operations.
//
// SliceOf[T] is a fluent slice that can chain operations like ToString (map), KeepIf (filter), etc.
//
// Mapper[T, R] is a fluent slice with one additional method, ToOther, for mapping to a specified type R.
// If you don't need to map to an arbitrary type, use SliceOf instead.
package fluent

func _() {
	_ = SliceOf[bool]{}.Each
	_ = SliceOf[bool]{}.KeepIf
	_ = SliceOf[bool]{}.Len
	_ = SliceOf[bool]{}.RemoveIf
	_ = SliceOf[bool]{}.TakeFirst
	_ = SliceOf[bool]{}.ToBool
	_ = SliceOf[bool]{}.ToByte
	_ = SliceOf[bool]{}.ToError
	_ = SliceOf[bool]{}.ToInt
	_ = SliceOf[bool]{}.ToRune
	_ = SliceOf[bool]{}.ToSame
	_ = SliceOf[bool]{}.ToString

	_ = Mapper[bool, bool]{}.Each
	_ = Mapper[bool, bool]{}.KeepIf
	_ = Mapper[bool, bool]{}.Len
	_ = Mapper[bool, bool]{}.RemoveIf
	_ = Mapper[bool, bool]{}.TakeFirst
	_ = Mapper[bool, bool]{}.ToBool
	_ = Mapper[bool, bool]{}.ToByte
	_ = Mapper[bool, bool]{}.ToError
	_ = Mapper[bool, bool]{}.ToInt
	_ = Mapper[bool, bool]{}.ToOther
	_ = Mapper[bool, bool]{}.ToRune
	_ = Mapper[bool, bool]{}.ToSame
	_ = Mapper[bool, bool]{}.ToString
}
