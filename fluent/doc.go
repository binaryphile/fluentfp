package fluent

func _() {
	_ = SliceOf[bool]{}.Convert
	_ = SliceOf[bool]{}.Each
	_ = SliceOf[bool]{}.KeepIf
	_ = SliceOf[bool]{}.Len
	_ = SliceOf[bool]{}.RemoveIf
	_ = SliceOf[bool]{}.TakeFirst
	_ = SliceOf[bool]{}.ToAnys
	_ = SliceOf[bool]{}.ToBools
	_ = SliceOf[bool]{}.ToBytes
	_ = SliceOf[bool]{}.ToErrors
	_ = SliceOf[bool]{}.ToInts
	_ = SliceOf[bool]{}.ToRunes
	_ = SliceOf[bool]{}.ToStrings

	_ = MappableSliceOf[bool, bool]{}.Convert
	_ = MappableSliceOf[bool, bool]{}.Each
	_ = MappableSliceOf[bool, bool]{}.KeepIf
	_ = MappableSliceOf[bool, bool]{}.Len
	_ = MappableSliceOf[bool, bool]{}.Map
	_ = MappableSliceOf[bool, bool]{}.RemoveIf
	_ = MappableSliceOf[bool, bool]{}.TakeFirst
	_ = MappableSliceOf[bool, bool]{}.ToAnys
	_ = MappableSliceOf[bool, bool]{}.ToBools
	_ = MappableSliceOf[bool, bool]{}.ToBytes
	_ = MappableSliceOf[bool, bool]{}.ToErrors
	_ = MappableSliceOf[bool, bool]{}.ToInts
	_ = MappableSliceOf[bool, bool]{}.ToRunes
	_ = MappableSliceOf[bool, bool]{}.ToStrings
}
