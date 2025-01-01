package pair

// X is a value of the Cartesian cross-product of the given types.
// Cross-product is represented by an x.
type X[V1, V2 any] struct {
	V1 V1
	V2 V2
}

func Of[V, V2 any](v V, v2 V2) X[V, V2] {
	return X[V, V2]{
		V1: v,
		V2: v2,
	}
}
