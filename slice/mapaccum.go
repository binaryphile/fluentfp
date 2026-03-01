package slice

// MapAccum threads state through a slice, producing both a final state and a mapped output.
// fn receives the accumulated state and current element, returning new state and an output value.
// Returns init and an empty slice if ts is empty.
func MapAccum[T, R, S any](ts []T, init S, fn func(S, T) (S, R)) (S, Mapper[R]) {
	acc := init
	rs := make([]R, len(ts))
	for i, t := range ts {
		acc, rs[i] = fn(acc, t)
	}

	return acc, rs
}
