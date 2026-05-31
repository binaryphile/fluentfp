//go:build ignore

// Package snippet is the verification harness for the pmap showcase
// entry in docs/showcase.md (Starship-style parallel module rendering).
// The showcase has FOUR fluentfp fences mapping to four slots:
//
//   slot=extracted  → Segment type + renderModule function (package level)
//   slot=sequential → RenderSequential body (slice.Map call + return)
//   slot=parallel   → RenderParallel body (slice.PMap call + return)
//   slot=method     → ActiveModules body (isEnabled closure + PKeepIf + return)
//
// The extracted slot sits at package level so the function-body slots
// can reference Segment and renderModule directly.
//
// The `go:build ignore` constraint excludes this file from default
// `go build ./...`; scripts/check-snippets.py strips the constraint
// when assembling into the tmpdir.
package snippet

import (
	"os"
	"os/exec"
	"strings"

	"github.com/binaryphile/fluentfp/slice"
)

// __SNIPPET_extracted__

func RenderSequential(enabledModules []string) []Segment {
	// __SNIPPET_sequential__
}

func RenderParallel(enabledModules []string) []Segment {
	// __SNIPPET_parallel__
}

func ActiveModules(allModules []string) []string {
	// __SNIPPET_method__
}

// Force-reference each import so the un-substituted harness parses
// with imports used. After substitution every import is exercised by
// the respective slot.
var _ = os.Stat
var _ = exec.Command
var _ = strings.TrimSpace
var _ = slice.From[string]
