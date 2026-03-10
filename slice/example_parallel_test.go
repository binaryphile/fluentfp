package slice_test

import (
	"fmt"
	"runtime"

	"github.com/binaryphile/fluentfp/slice"
)

func ExamplePMap() {
	double := func(n int) int { return n * 2 }
	result := slice.PMap(slice.From([]int{1, 2, 3, 4, 5}), runtime.GOMAXPROCS(0), double)
	fmt.Println([]int(result))
	// Output: [2 4 6 8 10]
}

func ExampleMapper_PKeepIf() {
	isEven := func(n int) bool { return n%2 == 0 }
	result := slice.From([]int{1, 2, 3, 4, 5, 6}).PKeepIf(4, isEven)
	fmt.Println([]int(result))
	// Output: [2 4 6]
}

func ExampleMapper_PEach() {
	slice.From([]string{"a", "b", "c"}).PEach(2, func(s string) {
		// process each element concurrently
		_ = s
	})
}
