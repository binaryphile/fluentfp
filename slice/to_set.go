package slice

import "github.com/binaryphile/fluentfp/internal/base"

// ToSet returns a map with each element as a key set to true.
// Requires comparable elements.
func ToSet[T comparable](ts []T) map[T]bool {
	return base.ToSet(ts)
}
