// Package hof provides pure function combinators: composition, partial
// application, and independent application. Based on Stone's
// "Algorithms: A Functional Programming Approach" (pipe, sect, cross).
//
// For resilience decorators (retry, circuit breaker, throttle, debounce),
// see the [cb] package.
package hof

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
}
