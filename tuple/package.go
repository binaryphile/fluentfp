package tuple

func NewPair[V, V2 any](v V, v2 V2) Pair[V, V2] {
	return Pair[V, V2]{
		V1: v,
		V2: v2,
	}
}
