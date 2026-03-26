package pipeline

import (
	"context"
	"sync"
)

// Merge combines multiple input channels into a single output channel.
// Output order is nondeterministic. Closes when all inputs are closed
// or ctx cancels.
func Merge[T any](ctx context.Context, ins ...<-chan T) <-chan T {
	out := make(chan T)

	var wg sync.WaitGroup

	for _, in := range ins {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case val, ok := <-in:
					if !ok {
						return
					}

					select {
					case <-ctx.Done():
						return
					case out <- val:
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
