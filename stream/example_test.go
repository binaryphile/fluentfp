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
