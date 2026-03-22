package toc

// Pipeline is a DAG topology descriptor for a toc pipeline. Configure
// with [Pipeline.AddStage] and [Pipeline.AddEdge], then call
// [Pipeline.Freeze]. After Freeze, the Pipeline is immutable and safe
// for concurrent reads.
//
// Pipeline does NOT own or create stages. It stores stage names and
// stats accessors only — a passive metadata layer over existing stages.
// Consumers use target-relative queries ([Pipeline.AncestorsOf],
// [Pipeline.HeadsTo], [Pipeline.DirectPredecessors]) to reason about
// upstream subgraphs relative to a chosen drum.
//
// Panics on misconfiguration (empty names, nil stats, cycles, etc.)
// following the toc convention: configuration errors are programming
// bugs, caught on first run during development.
type Pipeline struct {
	frozen  bool
	stages  map[string]pipelineEntry
	order   []string // registration order
	edges   []pipelineEdge
	forward map[string][]string // from -> []to (built on Freeze)
	reverse map[string][]string // to -> []from (built on Freeze)
	heads   []string            // zero in-degree (built on Freeze)
}

type pipelineEntry struct {
	stats func() Stats
}

type pipelineEdge struct {
	from  string
	to    string
	ratio int // consumption ratio: items from 'from' consumed per unit of output at 'to'. Always >= 1.
}

// NewPipeline creates an empty pipeline topology.
func NewPipeline() *Pipeline {
	return &Pipeline{
		stages: make(map[string]pipelineEntry),
	}
}

// AddStage registers a named stage with its stats accessor.
// Panics if name is empty, stats is nil, name is duplicate, or
// the pipeline is frozen.
func (p *Pipeline) AddStage(name string, stats func() Stats) {
	p.mustNotFrozen()
	if name == "" {
		panic("toc.Pipeline: stage name must not be empty")
	}
	if stats == nil {
		panic("toc.Pipeline: stats must not be nil")
	}
	if _, exists := p.stages[name]; exists {
		panic("toc.Pipeline: duplicate stage name: " + name)
	}

	p.stages[name] = pipelineEntry{stats: stats}
	p.order = append(p.order, name)
}

// AddEdge registers a directed edge from → to with a 1:1 consumption
// ratio. Both stages must already be registered. Panics if from or to
// is unknown, the edge is a duplicate, from == to (self-loop), or the
// pipeline is frozen.
func (p *Pipeline) AddEdge(from, to string) {
	p.addEdge(from, to, 1)
}

// AddEdgeWithRatio registers a directed edge with a consumption ratio.
// The ratio specifies how many items from 'from' are consumed to produce
// one unit of output at 'to' — the Bill of Materials ratio for this edge.
// For example, ratio=2 means 'to' consumes 2 items from 'from' per output.
// The rope uses this to release inputs at the correct ratio to
// maximize goodput at merge points.
//
// Panics if ratio <= 0, from or to is unknown, the edge is a duplicate,
// from == to (self-loop), or the pipeline is frozen.
func (p *Pipeline) AddEdgeWithRatio(from, to string, ratio int) {
	if ratio <= 0 {
		panic("toc.Pipeline: ratio must be positive")
	}
	p.addEdge(from, to, ratio)
}

func (p *Pipeline) addEdge(from, to string, ratio int) {
	p.mustNotFrozen()
	if _, ok := p.stages[from]; !ok {
		panic("toc.Pipeline: unknown stage in edge 'from': " + from)
	}
	if _, ok := p.stages[to]; !ok {
		panic("toc.Pipeline: unknown stage in edge 'to': " + to)
	}
	if from == to {
		panic("toc.Pipeline: self-loop: " + from)
	}

	for _, e := range p.edges {
		if e.from == from && e.to == to {
			panic("toc.Pipeline: duplicate edge: " + from + " → " + to)
		}
	}

	p.edges = append(p.edges, pipelineEdge{from: from, to: to, ratio: ratio})
}

