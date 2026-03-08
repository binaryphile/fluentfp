package hof

// Identity returns its argument unchanged.
// Use as a function value via type instantiation: hof.Identity[string]
func Identity[T any](t T) T {
	return t
}
