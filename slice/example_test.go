package slice_test

import (
	"fmt"

	"github.com/binaryphile/fluentfp/slice"
)

func ExampleMapper_KeepIf() {
	// isEven reports whether n is divisible by 2.
	isEven := func(n int) bool { return n%2 == 0 }

	evens := slice.From([]int{1, 2, 3, 4, 5}).KeepIf(isEven)
	fmt.Println(evens)
	// Output: [2 4]
}

func ExampleMap() {
	// label formats an int as a labeled string.
	label := func(n int) string { return fmt.Sprintf("n=%d", n) }

	strs := slice.Map([]int{1, 2, 3}, label)
	fmt.Println(strs)
	// Output: [n=1 n=2 n=3]
}

func ExampleFold() {
	// sum adds two integers.
	sum := func(acc, n int) int { return acc + n }

	total := slice.Fold([]int{1, 2, 3, 4}, 0, sum)
	fmt.Println(total)
	// Output: 10
}
