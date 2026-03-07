// Package fn provides function combinators for composition, partial application,
// and multi-function dispatch. Based on Stone's "Algorithms: A Functional
// Programming Approach" (pipe, sect, dispatch, cross).
package fn

// Compile-time export verification. Every fluentfp package uses this pattern
// to ensure exported symbols remain available across refactors.
func _() {
	// Composition
	_ = Pipe[int, int, int]

	// Partial application
	_ = Bind[int, int, int]
	_ = BindR[int, int, int]

	// Multi-function dispatch
	_ = Dispatch2[int, int, int]
	_ = Dispatch3[int, int, int, int]

	// Independent application
	_ = Cross[int, int, int, int]
}
