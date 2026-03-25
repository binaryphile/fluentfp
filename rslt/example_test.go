package rslt_test

import (
	"fmt"
	"strconv"

	"github.com/binaryphile/fluentfp/rslt"
)

func Example() {
	// Parse a string to int, double it, format back to string.
	// Errors propagate automatically — no if-err checks.
	result := rslt.Of(strconv.Atoi("21")).
		Transform(func(n int) int { return n * 2 })

	fmt.Println(result.Or(0))
	// Output: 42
}

func Example_errorPropagation() {
	// When the first step fails, Transform is skipped.
	result := rslt.Of(strconv.Atoi("not a number")).
		Transform(func(n int) int { return n * 2 })

	fmt.Println("ok:", result.IsOk())
	// Output: ok: false
}
