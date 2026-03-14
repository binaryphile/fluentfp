package base

import "math/rand/v2"

// Shuffle returns a randomly reordered copy of ts.
func (ts MapperTo[R, T]) Shuffle() MapperTo[R, T] {
	if len(ts) == 0 {
		return MapperTo[R, T]{}
	}

	c := make([]T, len(ts))
	copy(c, ts)
	rand.Shuffle(len(c), func(i, j int) { c[i], c[j] = c[j], c[i] })

	return c
}

// Samples returns count random elements from ts without replacement.
// If count >= len(ts), returns all elements in random order.
// Returns empty for count <= 0 or empty ts.
func (ts MapperTo[R, T]) Samples(count int) MapperTo[R, T] {
	if count <= 0 || len(ts) == 0 {
		return MapperTo[R, T]{}
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
