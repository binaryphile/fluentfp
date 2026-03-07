package fn

// Dispatch2 applies two functions to the same argument, returning both results.
// Panics if f or g is nil.
func Dispatch2[A, B, C any](f func(A) B, g func(A) C) func(A) (B, C) {
	if f == nil {
		panic("fn.Dispatch2: f must not be nil")
	}
	if g == nil {
		panic("fn.Dispatch2: g must not be nil")
	}

	return func(a A) (B, C) { return f(a), g(a) }
}

// Dispatch3 applies three functions to the same argument, returning all results.
// Panics if f, g, or h is nil.
func Dispatch3[A, B, C, D any](f func(A) B, g func(A) C, h func(A) D) func(A) (B, C, D) {
	if f == nil {
		panic("fn.Dispatch3: f must not be nil")
	}
	if g == nil {
		panic("fn.Dispatch3: g must not be nil")
	}
	if h == nil {
		panic("fn.Dispatch3: h must not be nil")
	}

	return func(a A) (B, C, D) { return f(a), g(a), h(a) }
}
