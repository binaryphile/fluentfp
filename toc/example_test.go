package toc_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/binaryphile/fluentfp/rslt"
	"github.com/binaryphile/fluentfp/toc"
)

func Example() {
	// double concatenates a string with itself.
	double := func(_ context.Context, s string) (string, error) {
		return s + s, nil
	}

	ctx := context.Background()

	// Start's ctx governs stage lifetime; Submit's ctx bounds only admission.
	stage := toc.Start(ctx, double, toc.Options[string]{Capacity: 3, Workers: 1})

	go func() {
		defer stage.CloseInput()

		for _, item := range []string{"a", "b", "c"} {
			if err := stage.Submit(ctx, item); err != nil {
				break
			}
		}
	}()

	for result := range stage.Out() {
		val, err := result.Unpack()
		if err != nil {
			fmt.Println("error:", err)
			continue
		}
		fmt.Println(val)
	}

	if err := stage.Wait(); err != nil {
		fmt.Println("stage error:", err)
	}

	// Output:
	// aa
	// bb
	// cc
}

// Example_pipeline demonstrates a four-handle pipeline modeled on
// the era-indexer: Start → Batcher → Pipe → Pipe, with error passthrough
// and reverse-order Wait.
func Example_pipeline() {
	ctx := context.Background()

	// Stage 1: "chunk" each number into a string representation.
	chunkFn := func(_ context.Context, n int) (string, error) {
		if n == 3 {
			return "", errors.New("bad input: 3")
		}

		return fmt.Sprintf("chunk(%d)", n), nil
	}
	chunker := toc.Start(ctx, chunkFn, toc.Options[int]{Capacity: 5, ContinueOnError: true})

	// Stage 2: Batch strings into groups of 2.
	batched := toc.NewBatcher(ctx, chunker.Out(), 2)

	// Stage 3: "embed" each batch by joining.
	embedFn := func(_ context.Context, batch []string) (string, error) {
		result := ""
		for i, s := range batch {
			if i > 0 {
				result += "+"
			}
			result += s
		}

		return fmt.Sprintf("embed[%s]", result), nil
	}
	embedder := toc.Pipe(ctx, batched.Out(), embedFn, toc.Options[[]string]{})

	// Stage 4: "store" by uppercasing (identity for this example).
	storeFn := func(_ context.Context, s string) (string, error) {
		return fmt.Sprintf("store(%s)", s), nil
	}
	storer := toc.Pipe(ctx, embedder.Out(), storeFn, toc.Options[string]{})

	// Feed the head stage.
	go func() {
		defer chunker.CloseInput()
		for _, n := range []int{1, 2, 3, 4, 5} {
			if err := chunker.Submit(ctx, n); err != nil {
				break
			}
		}
	}()

	// Drain the tail.
	for r := range storer.Out() {
		if v, err := r.Unpack(); err != nil {
			fmt.Println("error:", err)
		} else {
			fmt.Println(v)
		}
	}

	// Wait in reverse order (recommended).
	storer.Wait()
	embedder.Wait()
	batched.Wait()
	chunker.Wait()

	// Forwarded errors bypass worker queues, so the error from item 3
	// may arrive before the batch containing items 1-2 is processed.

	// Unordered output:
	// error: bad input: 3
	// store(embed[chunk(1)+chunk(2)])
	// store(embed[chunk(4)+chunk(5)])
}

// Example_pipe demonstrates basic error passthrough through a Pipe stage.
func Example_pipe() {
	ctx := context.Background()

	src := make(chan rslt.Result[int], 3)
	src <- rslt.Ok(10)
	src <- rslt.Err[int](errors.New("oops"))
	src <- rslt.Ok(20)
	close(src)

	// doubleFn doubles the input.
	doubleFn := func(_ context.Context, n int) (int, error) {
		return n * 2, nil
	}

	stage := toc.Pipe(ctx, src, doubleFn, toc.Options[int]{})

	for r := range stage.Out() {
		if v, err := r.Unpack(); err != nil {
			fmt.Println("error:", err)
		} else {
			fmt.Println(v)
		}
	}

	stage.Wait()

	// Forwarded errors bypass the worker queue, so the error
	// may arrive before queued Ok results complete.

	// Unordered output:
	// error: oops
	// 20
	// 40
}
