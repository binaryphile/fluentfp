//go:build ignore

// Package snippet is the verification harness for the topo showcase
// entry in docs/showcase.md (hashicorp/terraform DAG, separating
// DFS engine from per-algorithm behavior). The showcase fence was
// split into two slots so the generic `func dfs[V comparable]`
// declaration (which can't live as a function value due to Go's
// no-generic-lambdas rule) can sit at package level:
//
//   slot=engine     → dfs generic function (package level)
//   slot=algorithms → Topological sort + Reachability usage (function body)
//
// The algorithms slot lives inside Demo, which returns the
// (sorted, reachable) pair so both result variables are consumed by
// the snippet's own return statement.
//
// The `go:build ignore` constraint excludes this file from default
// `go build ./...`; scripts/check-snippets.py strips the constraint
// when assembling into the tmpdir.
package snippet

import (
	"slices"
)

// Vertex stubs the graph vertex type.
type Vertex int

// Graph stubs the DAG type. Only Neighbors + Vertices are exercised
// by the snippet.
type Graph struct {
	neighbors func(Vertex) []Vertex
	vertices  []Vertex
}

func (g Graph) Neighbors(v Vertex) []Vertex { return g.neighbors(v) }
func (g Graph) Vertices() []Vertex          { return g.vertices }

// __SNIPPET_engine__

func Demo(graph Graph, source Vertex) ([]Vertex, []Vertex) {
	// __SNIPPET_algorithms__
}

// Force-reference slices so the un-substituted harness parses with
// imports used. The algorithms slot exercises slices.Reverse.
var _ = slices.Reverse[[]int]
