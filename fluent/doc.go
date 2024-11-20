package fluent

func _() {
	_ = SliceOf[bool]{}.ToSame
	_ = SliceOf[bool]{}.Each
	_ = SliceOf[bool]{}.KeepIf
	_ = SliceOf[bool]{}.Len
	_ = SliceOf[bool]{}.RemoveIf
	_ = SliceOf[bool]{}.TakeFirst
	_ = SliceOf[bool]{}.ToAny
	_ = SliceOf[bool]{}.ToBool
	_ = SliceOf[bool]{}.ToByte
	_ = SliceOf[bool]{}.ToError
	_ = SliceOf[bool]{}.ToInt
	_ = SliceOf[bool]{}.ToRune
	_ = SliceOf[bool]{}.ToString

	_ = SliceToNamed[bool, bool]{}.ToSame
	_ = SliceToNamed[bool, bool]{}.Each
	_ = SliceToNamed[bool, bool]{}.KeepIf
	_ = SliceToNamed[bool, bool]{}.Len
	_ = SliceToNamed[bool, bool]{}.ToNamed
	_ = SliceToNamed[bool, bool]{}.RemoveIf
	_ = SliceToNamed[bool, bool]{}.TakeFirst
	_ = SliceToNamed[bool, bool]{}.ToAny
	_ = SliceToNamed[bool, bool]{}.ToBool
	_ = SliceToNamed[bool, bool]{}.ToByte
	_ = SliceToNamed[bool, bool]{}.ToError
	_ = SliceToNamed[bool, bool]{}.ToInt
	_ = SliceToNamed[bool, bool]{}.ToRune
	_ = SliceToNamed[bool, bool]{}.ToString
}
