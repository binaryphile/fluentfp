package slice_test

import (
	"fmt"

	"github.com/binaryphile/fluentfp/slice"
)

func ExampleMapper_KeyByString() {
	type user struct {
		name string
		age  int
	}

	// getName extracts the name field as the map key.
	getName := func(u user) string { return u.name }

	byName := slice.From([]user{{"alice", 30}, {"bob", 25}}).KeyByString(getName)

	fmt.Println(byName["alice"].age, byName["bob"].age)
	// Output: 30 25
}
