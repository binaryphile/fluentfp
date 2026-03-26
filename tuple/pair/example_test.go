package pair_test

import (
	"fmt"

	"github.com/binaryphile/fluentfp/tuple/pair"
)

func ExampleOf() {
	// Construct a pair of name and age.
	p := pair.Of("alice", 30)
	fmt.Printf("%s is %d\n", p.First, p.Second)
	// Output: alice is 30
}

func ExampleZip() {
	// Pair corresponding elements from two slices.
	names := []string{"alice", "bob", "carol"}
	ages := []int{30, 25, 35}

	pairs := pair.Zip(names, ages)

	for _, p := range pairs {
		fmt.Printf("%s: %d\n", p.First, p.Second)
	}
	// Output:
	// alice: 30
	// bob: 25
	// carol: 35
}

func ExampleZipWith() {
	// Combine two slices with a function.
	names := []string{"alice", "bob"}
	ages := []int{30, 25}

	// label combines name and age into a formatted string.
	label := func(name string, age int) string {
		return fmt.Sprintf("%s(%d)", name, age)
	}

	labels := pair.ZipWith(names, ages, label)
	fmt.Println(labels)
	// Output: [alice(30) bob(25)]
}
