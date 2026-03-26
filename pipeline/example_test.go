package pipeline_test

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/binaryphile/fluentfp/pipeline"
)

func ExampleFanOut() {
	// toUpper converts a string to uppercase.
	toUpper := func(_ context.Context, s string) (string, error) {
		return strings.ToUpper(s), nil
	}

	ctx := context.Background()
	in := pipeline.FromSlice(ctx, []string{"hello", "world", "go"})

	// Process with 2 workers, results in input order.
	out := pipeline.FanOut(ctx, in, 2, toUpper)

	for r := range out {
		fmt.Println(r.Or(""))
	}
	// Output:
	// HELLO
	// WORLD
	// GO
}

func ExampleFilter() {
	ctx := context.Background()
	in := pipeline.FromSlice(ctx, []int{1, 2, 3, 4, 5, 6})

	// isEven keeps only even numbers.
	isEven := func(n int) bool { return n%2 == 0 }

	out := pipeline.Filter(ctx, in, isEven)

	for v := range out {
		fmt.Println(v)
	}
	// Output:
	// 2
	// 4
	// 6
}

func ExampleBatch() {
	ctx := context.Background()
	in := pipeline.FromSlice(ctx, []int{1, 2, 3, 4, 5})

	// Collect into batches of 2. Last batch may be partial.
	out := pipeline.Batch(ctx, in, 2)

	for batch := range out {
		fmt.Println(batch)
	}
	// Output:
	// [1 2]
	// [3 4]
	// [5]
}

func ExampleMerge() {
	ctx := context.Background()
	evens := pipeline.FromSlice(ctx, []int{2, 4, 6})
	odds := pipeline.FromSlice(ctx, []int{1, 3, 5})

	// Combine two streams. Output order is nondeterministic.
	merged := pipeline.Merge(ctx, evens, odds)

	var all []int
	for v := range merged {
		all = append(all, v)
	}

	sort.Ints(all)
	fmt.Println(all)
	// Output: [1 2 3 4 5 6]
}

func ExampleTee() {
	ctx := context.Background()
	in := pipeline.FromSlice(ctx, []string{"a", "b", "c"})

	// Duplicate to 2 consumers. Both see every item.
	outs := pipeline.Tee(ctx, in, 2)

	// Consumer 1: collect items.
	done := make(chan []string, 2)
	for _, out := range outs {
		go func() {
			var items []string
			for v := range out {
				items = append(items, v)
			}
			done <- items
		}()
	}

	first := <-done
	second := <-done
	fmt.Println(first)
	fmt.Println(second)
	// Output:
	// [a b c]
	// [a b c]
}
