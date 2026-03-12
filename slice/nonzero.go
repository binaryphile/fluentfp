package slice

// NonZero removes zero-value elements from ts.
func NonZero[T comparable](ts []T) Mapper[T] {
	var zero T
	return From(ts).KeepIf(func(t T) bool { return t != zero })
}

// NonEmpty removes empty strings from ts.
func NonEmpty(ts []string) Mapper[string] {
	return NonZero(ts)
}
