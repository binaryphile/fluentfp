package slice_test

import (
	"fmt"

	"github.com/binaryphile/fluentfp/slice"
)

func ExampleSortBy() {
	// wordLen returns the length of a string.
	wordLen := func(s string) int { return len(s) }

	sorted := slice.SortBy([]string{"banana", "fig", "apple", "kiwi"}, wordLen)
	fmt.Println(sorted)
	// Output: [fig kiwi apple banana]
}

func ExampleSortByDesc() {
	// wordLen returns the length of a string.
	wordLen := func(s string) int { return len(s) }

	sorted := slice.SortByDesc([]string{"go", "rust", "python"}, wordLen)
	fmt.Println(sorted)
	// Output: [python rust go]
}

func ExampleChunk() {
	chunks := slice.Chunk([]int{1, 2, 3, 4, 5}, 2)
	for _, c := range chunks {
		fmt.Println(c)
	}
	// Output:
	// [1 2]
	// [3 4]
	// [5]
}

func ExampleWindow() {
	windows := slice.Window([]int{10, 20, 30, 40, 50}, 3)
	for _, w := range windows {
		fmt.Println(w)
	}
	// Output:
	// [10 20 30]
	// [20 30 40]
	// [30 40 50]
}

func ExampleZip() {
	names := []string{"alice", "bob", "carol"}
	scores := []int{92, 85, 97}

	pairs := slice.Zip(names, scores)
	for _, p := range pairs {
		fmt.Printf("%s: %d\n", p.First, p.Second)
	}
	// Output:
	// alice: 92
	// bob: 85
	// carol: 97
}

func ExampleZipWith() {
	// add sums two integers.
	add := func(a, b int) int { return a + b }

	sums := slice.ZipWith([]int{1, 2, 3}, []int{10, 20, 30}, add)
	fmt.Println(sums)
	// Output: [11 22 33]
}

func ExampleEnumerate() {
	items := slice.Enumerate([]string{"a", "b", "c"})
	for _, p := range items {
		fmt.Printf("%d: %s\n", p.First, p.Second)
	}
	// Output:
	// 0: a
	// 1: b
	// 2: c
}

func ExampleMapper_Intersperse() {
	path := slice.From([]string{"usr", "local", "bin"}).Intersperse("/")
	fmt.Println(path)
	// Output: [usr / local / bin]
}

func ExampleMapper_Reverse() {
	rev := slice.From([]string{"a", "b", "c"}).Reverse()
	fmt.Println(rev)
	// Output: [c b a]
}

func ExamplePartition() {
	// isEven reports whether n is divisible by 2.
	isEven := func(n int) bool { return n%2 == 0 }

	evens, odds := slice.Partition([]int{1, 2, 3, 4, 5}, isEven)
	fmt.Println(evens)
	fmt.Println(odds)
	// Output:
	// [2 4]
	// [1 3 5]
}

func ExampleMapper_TakeWhile() {
	// isPositive reports whether n is greater than zero.
	isPositive := func(n int) bool { return n > 0 }

	leading := slice.From([]int{3, 1, 4, -1, 5}).TakeWhile(isPositive)
	fmt.Println(leading)
	// Output: [3 1 4]
}

func ExampleRange() {
	fmt.Println(slice.Range(5))
	// Output: [0 1 2 3 4]
}
