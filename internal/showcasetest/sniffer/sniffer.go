// Package sniffer compile-checks the showcase entry for chenjiandongx/sniffer.
// The fluentfp rewrite below is the same code shown in docs/showcase.md;
// stubbing the project-specific types means a method/signature change in
// fluentfp will break this build.
package sniffer

import (
	"github.com/binaryphile/fluentfp/kv"
	"github.com/binaryphile/fluentfp/slice"
)

// --- stubs for the sniffer project's types ---

type ViewMode int

const (
	ModeTableBytes ViewMode = iota
	ModeTablePackets
)

type ProcessesResult struct {
	bytes, packets int
}

func (p ProcessesResult) TotalBytes() int   { return p.bytes }
func (p ProcessesResult) TotalPackets() int { return p.packets }

type Snapshot struct {
	Processes map[string]int
}

// NewResult matches the constructor signature kv.Map expects: func(K, V) R.
func NewResult(k string, v int) ProcessesResult { return ProcessesResult{} }

// --- the fluentfp rewrite from docs/showcase.md (verbatim) ---

var sortFuncs = map[ViewMode]func(ProcessesResult) int{
	ModeTableBytes:   ProcessesResult.TotalBytes,
	ModeTablePackets: ProcessesResult.TotalPackets,
}

// TopNProcesses wraps the showcase chain in a callable function so the
// package has a discoverable entry point.
func (s *Snapshot) TopNProcesses(n int, mode ViewMode) []ProcessesResult {
	byViewModeDesc := slice.Desc(sortFuncs[mode]) // slice.Desc creates a comparator for .Sort
	results := kv.Map(s.Processes, NewResult).Sort(byViewModeDesc).Take(n)
	return results
}
