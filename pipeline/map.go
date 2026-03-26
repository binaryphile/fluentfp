package pipeline

import (
	"context"
	"runtime/debug"
	"sync"

	"github.com/binaryphile/fluentfp/call"
	"github.com/binaryphile/fluentfp/rslt"
)

type workItem[T any] struct {
	seq uint64
	val T
}

type workResult[R any] struct {
	seq uint64
	res rslt.Result[R]
}

// Map applies fn to each input using workers persistent goroutines (pull model).
// Output order matches input order via a reorder buffer.
// Workers pull from input — blocked workers create natural backpressure.
// Panics in fn are recovered as *[rslt.PanicError].
// Panics if workers <= 0. fn must not be nil.
func FanOut[T, R any](ctx context.Context, in <-chan T, workers int, fn call.Func[T, R]) <-chan rslt.Result[R] {
	if workers <= 0 {
		panic("pipeline.FanOut: workers must be > 0")
	}

	out := make(chan rslt.Result[R])
	workCh := make(chan workItem[T])
	collectCh := make(chan workResult[R], workers)

	// Dispatcher: read input, assign sequence numbers, send to workers.
	go func() {
		defer close(workCh)

		var seq uint64

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
				case workCh <- workItem[T]{seq: seq, val: val}:
					seq++
				}
			}
		}
	}()

	// Workers: pull work items, call fn with panic recovery, send results.
	var wg sync.WaitGroup

	for range workers {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for item := range workCh {
				res := runItem(ctx, item.val, fn)

				select {
				case <-ctx.Done():
					return
				case collectCh <- workResult[R]{seq: item.seq, res: res}:
				}
			}
		}()
	}

	// Close collectCh when all workers exit.
	go func() {
		wg.Wait()
		close(collectCh)
	}()

	// Collector: reorder results and emit in sequence order.
	go func() {
		defer close(out)

		var nextSeq uint64
		buf := make(map[uint64]rslt.Result[R])

		for wr := range collectCh {
			buf[wr.seq] = wr.res

			for {
				res, ok := buf[nextSeq]
				if !ok {
					break
				}

				delete(buf, nextSeq)

				select {
				case <-ctx.Done():
					return
				case out <- res:
				}

				nextSeq++
			}
		}

		// Drain remaining buffered items after collectCh closes.
		for {
			res, ok := buf[nextSeq]
			if !ok {
				break
			}

			delete(buf, nextSeq)

			select {
			case <-ctx.Done():
				return
			case out <- res:
			}

			nextSeq++
		}
	}()

	return out
}


// runItem calls fn with panic recovery, returning the result as rslt.Result.
func runItem[T, R any](ctx context.Context, t T, fn func(context.Context, T) (R, error)) (res rslt.Result[R]) {
	defer func() {
		if v := recover(); v != nil {
			res = rslt.Err[R](&rslt.PanicError{Value: v, Stack: debug.Stack()})
		}
	}()

	r, err := fn(ctx, t)
	if err != nil {
		return rslt.Err[R](err)
	}

	return rslt.Ok(r)
}
