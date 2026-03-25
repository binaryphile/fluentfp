package hof_test

import . "github.com/binaryphile/fluentfp/hof"

// Compile-time export verification.
func _() {
	// Composition
	_ = Pipe[int, int, int]

	// Partial application
	_ = Bind[int, int, int]
	_ = BindR[int, int, int]

	// Independent application
	_ = Cross[int, int, int, int]

	// Building blocks
	_ = Eq[int]

	// Call coalescing
	_ = NewDebouncer[int]
	_ = MaxWait
}
