package slice

import "github.com/binaryphile/fluentfp/internal/base"

// KeyBy indexes elements by a key derived from fn, returning a map from key to element.
// If multiple elements produce the same key, the last one wins.
// For common key types, prefer the method forms KeyByString and KeyByInt on Mapper.
func KeyBy[T any, K comparable](ts Mapper[T], fn func(T) K) map[K]T {
	return base.KeyBy(ts, fn)
}
