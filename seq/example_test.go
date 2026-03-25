package seq_test

import (
	"fmt"

	"github.com/binaryphile/fluentfp/seq"
)

func ExampleSeq_pipeline() {
	// double multiplies n by 2.
	double := func(n int) int { return n * 2 }

	// Generate an infinite sequence of naturals starting at 1,
	// take the first 5, double each, and collect.
	result := seq.Map(seq.From([]int{1, 2, 3, 4, 5}), double).
		Take(3).
		Collect()
	fmt.Println(result)
	// Output: [2 4 6]
}
