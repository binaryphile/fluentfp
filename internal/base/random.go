package base

import (
	"math/rand/v2"

	"github.com/binaryphile/fluentfp/option"
)

// Shuffle returns a randomly reordered copy of ts.
func (ts Mapper[T]) Shuffle() Mapper[T] {
	if len(ts) == 0 {
		return Mapper[T]{}
	}

	c := make([]T, len(ts))
	copy(c, ts)
	rand.Shuffle(len(c), func(i, j int) { c[i], c[j] = c[j], c[i] })

	return c
}

// Sample returns a random element from ts, or not-ok if ts is empty.
func (ts Mapper[T]) Sample() option.Option[T] {
	if len(ts) == 0 {
		return option.NotOk[T]()
	}

	return option.Of(ts[rand.IntN(len(ts))])
}

// Samples returns count random elements from ts without replacement.
// If count >= len(ts), returns all elements in random order.
// Returns empty for count <= 0 or empty ts.
func (ts Mapper[T]) Samples(count int) Mapper[T] {
	if count <= 0 || len(ts) == 0 {
		return Mapper[T]{}
	}

	c := make([]T, len(ts))
	copy(c, ts)

	if count > len(c) {
		count = len(c)
	}

	for i := 0; i < count; i++ {
		j := i + rand.IntN(len(c)-i)
		c[i], c[j] = c[j], c[i]
	}

	return c[:count]
}
