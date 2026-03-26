package combo_test

import (
	"fmt"

	"github.com/binaryphile/fluentfp/combo"
)

func ExampleCombinations() {
	// Choose 2 items from 4.
	results := combo.Combinations([]string{"a", "b", "c", "d"}, 2)

	for _, c := range results {
		fmt.Println(c)
	}
	// Output:
	// [a b]
	// [a c]
	// [a d]
	// [b c]
	// [b d]
	// [c d]
}

func ExamplePermutations() {
	// All orderings of 3 items.
	results := combo.Permutations([]int{1, 2, 3})
	fmt.Println(results.Len(), "permutations")
	fmt.Println(results[0])
	// Output:
	// 6 permutations
	// [1 2 3]
}

func ExampleCartesianProduct() {
	// All pairs from two lists.
	results := combo.CartesianProduct([]string{"a", "b"}, []int{1, 2})

	for _, p := range results {
		fmt.Printf("(%s, %d)\n", p.First, p.Second)
	}
	// Output:
	// (a, 1)
	// (a, 2)
	// (b, 1)
	// (b, 2)
}

func ExamplePowerSet() {
	// All subsets of a set.
	results := combo.PowerSet([]string{"x", "y"})

	for _, s := range results {
		fmt.Println(s)
	}
	// Output:
	// []
	// [y]
	// [x]
	// [x y]
}
