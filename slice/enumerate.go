package slice

import "github.com/binaryphile/fluentfp/tuple/pair"

// Enumerate pairs each element with its zero-based index.
// Preserves input order. Returns nil for nil input.
func Enumerate[T any](ts []T) Mapper[pair.Pair[int, T]] {
	if ts == nil {
		return nil
	}

	result := make(Mapper[pair.Pair[int, T]], len(ts))

	for i, t := range ts {
		result[i] = pair.Of(i, t)
	}

	return result
}
