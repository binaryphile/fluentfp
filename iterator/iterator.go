package iterator

type Of[T any] func() (T, bool)

func FromSlice[T any](ts []T) func() (T, bool) {
	mine := make([]T, len(ts))
	_ = copy(mine, ts)

	i := -1
	return func() (_ T, _ bool) {
		if i == len(mine)-1 {
			return
		}

		i += 1
		return mine[i], true
	}
}
