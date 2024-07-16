package iterator

import "github.com/binaryphile/funcTrunk/option"

type Iterator[T any] struct {
	option.Basic[func() (T, Iterator[T])]
}

func IteratorOf[T any](fn func() (T, Iterator[T])) (_ Iterator[T]) {
	if fn == nil {
		return
	}

	return Iterator[T]{
		Basic: option.BasicOf(fn),
	}
}

func FromSlice[T any](ts []T) Iterator[T] {
	if len(ts) == 0 {
		return Iterator[T]{}
	}

	i := 0

	var next Iterator[T]
	next = IteratorOf(func() (T, Iterator[T]) {
		if i == len(ts)-1 {
			return ts[i], Iterator[T]{}
		}

		t := ts[i]
		i += 1
		return t, next
	})

	return next
}
