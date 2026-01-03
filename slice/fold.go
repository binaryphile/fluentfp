package slice

// Fold reduces a slice to a single value by applying fn to each element.
// It starts with initial and applies fn(accumulator, element) for each element from left to right.
// Returns initial if the slice is empty.
func Fold[T, R any](ts []T, initial R, fn func(R, T) R) R {
	acc := initial
	for _, t := range ts {
		acc = fn(acc, t)
	}

	return acc
}
