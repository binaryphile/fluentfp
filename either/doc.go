// Package either provides a sum type representing a value of one of two types.
//
// Convention: Left represents failure/error, Right represents success.
// Mnemonic: "Right is right" (correct).
//
// Either is right-biased: Map, MustGet, IfRight, GetOr operate on the Right value.
// Use MapLeft, MustGetLeft, IfLeft, LeftOr for Left-side operations.
package either

// Compile-time API verification
func _() {
	// Constructors
	_ = Left[int, string]
	_ = Right[int, string]

	// Accessors
	_ = Either[int, string].Get
	_ = Either[int, string].GetLeft
	_ = Either[int, string].IsLeft
	_ = Either[int, string].IsRight
	_ = Either[int, string].MustGet
	_ = Either[int, string].MustGetLeft

	// Defaults
	_ = Either[int, string].GetOr
	_ = Either[int, string].LeftOr
	_ = Either[int, string].GetOrCall
	_ = Either[int, string].LeftOrCall

	// Transforms
	_ = Either[int, string].Map
	_ = Fold[int, string, bool]
	_ = Map[int, string, bool]
	_ = MapLeft[int, string, bool]

	// Side effects
	_ = Either[int, string].IfRight
	_ = Either[int, string].IfLeft
}
