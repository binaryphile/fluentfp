package pipeline

import "context"

// FromSlice sends each element of ts to the returned channel, then closes it.
// Respects ctx cancellation.
func FromSlice[T any](ctx context.Context, ts []T) <-chan T {
	out := make(chan T)

	go func() {
		defer close(out)

		for _, t := range ts {
			select {
			case <-ctx.Done():
				return
			case out <- t:
			}
		}
	}()

	return out
}

// Generate calls fn repeatedly, sending results to the returned channel.
// fn returns (value, more). When more is false or ctx cancels, the channel
// closes. fn must not be nil.
func Generate[T any](ctx context.Context, fn func() (T, bool)) <-chan T {
	out := make(chan T)

	go func() {
		defer close(out)

		for {
			val, more := fn()
			if !more {
				return
			}

			select {
			case <-ctx.Done():
				return
			case out <- val:
			}
		}
	}()

	return out
}
