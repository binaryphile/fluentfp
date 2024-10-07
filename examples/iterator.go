//go:build iterator

package main

import (
	"fmt"
	"github.com/binaryphile/fluentfp/iterator"
)

func main() {
	next := iterator.FromSlice([]int{0, 1, 2})

	for i, ok := next(); ok; i, ok = next() {
		fmt.Println("next integer is", i)
	}
}
