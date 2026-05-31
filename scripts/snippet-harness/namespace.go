//go:build ignore

// Package snippet is the verification harness for the namespace showcase
// entry in docs/showcase.md (kubernetes/client-go inClusterClientConfig
// Namespace rewrite). The snippet declares a package-level const,
// two helper vars, and a method on *inClusterClientConfig — all at
// package scope, so the substitution is at package level with no
// function wrapper.
//
// The `go:build ignore` constraint excludes this file from default
// `go build ./...`; scripts/check-snippets.py strips the constraint
// when assembling into the tmpdir.
package snippet

import (
	"os"
	"strings"

	"github.com/binaryphile/fluentfp/option"
)

// inClusterClientConfig stubs kubernetes/client-go's type — only its
// identity is used by the snippet's method declaration.
type inClusterClientConfig struct{}

// __SNIPPET__
