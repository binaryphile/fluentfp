package option

// NonZeroMap returns an ok option of fn(t) provided that t is not the zero value for T, or not-ok otherwise.
// It combines NonZero and Map in one call — check presence and transform in a single step.
func NonZeroMap[T comparable, R any](t T, fn func(T) R) (_ Basic[R]) {
	var zero T
	if t == zero {
		return
	}

	return Of(fn(t))
}

// NonEmptyMap returns an ok option of fn(s) provided that s is not empty, or not-ok otherwise.
// It is the string-specific variant of NonZeroMap.
func NonEmptyMap[R any](s string, fn func(string) R) (_ Basic[R]) {
	if s == "" {
		return
	}

	return Of(fn(s))
}

// NonNilMap returns an ok option of fn(*t) provided that t is not nil, or not-ok otherwise.
// It dereferences the pointer before passing to fn, matching NonNil's behavior.
func NonNilMap[T any, R any](t *T, fn func(T) R) (_ Basic[R]) {
	if t == nil {
		return
	}

	return Of(fn(*t))
}
