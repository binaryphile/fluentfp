package pipeline

import "context"

// Filter sends only elements where fn returns true.
// Single goroutine — no concurrency needed for a pure predicate.
// fn must not be nil.
func Filter[T any](ctx context.Context, in <-chan T, fn func(T) bool) <-chan T {
	out := make(chan T)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case val, ok := <-in:
				if !ok {
					return
				}

				if fn(val) {
					select {
					case <-ctx.Done():
						return
					case out <- val:
					}
				}
			}
		}
	}()

	return out
}
