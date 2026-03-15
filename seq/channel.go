package seq

import "context"

// FromChannel creates a Seq that yields values received from ch.
// Iteration blocks on each receive. The sequence ends when ch is closed.
//
// Cancellation is best-effort: if cancellation races with ready channel
// receives, cancellation does not necessarily win immediately. The
// sequence may yield additional values until the ctx.Done() case is
// selected by Go's pseudo-random select.
//
// The provided ctx is captured in the returned Seq for its lifetime.
// Unlike most Go APIs, cancellation scope is fixed at construction time,
// not iteration time.
//
// Like [FromNext], the returned Seq is stateful — re-iteration continues
// from whatever channel state exists, not from the beginning.
//
// Panics if ctx or ch is nil.
func FromChannel[T any](ctx context.Context, ch <-chan T) Seq[T] {
	if ctx == nil {
		panic("seq.FromChannel: ctx must not be nil")
	}

	if ch == nil {
		panic("seq.FromChannel: ch must not be nil")
	}

	return Seq[T](func(yield func(T) bool) {
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-ch:
				if !ok {
					return
				}

				if !yield(v) {
					return
				}
			}
		}
	})
}

// ToChannel sends values from s into a new channel, returning it.
// A goroutine is spawned to drive iteration; it closes the returned
// channel when iteration ends.
//
// Cancellation is cooperative: the goroutine checks ctx at each
// yield/send boundary. If s blocks internally before yielding the
// next value (e.g., a nested [FromChannel] on a slow source), ctx
// cancellation cannot interrupt it — the goroutine remains blocked
// until s yields or terminates. The caller must drain or cancel to
// avoid goroutine leaks. If both ctx.Done() and the channel send are
// selectable, sends may continue until the cancellation branch is
// selected. Buffered channels or an actively receiving consumer can
// therefore observe additional post-cancel values.
//
// If ctx is already canceled, no goroutine is spawned and a closed
// empty channel is returned.
//
// If s is nil (zero value), returns a closed empty channel.
// buf sets the channel buffer size (0 for unbuffered).
// Panics if ctx is nil or buf < 0.
//
// ToChannel assumes s follows the [iter.Seq] protocol: no concurrent
// yield calls, no retaining yield after return.
func (s Seq[T]) ToChannel(ctx context.Context, buf int) <-chan T {
	if ctx == nil {
		panic("seq.ToChannel: ctx must not be nil")
	}

	if buf < 0 {
		panic("seq.ToChannel: buf must be >= 0")
	}

	out := make(chan T, buf)

	if s == nil {
		close(out)

		return out
	}

	select {
	case <-ctx.Done():
		close(out)

		return out
	default:
	}

	go func() {
		defer close(out)

		s(func(v T) bool {
			select {
			case <-ctx.Done():
				return false
			case out <- v:
				return true
			}
		})
	}()

	return out
}
