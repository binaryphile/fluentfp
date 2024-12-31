package pair

type For[V1, V2 any] struct {
	V1 V1
	V2 V2
}

func New[V, V2 any](v V, v2 V2) For[V, V2] {
	return For[V, V2]{
		V1: v,
		V2: v2,
	}
}
