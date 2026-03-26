package slice_test

import (
	"fmt"

	"github.com/binaryphile/fluentfp/slice"
)

func ExampleGroupBy() {
	// firstLetter extracts the first character of a string.
	firstLetter := func(s string) byte { return s[0] }

	groups := slice.GroupBy([]string{"apple", "avocado", "banana", "blueberry"}, firstLetter)
	for _, g := range groups {
		fmt.Printf("%c: %v\n", g.Key, g.Items)
	}
	// Output:
	// a: [apple avocado]
	// b: [banana blueberry]
}

func ExampleAssociate() {
	type person struct {
		name string
		age  int
	}

	people := []person{{"alice", 30}, {"bob", 25}, {"carol", 35}}

	// nameAndAge extracts the name as key and age as value.
	nameAndAge := func(p person) (string, int) { return p.name, p.age }

	m := slice.Associate(people, nameAndAge)
	fmt.Println(m["alice"], m["bob"], m["carol"])
	// Output: 30 25 35
}
