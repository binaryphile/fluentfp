package heap_test

import (
	"cmp"
	"fmt"

	"github.com/binaryphile/fluentfp/heap"
)

func ExampleNew() {
	// Priority queue: smallest number has highest priority.
	h := heap.New[int](cmp.Compare)

	h = h.Insert(3).Insert(1).Insert(2)
	fmt.Println(h.Min().Or(0))
	// Output: 1
}

func ExampleFrom() {
	// Build a heap from a slice.
	h := heap.From([]int{5, 3, 1, 4, 2}, cmp.Compare)
	fmt.Println(h.Collect())
	// Output: [1 2 3 4 5]
}

func ExampleHeap_Pop() {
	// Pop returns the minimum, the remaining heap, and true.
	h := heap.From([]int{3, 1, 2}, cmp.Compare)

	min, rest, ok := h.Pop()
	fmt.Printf("min=%d ok=%t remaining=%d\n", min, ok, rest.Len())
	// Output: min=1 ok=true remaining=2
}

func ExampleHeap_Merge() {
	// Merge two heaps in O(1).
	a := heap.From([]int{1, 3, 5}, cmp.Compare)
	b := heap.From([]int{2, 4, 6}, cmp.Compare)

	merged := a.Merge(b)
	fmt.Println(merged.Collect())
	// Output: [1 2 3 4 5 6]
}

func ExampleHeap_Collect() {
	// Collect drains the heap in sorted order.
	h := heap.From([]string{"banana", "apple", "cherry"}, cmp.Compare)
	fmt.Println(h.Collect())
	// Output: [apple banana cherry]
}
