package option_test

import (
	"fmt"

	"github.com/binaryphile/fluentfp/option"
)

func ExampleWhen() {
	overdue := true
	level := option.When(overdue, "critical").Or("info")
	fmt.Println(level)

	overdue = false
	level = option.When(overdue, "critical").Or("info")
	fmt.Println(level)
	// Output:
	// critical
	// info
}

func ExampleWhenCall() {
	// fetchConfig is only called when needsFetch is true.
	fetchConfig := func() string { return "fetched" }

	result := option.WhenCall(true, fetchConfig).Or("default")
	fmt.Println(result)

	result = option.WhenCall(false, fetchConfig).Or("default")
	fmt.Println(result)
	// Output:
	// fetched
	// default
}
