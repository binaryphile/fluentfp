package pipeline

import "context"

// Batch collects elements into slices of the given size.
// Emits a partial batch when input closes mid-batch.
// Each emitted slice is an independent copy.
// Panics if size <= 0.
func Batch[T any](ctx context.Context, in <-chan T, size int) <-chan []T {
	if size <= 0 {
		panic("pipeline.Batch: size must be > 0")
	}

	out := make(chan []T)

	go func() {
		defer close(out)

		buf := make([]T, 0, size)

		for {
			select {
			case <-ctx.Done():
				return
			case val, ok := <-in:
				if !ok {
					if len(buf) > 0 {
						select {
						case <-ctx.Done():
						case out <- buf:
						}
					}

					return
				}

				buf = append(buf, val)

				if len(buf) == size {
					select {
					case <-ctx.Done():
						return
					case out <- buf:
					}

					buf = make([]T, 0, size)
				}
			}
		}
	}()

	return out
}
