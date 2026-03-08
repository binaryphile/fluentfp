package hof

// Cross applies two functions independently to two separate arguments.
// Cross(f, g)(a, b) = (f(a), g(b)).
// Panics if f or g is nil.
func Cross[A, B, C, D any](f func(A) C, g func(B) D) func(A, B) (C, D) {
	if f == nil {
		panic("hof.Cross: f must not be nil")
	}
	if g == nil {
		panic("hof.Cross: g must not be nil")
	}

	return func(a A, b B) (C, D) { return f(a), g(b) }
}
