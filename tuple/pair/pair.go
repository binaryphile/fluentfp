package pair

// Pair holds two values of arbitrary types.
type Pair[A, B any] struct {
	First  A
	Second B
}

func Of[A, B any](a A, b B) Pair[A, B] {
	return Pair[A, B]{
		First:  a,
		Second: b,
	}
}
