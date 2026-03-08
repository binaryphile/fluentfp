package fn

// Identity returns its argument unchanged.
// Use as a function value via type instantiation: fn.Identity[string]
func Identity[T any](t T) T {
	return t
}
