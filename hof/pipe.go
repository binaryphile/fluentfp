package hof

// Pipe composes two functions left-to-right: Pipe(f, g)(x) = g(f(x)).
// Panics if f or g is nil.
func Pipe[A, B, C any](f func(A) B, g func(B) C) func(A) C {
	if f == nil {
		panic("hof.Pipe: f must not be nil")
	}
	if g == nil {
		panic("hof.Pipe: g must not be nil")
	}

	return func(a A) C { return g(f(a)) }
}
