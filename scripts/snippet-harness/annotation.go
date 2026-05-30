//go:build ignore

// Package snippet is the verification harness for the annotation showcase
// entry in docs/showcase.md (kubernetes/ttl_controller rewrite). The
// snippet is package-level Go (a var-declared helper + a top-level
// function), so the substitution is at package scope — no function
// wrapper. The `Node` type stub stands in for kubernetes/api/core/v1.Node.
//
// The `go:build ignore` constraint excludes this file from default
// `go build ./...`; scripts/check-snippets.py strips the constraint
// when assembling into the tmpdir.
package snippet

import (
	"strconv"

	"github.com/binaryphile/fluentfp/option"
)

// Node stubs kubernetes/api/core/v1.Node — only the Annotations field
// is exercised by the snippet.
type Node struct {
	Annotations map[string]string
}

// __SNIPPET__
