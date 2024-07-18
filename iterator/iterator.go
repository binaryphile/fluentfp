package iterator

type Of[T any] func() (T, bool)

func FromSlice[T any](ts []T) func() (T, bool) {
	mine := make([]T, len(ts))
	_ = copy(mine, ts)

	i := 0
	return func() (_ T, _ bool) {
		if i == len(mine) {
			return
		}

		return mine[i], true
	}
}
