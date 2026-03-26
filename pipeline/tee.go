package pipeline

import "context"

// Tee duplicates input to n output channels. All consumers must keep up —
// the slowest consumer determines throughput.
// Closes all outputs when input closes or ctx cancels.
// Panics if n <= 0.
func Tee[T any](ctx context.Context, in <-chan T, n int) []<-chan T {
	if n <= 0 {
		panic("pipeline.Tee: n must be > 0")
	}

	outs := make([]chan T, n)

	for i := range outs {
		outs[i] = make(chan T)
	}

	go func() {
		defer func() {
			for _, ch := range outs {
				close(ch)
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case val, ok := <-in:
				if !ok {
					return
				}

				for _, ch := range outs {
					select {
					case <-ctx.Done():
						return
					case ch <- val:
					}
				}
			}
		}
	}()

	result := make([]<-chan T, n)

	for i, ch := range outs {
		result[i] = ch
	}

	return result
}
