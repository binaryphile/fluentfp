package stream_test

import (
	"fmt"

	"github.com/binaryphile/fluentfp/stream"
)

func ExampleStream_infinite() {
	// addOne increments n by 1.
	addOne := func(n int) int { return n + 1 }

	// Generate an infinite stream starting at 1, take the first 5 elements.
	naturals := stream.Generate(1, addOne)
	first5 := naturals.Take(5).Collect()
	fmt.Println(first5)
	// Output: [1 2 3 4 5]
}

func ExampleFrom() {
	// isEven reports whether n is divisible by 2.
	isEven := func(n int) bool { return n%2 == 0 }

	evens := stream.From([]int{1, 2, 3, 4, 5}).KeepIf(isEven).Collect()
	fmt.Println(evens)
	// Output: [2 4]
}

func ExampleMap() {
	// label formats an int as a labeled string.
	label := func(n int) string { return fmt.Sprintf("item-%d", n) }

	labels := stream.Map(stream.From([]int{1, 2, 3}), label).Collect()
	fmt.Println(labels)
	// Output: [item-1 item-2 item-3]
}

func ExampleConcat() {
	a := stream.From([]int{1, 2})
	b := stream.From([]int{3, 4})

	combined := stream.Concat(a, b).Collect()
	fmt.Println(combined)
	// Output: [1 2 3 4]
}
