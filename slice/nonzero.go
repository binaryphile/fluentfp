package slice

// NonZero removes zero-value elements from ts.
func NonZero[T comparable](ts Mapper[T]) Mapper[T] {
	var zero T
	return From(ts).KeepIf(func(t T) bool { return t != zero })
}

// NonEmpty removes empty strings from ts.
func NonEmpty(ts Mapper[string]) Mapper[string] {
	return NonZero(ts)
}
