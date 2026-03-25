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
