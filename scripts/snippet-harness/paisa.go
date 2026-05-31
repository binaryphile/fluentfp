//go:build ignore

// Package snippet is the verification harness for the paisa showcase
// entry in docs/showcase.md (ananthakumaran/paisa tf_idf tokenize
// rewrite). The fluentfp snippet is a one-line chain inside a
// function body; the harness wraps it in Tokenize and supplies
// the splitTokens helper that the snippet references.
//
// The `go:build ignore` constraint excludes this file from default
// `go build ./...`; scripts/check-snippets.py strips the constraint
// when assembling into the tmpdir.
package snippet

import (
	"regexp"
	"strings"

	"github.com/binaryphile/fluentfp/lof"
	"github.com/binaryphile/fluentfp/slice"
)

// splitTokens is shared between the original and the rewrite (the
// extracted-helper block in the showcase). The harness keeps it at
// package level so the snippet's chain has a `splitTokens` symbol
// to call.
func splitTokens(s string) slice.Mapper[string] {
	return regexp.MustCompile("[ .()/:]+").Split(s, -1)
}

func Tokenize(s string) []string {
	// __SNIPPET__
}
