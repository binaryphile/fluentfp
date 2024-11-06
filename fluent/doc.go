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

	_ = SliceToNamed[bool, bool]{}.Convert
	_ = SliceToNamed[bool, bool]{}.Each
	_ = SliceToNamed[bool, bool]{}.KeepIf
	_ = SliceToNamed[bool, bool]{}.Len
	_ = SliceToNamed[bool, bool]{}.Map
	_ = SliceToNamed[bool, bool]{}.RemoveIf
	_ = SliceToNamed[bool, bool]{}.TakeFirst
	_ = SliceToNamed[bool, bool]{}.ToAnys
	_ = SliceToNamed[bool, bool]{}.ToBools
	_ = SliceToNamed[bool, bool]{}.ToBytes
	_ = SliceToNamed[bool, bool]{}.ToErrors
	_ = SliceToNamed[bool, bool]{}.ToInts
	_ = SliceToNamed[bool, bool]{}.ToRunes
	_ = SliceToNamed[bool, bool]{}.ToStrings
}
