package option

// NonZeroWith returns an ok option of fn(t) provided that t is not the zero value for T, or not-ok otherwise.
// It combines NonZero and a transform in one call — check presence and transform in a single step.
func NonZeroWith[T comparable, R any](t T, fn func(T) R) (_ Option[R]) {
	var zero T
	if t == zero {
		return
	}

	return Of(fn(t))
}

// NonEmptyWith returns an ok option of fn(s) provided that s is not empty, or not-ok otherwise.
// It is the string-specific variant of NonZeroWith.
func NonEmptyWith[R any](s string, fn func(string) R) (_ Option[R]) {
	if s == "" {
		return
	}

	return Of(fn(s))
}

// NonNilWith returns an ok option of fn(*t) provided that t is not nil, or not-ok otherwise.
// It dereferences the pointer before passing to fn, matching NonNil's behavior.
func NonNilWith[T any, R any](t *T, fn func(T) R) (_ Option[R]) {
	if t == nil {
		return
	}

	return Of(fn(*t))
}
