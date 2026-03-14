package slice_test

import (
	"fmt"
	"strings"

	"github.com/binaryphile/fluentfp/slice"
)

func ExampleGroupSame() {
	type G = slice.Group[string, string]

	// formatGroup formats a group as "value(count)".
	formatGroup := func(g G) string {
		return fmt.Sprintf("%s(%d)", g.Key, g.Len())
	}

	statuses := slice.Mapper[string]{"running", "exited", "running", "running"}
	formatted := slice.GroupSame(statuses).ToString(formatGroup)
	fmt.Println(strings.Join(formatted, ", "))
	// Output: running(3), exited(1)
}
