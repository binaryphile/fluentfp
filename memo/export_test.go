package memo_test

import . "github.com/binaryphile/fluentfp/memo"

// Compile-time export verification. Every fluentfp package uses this pattern
// to ensure exported symbols remain available across refactors.
func _() {
	// Zero-arg memoization
	_ = From[int]

	// Keyed memoization
	_ = Fn[int, int]
	_ = FnErr[int, int]
	_ = FnWith[int, int]
	_ = FnErrWith[int, int]

	// Cache constructors
	_ = NewMap[int, int]
	_ = NewLRU[int, int]
}
