package pair

// X is a value of the Cartesian cross-product of the given types.
// Cross-product is represented by an x.
type X[A, B any] struct {
	First  A
	Second B
}

func Of[A, B any](a A, b B) X[A, B] {
	return X[A, B]{
		First:  a,
		Second: b,
	}
}
