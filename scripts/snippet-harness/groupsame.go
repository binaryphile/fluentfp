//go:build ignore

// Package snippet is the verification harness for the groupsame showcase
// entry in docs/showcase.md. The snippet body lands at __SNIPPET__ via
// scripts/check-snippets.py; the marker line is replaced verbatim.
//
// The `go:build ignore` constraint excludes this file from default
// `go build ./...` (which would fail on the incomplete function body and
// unused-until-substitution imports). scripts/check-snippets.py strips
// the constraint when assembling into the tmpdir, so the verification
// path sees a buildable file.
package snippet

import (
	"fmt"

	"github.com/binaryphile/fluentfp/slice"
)

// CombinedStatus is a function shell whose body is the groupsame snippet.
// The snippet is a complete function body (declares its own intermediates
// and ends with its own `return` statement). The harness supplies only the
// surrounding signature + imports; there is no implicit variable-name
// contract between snippet and scaffold.
//
// Imports above must each be used by the snippet (Go bans unused imports).
// The groupsame snippet uses fmt.Sprintf in countByStatus and slice.* in
// the chain — both imports are exercised by the snippet itself.
func CombinedStatus(statuses []string) string {
	// __SNIPPET__
}