// Freeze validates the topology and makes the pipeline read-only.
// Builds adjacency lists and computes heads (zero in-degree stages).
// Panics if the graph contains a cycle, has no stages, or Freeze
// was already called.
func (p *Pipeline) Freeze() {
	p.mustNotFrozen()
	if len(p.stages) == 0 {
		panic("toc.Pipeline: no stages registered")
	}

	// Build adjacency.
	p.forward = make(map[string][]string, len(p.stages))
	p.reverse = make(map[string][]string, len(p.stages))
	for _, name := range p.order {
		p.forward[name] = nil
		p.reverse[name] = nil
	}
	for _, e := range p.edges {
		p.forward[e.from] = append(p.forward[e.from], e.to)
		p.reverse[e.to] = append(p.reverse[e.to], e.from)
	}

	// Cycle detection via Kahn's algorithm.
	inDegree := make(map[string]int, len(p.stages))
	for _, name := range p.order {
		inDegree[name] = len(p.reverse[name])
	}

	queue := make([]string, 0, len(p.stages))
	for _, name := range p.order {
		if inDegree[name] == 0 {
			queue = append(queue, name)
		}
	}

	sorted := 0
	for i := 0; i < len(queue); i++ {
		sorted++
		for _, next := range p.forward[queue[i]] {
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}

	if sorted != len(p.stages) {
		panic("toc.Pipeline: cycle detected")
	}

	// Heads: zero in-degree stages, in registration order.
	for _, name := range p.order {
		if len(p.reverse[name]) == 0 {
			p.heads = append(p.heads, name)
		}
	}

	p.frozen = true
}

// Stages returns stage names in registration order.
// Panics if not frozen.
func (p *Pipeline) Stages() []string {
	p.mustFrozen()
	result := make([]string, len(p.order))
	copy(result, p.order)
	return result
}

// Heads returns all zero in-degree stages (pipeline entry points),
// in registration order. Panics if not frozen.
func (p *Pipeline) Heads() []string {
	p.mustFrozen()
	result := make([]string, len(p.heads))
	copy(result, p.heads)
	return result
}

// HeadsTo returns the subset of heads that can reach the target stage.
// If the pipeline has multiple entry points but only some feed the
// target (drum), only those heads are returned. Uses a single reverse
// BFS from target, then intersects with heads. Panics if target is
// unknown or pipeline is not frozen.
func (p *Pipeline) HeadsTo(target string) []string {
	p.mustFrozen()
	p.mustStage(target)

	// Reverse BFS from target to find all ancestors.
	reachable := make(map[string]bool, len(p.stages))
	reachable[target] = true
	queue := []string{target}
	for i := 0; i < len(queue); i++ {
		for _, pred := range p.reverse[queue[i]] {
			if !reachable[pred] {
				reachable[pred] = true
				queue = append(queue, pred)
			}
		}
	}

	// Intersect with heads, preserving registration order.
	var result []string
	for _, h := range p.heads {
		if reachable[h] {
			result = append(result, h)
		}
	}
	return result
}

// AncestorsOf returns all stages transitively upstream of the target,
// in BFS order (closest first). Excludes the target itself.
// Panics if name is unknown or pipeline is not frozen.
func (p *Pipeline) AncestorsOf(target string) []string {
	p.mustFrozen()
	p.mustStage(target)

	visited := make(map[string]bool, len(p.stages))
	visited[target] = true

	queue := make([]string, 0, len(p.reverse[target]))
	for _, pred := range p.reverse[target] {
		if !visited[pred] {
			visited[pred] = true
			queue = append(queue, pred)
		}
	}

	var result []string
	for i := 0; i < len(queue); i++ {
		current := queue[i]
		result = append(result, current)
		for _, pred := range p.reverse[current] {
			if !visited[pred] {
				visited[pred] = true
				queue = append(queue, pred)
			}
		}
	}

	return result
}

// DirectPredecessors returns the immediate upstream stages of the
// named stage. Panics if name is unknown or pipeline is not frozen.
func (p *Pipeline) DirectPredecessors(name string) []string {
	p.mustFrozen()
	p.mustStage(name)

	preds := p.reverse[name]
	if len(preds) == 0 {
		return nil
	}
	result := make([]string, len(preds))
	copy(result, preds)
	return result
}

// IncomingEdge describes one incoming edge to a stage.
type IncomingEdge struct {
	From  string // predecessor stage name
	Ratio int    // consumption ratio (items from From per unit of output)
}

// Incoming returns the incoming edges to the named stage, each with
// its predecessor name and consumption ratio. Phase 3 uses this to
// compute per-predecessor budget allocation at merge points.
// Panics if name is unknown or pipeline is not frozen.
func (p *Pipeline) Incoming(name string) []IncomingEdge {
	p.mustFrozen()
	p.mustStage(name)

	preds := p.reverse[name]
	if len(preds) == 0 {
		return nil
	}

	result := make([]IncomingEdge, 0, len(preds))
	for _, pred := range preds {
		for _, e := range p.edges {
			if e.from == pred && e.to == name {
				result = append(result, IncomingEdge{From: pred, Ratio: e.ratio})
				break
			}
		}
	}
	return result
}

// HasPath returns true if there is a directed path from → to.
// Panics if either name is unknown or pipeline is not frozen.
func (p *Pipeline) HasPath(from, to string) bool {
	p.mustFrozen()
	p.mustStage(from)
	p.mustStage(to)
	return p.hasPath(from, to)
}

// hasPath is the internal BFS reachability check (no validation).
func (p *Pipeline) hasPath(from, to string) bool {
	if from == to {
		return true
	}

	visited := make(map[string]bool, len(p.stages))
	visited[from] = true

	queue := []string{from}
	for i := 0; i < len(queue); i++ {
		for _, next := range p.forward[queue[i]] {
			if next == to {
				return true
			}
			if !visited[next] {
				visited[next] = true
				queue = append(queue, next)
			}
		}
	}
	return false
}

// StageStats returns the stats accessor for a named stage.
// Panics if name is unknown or pipeline is not frozen.
func (p *Pipeline) StageStats(name string) func() Stats {
	p.mustFrozen()
	p.mustStage(name)
	return p.stages[name].stats
}

// EdgeRatio returns the consumption ratio for the edge from → to.
// Returns 1 for edges added with [Pipeline.AddEdge]. Returns the
// explicit ratio for edges added with [Pipeline.AddEdgeWithRatio].
// Panics if the edge does not exist or pipeline is not frozen.
func (p *Pipeline) EdgeRatio(from, to string) int {
	p.mustFrozen()
	for _, e := range p.edges {
		if e.from == from && e.to == to {
			return e.ratio
		}
	}
	panic("toc.Pipeline: unknown edge: " + from + " → " + to)
}

func (p *Pipeline) mustFrozen() {
	if !p.frozen {
		panic("toc.Pipeline: not frozen")
	}
}

func (p *Pipeline) mustNotFrozen() {
	if p.frozen {
		panic("toc.Pipeline: already frozen")
	}
}

func (p *Pipeline) mustStage(name string) {
	if _, ok := p.stages[name]; !ok {
		panic("toc.Pipeline: unknown stage: " + name)
	}
}
