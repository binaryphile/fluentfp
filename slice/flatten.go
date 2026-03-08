package slice

// Flatten concatenates nested slices into a single flat slice,
// preserving element order.
func Flatten[T any](tss [][]T) Mapper[T] {
	if len(tss) == 0 {
		return Mapper[T]{}
	}

	n := 0
	for _, ts := range tss {
		n += len(ts)
	}

	result := make(Mapper[T], 0, n)
	for _, ts := range tss {
		result = append(result, ts...)
	}

	return result
}
