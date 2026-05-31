//go:build ignore

// Package snippet is the verification harness for the sniffer showcase
// entry in docs/showcase.md (chenjiandongx/sniffer TopNProcesses
// rewrite). The snippet is function-body code (declares a local var
// + two := + return). The harness wraps it in the method shell on
// *Snapshot and stubs the sniffer-specific vocabulary.
//
// The `go:build ignore` constraint excludes this file from default
// `go build ./...`; scripts/check-snippets.py strips the constraint
// when assembling into the tmpdir.
package snippet

import (
	"github.com/binaryphile/fluentfp/kv"
	"github.com/binaryphile/fluentfp/slice"
)

// ViewMode enumerates the sort modes the snippet's sortFuncs table
// keys on.
type ViewMode int

const (
	ModeTableBytes ViewMode = iota
	ModeTablePackets
)

// ProcessesResult stubs the per-process aggregate row. The method
// expressions ProcessesResult.TotalBytes / TotalPackets are used by
// the snippet's sortFuncs map.
type ProcessesResult struct {
	bytes, packets int
}

func (p ProcessesResult) TotalBytes() int   { return p.bytes }
func (p ProcessesResult) TotalPackets() int { return p.packets }

// Snapshot stubs the per-tick statistics snapshot. Only Processes is
// exercised by the snippet.
type Snapshot struct {
	Processes map[string]int
}

// NewResult matches the constructor signature kv.Map expects: func(K, V) R.
func NewResult(k string, v int) ProcessesResult { return ProcessesResult{} }

func (s *Snapshot) TopNProcesses(n int, mode ViewMode) []ProcessesResult {
	// __SNIPPET__
}
