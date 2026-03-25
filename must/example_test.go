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
