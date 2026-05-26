// Package topo compile-checks the showcase entry for hashicorp/terraform DAG.
// This entry doesn't use fluentfp APIs directly — it demonstrates engine/algorithm
// decomposition in pure Go. Compile-check verifies the Go is valid.
package topo

import (
	"slices"
)

// --- stubs for the graph types ---

type Vertex int

type Graph struct {
	neighbors func(Vertex) []Vertex
	vertices  []Vertex
}

func (g Graph) Neighbors(v Vertex) []Vertex { return g.neighbors(v) }
func (g Graph) Vertices() []Vertex          { return g.vertices }

// --- the fluentfp rewrite from docs/showcase.md (verbatim, in a function so vars are reachable) ---

// dfs traverses all vertices depth-first, calling arrive on entry and depart on exit.
// The engine is reusable; the algorithm lives in arrive and depart.
func dfs[V comparable](
	neighbors func(V) []V,
	arrive func(V, []V) []V,
	depart func(V, []V) []V,
) func(vertices []V) []V {
	return func(vertices []V) []V {
		visited := make(map[V]bool)
		acc := []V(nil)

		var visit func(V)
		visit = func(v V) {
			if visited[v] {
				return
			}
			visited[v] = true
			acc = arrive(v, acc)
			for _, u := range neighbors(v) {
				visit(u)
			}
			acc = depart(v, acc)
		}

		for _, v := range vertices {
			visit(v)
		}
		return acc
	}
}

func TopologicalSort(graph Graph) []Vertex {
	// Topological sort: ignore on arrive, collect on depart, reverse.
	noop := func(_ Vertex, acc []Vertex) []Vertex { return acc }
	collect := func(v Vertex, acc []Vertex) []Vertex { return append(acc, v) }
	topoSort := dfs(graph.Neighbors, noop, collect)
	sorted := topoSort(graph.Vertices())
	slices.Reverse(sorted)
	return sorted
}

func Reachable(graph Graph, source Vertex) []Vertex {
	collect := func(v Vertex, acc []Vertex) []Vertex { return append(acc, v) }
	noop := func(_ Vertex, acc []Vertex) []Vertex { return acc }
	// Reachability from a single source: collect on arrive, ignore on depart.
	reachFrom := dfs(graph.Neighbors, collect, noop)
	return reachFrom([]Vertex{source})
}
