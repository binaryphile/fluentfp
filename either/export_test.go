package either_test

import . "github.com/binaryphile/fluentfp/either"

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
