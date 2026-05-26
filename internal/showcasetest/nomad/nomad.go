// Package nomad compile-checks the showcase entry for hashicorp/nomad config merge.
package nomad

import (
	"cmp"

	"github.com/binaryphile/fluentfp/option"
)

// --- stub for the Config type ---

type Config struct {
	AuthoritativeRegion string
	BootstrapExpect     int
	RaftProtocol        int
}

// --- the fluentfp rewrite from docs/showcase.md (verbatim, in a function so the vars are reachable) ---

func Merge(s, b Config) Config {
	var result Config
	result.AuthoritativeRegion = cmp.Or(b.AuthoritativeRegion, s.AuthoritativeRegion)
	result.BootstrapExpect = option.When(b.BootstrapExpect > 0, b.BootstrapExpect).Or(s.BootstrapExpect)
	result.RaftProtocol = cmp.Or(b.RaftProtocol, s.RaftProtocol)
	return result
}
