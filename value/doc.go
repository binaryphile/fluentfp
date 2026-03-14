// Package value provides value-first conditional selection.
//
// Use this for selecting a value with a fallback, not for executing
// branches of logic.
//
//	color := value.Of(warn).When(critical).Or(calm)
//
// [Of] wraps a value eagerly; [LazyOf] wraps a function that is only
// called if the condition is true. Both return [option.Option] from
// [Cond.When] / [LazyCond.When], so the full option API is available
// for resolution (.Or, .OrCall, .Get, etc.).
//
// [FirstNonZero], [FirstNonEmpty], and [FirstNonNilValue] provide
// multi-value fallback without the option intermediary.
package value

// Compile-time API verification
func _() {
	// Conditional selection
	_ = Of[int]
	_ = Cond[int].When
	_ = LazyOf[int]
	_ = LazyCond[int].When

	// Fallback helpers
	_ = FirstNonEmpty
	_ = FirstNonNilValue[int]
	_ = FirstNonZero[int]

	// Re-exported option constructors
	_ = NonEmpty
	_ = NonEmptyCall[int]
	_ = NonNil[int]
	_ = NonNilCall[int, int]
	_ = NonZero[int]
	_ = NonZeroCall[int, int]
}
