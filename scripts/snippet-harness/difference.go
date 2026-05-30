//go:build ignore

// Package snippet is the verification harness for the difference showcase
// entry in docs/showcase.md (hashicorp/go-secure-stdlib strutil.Difference
// rewrite). The snippet is a complete top-level function declaration, so
// the substitution is at package scope with no wrapper.
//
// The `go:build ignore` constraint excludes this file from default
// `go build ./...`; scripts/check-snippets.py strips the constraint
// when assembling into the tmpdir.
package snippet

import (
	"strings"

	"github.com/binaryphile/fluentfp/hof"
	"github.com/binaryphile/fluentfp/lof"
	"github.com/binaryphile/fluentfp/option"
	"github.com/binaryphile/fluentfp/slice"
)

// __SNIPPET__
