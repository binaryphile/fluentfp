package either_test

import (
	"fmt"

	"github.com/binaryphile/fluentfp/either"
)

func ExampleFold() {
	// showError formats an error count for display.
	showError := func(count int) string { return fmt.Sprintf("errors: %d", count) }
	// showValue formats a success value for display.
	showValue := func(v string) string { return fmt.Sprintf("ok: %s", v) }

	success := either.Right[int, string]("hello")
	failure := either.Left[int, string](3)

	fmt.Println(either.Fold(success, showError, showValue))
	fmt.Println(either.Fold(failure, showError, showValue))
	// Output:
	// ok: hello
	// errors: 3
}

func ExampleRight() {
	e := either.Right[string, int](42)
	val, ok := e.Get()
	fmt.Println(val, ok)

	left := either.Left[string, int]("error")
	val, ok = left.Get()
	fmt.Println(val, ok)
	// Output:
	// 42 true
	// 0 false
}

func ExampleMap() {
	// describe formats an int as a descriptive string.
	describe := func(n int) string { return fmt.Sprintf("got %d", n) }

	right := either.Right[string, int](42)
	mapped := either.Map(right, describe)
	fmt.Println(mapped.Or(""))

	left := either.Left[string, int]("fail")
	mapped = either.Map(left, describe)
	fmt.Println(mapped.Or("none"))
	// Output:
	// got 42
	// none
}

func ExampleEither_Swap() {
	e := either.Right[string, int](42)
	swapped := e.Swap()

	// After swap, the int is on the left side.
	leftVal, ok := swapped.GetLeft()
	fmt.Println(leftVal, ok)
	// Output: 42 true
}
