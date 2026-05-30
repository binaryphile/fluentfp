// Package groupsame compile-checks the showcase entry for docker/compose.
package groupsame

import (
	"fmt"

	"github.com/binaryphile/fluentfp/slice"
)

// --- the fluentfp rewrite from docs/showcase.md (verbatim) ---

// slice.Group[K, V] is { Key K; Items []V } with .Len() returning len(Items).
type G = slice.Group[string, string]

// byKey: ascending comparator on Key.
var byKey = slice.Asc(G.GetKey)

// countByStatus: formats one group as "status(count)".
var countByStatus = func(g G) string {
	return fmt.Sprintf("%s(%d)", g.Key, g.Len())
}

func CombinedStatus(statuses []string) string {
	// GroupSame returns one Group per distinct value, where Key == Items[0].
	statusGroups := slice.GroupSame(statuses).Sort(byKey)
	combined := statusGroups.ToString(countByStatus).Join(", ")
	return combined
}
