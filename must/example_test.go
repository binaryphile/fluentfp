package must_test

import (
	"fmt"
	"strconv"

	"github.com/binaryphile/fluentfp/must"
	"github.com/binaryphile/fluentfp/slice"
)

func ExampleFrom() {
	// mustAtoi wraps strconv.Atoi to panic on parse error.
	mustAtoi := must.From(strconv.Atoi)

	nums := slice.From([]string{"1", "2", "3"}).ToInt(mustAtoi)
	fmt.Println(nums)
	// Output: [1 2 3]
}

func ExampleGet() {
	// Get unwraps (T, error) or panics — use for known-good invariants.
	n := must.Get(strconv.Atoi("42"))
	fmt.Println(n)
	// Output: 42
}

func ExampleBeNil() {
	// BeNil panics if err is non-nil — use for setup that must succeed.
	var err error
	must.BeNil(err)
	fmt.Println("setup ok")
	// Output: setup ok
}
