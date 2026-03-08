package hof

// Bind fixes the first argument of a binary function: Bind(f, x)(y) = f(x, y).
// Panics if f is nil.
func Bind[A, B, C any](f func(A, B) C, a A) func(B) C {
	if f == nil {
		panic("hof.Bind: f must not be nil")
	}

	return func(b B) C { return f(a, b) }
}

// BindR fixes the second argument of a binary function: BindR(f, y)(x) = f(x, y).
// Panics if f is nil.
func BindR[A, B, C any](f func(A, B) C, b B) func(A) C {
	if f == nil {
		panic("hof.BindR: f must not be nil")
	}

	return func(a A) C { return f(a, b) }
}
