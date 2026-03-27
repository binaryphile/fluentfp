package seq_test

import (
	"fmt"
	"strconv"

	"github.com/binaryphile/fluentfp/seq"
)

func ExampleSeq_pipeline() {
	// double multiplies n by 2.
	double := func(n int) int { return n * 2 }

	// Double each element, take the first 3, and collect.
	result := seq.Map(seq.From([]int{1, 2, 3, 4, 5}), double).
		Take(3).
		Collect()
	fmt.Println(result)
	// Output: [2 4 6]
}

func ExampleGenerate() {
	// doubleIt multiplies n by 2.
	doubleIt := func(n int) int { return n * 2 }

	// Infinite sequence of powers of 2, take first 6.
	powers := seq.Generate(1, doubleIt).Take(6).Collect()
	fmt.Println(powers)
	// Output: [1 2 4 8 16 32]
}

func ExampleFold() {
	// add sums two integers.
	add := func(acc, n int) int { return acc + n }

	total := seq.Fold(seq.From([]int{1, 2, 3, 4, 5}), 0, add)
	fmt.Println(total)
	// Output: 15
}

func ExampleFilterMap() {
	// tryParseInt parses a string as an integer, reporting success via the bool.
	tryParseInt := func(s string) (int, bool) {
		n, err := strconv.Atoi(s)
		return n, err == nil
	}

	nums := seq.FilterMap(seq.From([]string{"1", "abc", "3", "5"}), tryParseInt).Collect()
	fmt.Println(nums)
	// Output: [1 3 5]
}

func ExampleUnfold() {
	// nextFib produces the next Fibonacci number from (a, b) state.
	nextFib := func(state [2]int) (int, [2]int, bool) {
		a, b := state[0], state[1]
		return a, [2]int{b, a + b}, true
	}

	fibs := seq.Unfold([2]int{0, 1}, nextFib).Take(8).Collect()
	fmt.Println(fibs)
	// Output: [0 1 1 2 3 5 8 13]
}
