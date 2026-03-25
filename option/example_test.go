package option_test

import (
	"fmt"

	"github.com/binaryphile/fluentfp/option"
)

func ExampleOption_Or() {
	// NonZero wraps a value as ok if it's not the zero value.
	port := option.NonZero(8080).Or(3000)
	fmt.Println(port)

	missing := option.NonZero(0).Or(3000)
	fmt.Println(missing)
	// Output:
	// 8080
	// 3000
}

func ExampleOption_OrElse() {
	// Multi-level fallback: try primary, then secondary, then default.
	primary := option.NotOk[string]()
	secondary := option.Of("backup")

	// primaryLookup returns the primary option.
	primaryLookup := func() option.Option[string] { return primary }
	// secondaryLookup returns the secondary option.
	secondaryLookup := func() option.Option[string] { return secondary }

	result := primaryLookup().
		OrElse(secondaryLookup).
		Or("default")
	fmt.Println(result)
	// Output: backup
}

func ExampleNonZero() {
	fmt.Println(option.NonZero(42).IsOk())
	fmt.Println(option.NonZero(0).IsOk())
	fmt.Println(option.NonZero("hello").IsOk())
	fmt.Println(option.NonZero("").IsOk())
	// Output:
	// true
	// false
	// true
	// false
}
