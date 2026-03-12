package slice

// Tally groups elements by their value, returning a group per distinct value.
// Equivalent to GroupBy with an identity key extractor.
func Tally[T comparable](ts []T) Mapper[Group[T, T]] {
	identity := func(t T) T { return t }
	return GroupBy(ts, identity)
}
