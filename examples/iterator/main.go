package main

import (
	"fmt"
	"github.com/binaryphile/funcTrunk/iterator"
)

func main() {
	next := iterator.FromSlice([]int{0, 1, 2})
	for i, ok := next(); ok; {
		fmt.Println("next integer is ", i)
		i, ok = next()
	}
}
