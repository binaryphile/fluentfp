package slice

// Compact removes zero-value elements from ts.
func Compact[T comparable](ts []T) Mapper[T] {
	var zero T
	return From(ts).KeepIf(func(t T) bool { return t != zero })
}
