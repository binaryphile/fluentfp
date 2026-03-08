// Package hof provides function combinators for composition, partial application,
// independent application, and concurrency control. Based on Stone's "Algorithms:
// A Functional Programming Approach" (pipe, sect, cross).
package hof

// Compile-time export verification. Every fluentfp package uses this pattern
// to ensure exported symbols remain available across refactors.
func _() {
	// Composition
	_ = Pipe[int, int, int]

	// Partial application
	_ = Bind[int, int, int]
	_ = BindR[int, int, int]

	// Independent application
	_ = Cross[int, int, int, int]

	// Building blocks
	_ = Identity[int]
	_ = Eq[int]

	// Concurrency control
	_ = Throttle[int, int]
	_ = ThrottleWeighted[int, int]
}
