package slice

// GroupSame groups elements by their own value, returning a group per distinct value.
// Equivalent to GroupBy with an identity key extractor.
func GroupSame[T comparable](ts []T) Mapper[Group[T, T]] {
	identity := func(t T) T { return t }
	return GroupBy(ts, identity)
}
