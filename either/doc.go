// Package either provides a sum type representing either a Left or a Right value.
//
// Either[L, R] represents one of two possible values. The zero value is Left
// containing L's zero value — in particular, Either[error, R]{} is Left(nil),
// a Left with no meaningful error. Always construct explicitly via [Left] or [Right].
//
// Either is right-biased: the primary transform and extraction methods (Transform,
// FlatMap, MustGet, IfRight, Or) operate on the Right value. Left-side accessors
// (MustGetLeft, IfLeft, LeftOr) and the recovery combinator [FlatMapLeft] are provided
// for cases where the Left value matters, but the API favors Right-side chaining.
//
// Convention: Left represents failure/error, Right represents success.
// Mnemonic: "Right is right" (correct).
//
// # Methods vs standalone functions
//
// Methods are used when the return type can be expressed using the receiver's
// existing type parameters, including reordering (Swap returns Either[R, L]).
// Standalone functions are needed when new type parameters must be introduced
// (Map, MapLeft, cross-type FlatMap, Fold).
//
// Method FlatMap is same-right-type only: func(R) Either[L, R].
// Standalone FlatMap allows changing the right type: func(R) Either[L, R2].
//
// # Storage
//
// Either stores both L and R fields inline. For large payload types,
// prefer pointers to avoid copy overhead on value-receiver method calls.
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
	_ = Either[int, string].LeftOr
	_ = Either[int, string].LeftOrCall
	_ = Either[int, string].Or
	_ = Either[int, string].OrCall

	// Transforms — methods
	_ = Either[int, string].Transform
	_ = Either[int, string].FlatMap
	_ = Either[int, string].FlatMapLeft
	_ = Either[int, string].Swap

	// Transforms — standalone (new type parameters)
	_ = FlatMap[int, string, bool]
	_ = Fold[int, string, bool]
	_ = Map[int, string, bool]
	_ = MapLeft[int, string, bool]

	// Side effects
	_ = Either[int, string].IfLeft
	_ = Either[int, string].IfRight
}
