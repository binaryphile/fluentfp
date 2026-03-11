package slice_test

import (
	"fmt"
	"strings"

	"github.com/binaryphile/fluentfp/slice"
)

func ExampleTally() {
	type G = slice.Group[string, string]

	// formatGroup formats a tally group as "value(count)".
	formatGroup := func(g G) string {
		return fmt.Sprintf("%s(%d)", g.Key, g.Len())
	}

	statuses := slice.Mapper[string]{"running", "exited", "running", "running"}
	formatted := slice.Tally(statuses).ToString(formatGroup)
	fmt.Println(strings.Join(formatted, ", "))
	// Output: running(3), exited(1)
}
