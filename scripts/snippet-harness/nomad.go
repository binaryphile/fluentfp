//go:build ignore

// Package snippet is the verification harness for the nomad showcase
// entry in docs/showcase.md (hashicorp/nomad agent config Merge rewrite).
//
// The Config stub carries the three representative fields shown in the
// snippet (the real nomad Config has ~48). The snippet is just the
// `return Config{...}` body — the harness supplies the function shell
// so the snippet stays focused on the win (struct-literal return,
// no pre-construction variable, no post-construction overrides).
//
// The `go:build ignore` constraint excludes this file from default
// `go build ./...`; scripts/check-snippets.py strips the constraint
// when assembling into the tmpdir.
package snippet

import (
	"cmp"

	"github.com/binaryphile/fluentfp/option"
)

// Config stubs hashicorp/nomad/command/agent.Config — only the three
// fields shown in the showcase snippet are exercised.
type Config struct {
	AuthoritativeRegion string
	BootstrapExpect     int
	RaftProtocol        int
}

func Merge(s, b Config) Config {
	// __SNIPPET__
}
