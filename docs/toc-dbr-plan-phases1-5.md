# TOC: Drum-Buffer-Rope for Software Pipelines

## What TOC actually says

Goldratt's Theory of Constraints has one central insight: **the system's throughput is determined by its constraint.** Everything else follows from that.

The Five Focusing Steps:
1. **Identify** the constraint
2. **Exploit** it (don't waste it)
3. **Subordinate** everything else to it
4. **Elevate** it (increase its capacity)
5. **If the constraint has moved, go back to Step 1** (prevent inertia)

Drum-Buffer-Rope is the operational policy:
- **Drum**: the constraint's pace. System throughput = drum throughput. Period.
- **Buffer**: work placed in front of the constraint to protect it from starvation. Not a queue everywhere — specifically a queue protecting the drum. Size = drum throughput × upstream variability window. An hour lost at the constraint is an hour lost for the system forever.
- **Rope**: a signal from the drum to the release point. Limits total WIP between release and drum. Prevents upstream overproduction. Rope length = drum throughput × total upstream processing time. The rope PULLS — work is released only when the drum has capacity.

## What a toc package should provide

### The Drum

Automatic identification. The analyzer classifies stages and identifies the bottleneck. This is Step 1, running continuously.

```go
pipeline.Drum() // returns the identified constraint stage
```

### The Buffer

A protective queue in front of the drum. Not every stage gets a buffer — only the one directly upstream of the drum. Its size is computed, not arbitrary:

```
bufferSize = drumThroughput × variabilityWindow
```

If the drum processes 100 items/sec and upstream has 2-second hiccups, the buffer holds 200 items. The buffer's health is monitored via Goldratt's fever chart (green/yellow/red penetration zones).

```go
pipeline.BufferHealth() // green, yellow, red
```

### The Rope

One control: total WIP (by weight) between the release point (pipeline entry) and the drum. Not per-stage. Not a local admission limit. A single global signal that says "the drum has consumed capacity, you may release more."

```go
// At the pipeline head:
pipeline.Submit(ctx, item) // blocks if weight(release→drum) ≥ ropeWeightBudget
```

When the drum completes an item, the rope releases weight at the head. Pull-based scheduling tied to the drum's actual capacity — not backpressure from full queues.

**Rope length formula (yield-adjusted):**
```
requiredReleaseRate = targetGoodput / (1 - errorRate)
ropeWeightBudget = requiredReleaseRate × avgItemWeight × upstreamProcessingTime × safetyFactor
```

The rope is weight-aware because items have variable processing cost. And yield-adjusted because errors at the drum mean we must release more than goodput to sustain it.

**Signals the rope needs:**
- Drum goodput (successful completions/sec)
- Drum error rate (failed / completed)
- Arrival rate (demand entering the pipeline)
- Average item weight at the drum
- Upstream processing time (release→drum)
- Backlog = arrivalRate - goodput (positive = can't keep up)

### Elevation (Step 4)

Move resources to the drum. In software: goroutines, memory budget, priority. The rebalancer does this. When the drum is embed and walk has excess workers, move a worker from walk to embed.

### Inertia Prevention (Step 5)

When elevation succeeds, the drum may move. The old bottleneck is no longer the bottleneck. The system must re-identify — and NOT keep optimizing the old drum. The analyzer handles this via continuous monitoring.

## What changes from current implementation

### Current: per-stage WIP limits called "rope"
**Wrong.** Per-stage admission limits are useful but they are not the rope. They don't reference the drum. They don't limit aggregate WIP from release to drum. They're station-level caps.

### Correct: single aggregate WIP limit from release to drum
The rope is one number: total items allowed between the pipeline entry and the identified drum. When the drum completes, the head releases. This requires:
- Pipeline topology (who is upstream of the drum)
- Aggregate WIP counting across upstream stages
- A release gate at the pipeline head

### Current: arbitrary Capacity per stage
**Insufficient.** Buffer sizing should be relative to the drum, not arbitrary. The buffer that matters is the one directly protecting the drum from starvation.

### Correct: buffer sized by drum throughput × variability
The analyzer measures drum throughput and upstream variability. Buffer size is computed. Dynamic resize would be ideal; for now, the analyzer recommends a size.

### Current: per-stage "rope" Stats (RopeWaitCount, RopeWaitNs)
**Misnamed.** These measure per-stage admission wait. Rename to WIPWaitCount/WIPWaitNs.

### New: pipeline-level rope Stats
- `RopeLength` — current aggregate WIP limit
- `RopeWIP` — current items between release and drum
- `RopeUtilization` — WIP / Length
- `RopeWaitCount` — items that waited at the release gate

## Use Case

### UC-15: Operate a Pipeline Under Drum-Buffer-Rope

**Actor:** Developer configures the pipeline. Operator observes it.

**Main scenario:**
1. Developer creates stages and connects them: `walk → chunk → batch → embed → store`
2. Developer registers stages with the pipeline controller
3. Pipeline starts. Analyzer identifies the drum (embed, saturated at 95% utilization)
4. Rope set: total WIP from walk.Submit to embed.completion = drum throughput × upstream time
5. Buffer health monitored: fever chart shows yellow (45% penetration in front of embed)
6. As items flow: walk.Submit blocks when rope is full. Embed completes → rope releases next item at walk
7. Operator sees: `drum=embed rope=12/12 buffer=yellow(45%) | walk=subordinated chunk=subordinated embed=drum store=subordinated`
8. Developer calls SetWorkers to elevate: move 2 workers from walk to embed
9. Embed throughput increases. Rope adjusts (longer rope = higher drum throughput)
10. Eventually embed is no longer the drum — chunk is. System re-identifies. Rope re-anchors.

**Extensions:**
- Buffer enters red (>66%): drum is about to starve. Upstream problem or rope too tight.
- Buffer enters green (<33%): drum is underutilized. Rope may be too loose or upstream is slow.
- Drum changes: rope detaches from old drum, re-anchors to new drum. Buffer re-positions.
- Memory pressure: memctl.Watch fires, PauseAdmission at release gate. Rope effectively tightened to 0.

## Design Decisions

### D-ROPE: The rope is a pipeline-level pull signal, not per-stage push

Per-stage WIP limits push back locally. The rope PULLS: it signals from the drum to the release point. "I've consumed one, you may release one." This is fundamentally different because:
- It references the drum explicitly
- It limits aggregate WIP, not local WIP
- It adjusts when the drum changes
- It produces predictable pipeline throughput = drum throughput

### D-BUFFER: The buffer is positioned and sized relative to the drum

Only one buffer matters operationally: the one protecting the drum. Other stage buffers are convenience. The critical buffer's size is:
```
size = drumThroughput × max(upstreamProcessingTime, variabilityWindow)
```

### D-RENAME: Per-stage admission limits are "WIP limits," not "rope"

The rope is aggregate. WIP limits are local. Both are useful. Don't confuse them.

### D-TOPOLOGY: The pipeline must know its topology

The rope needs to know which stages are upstream of the drum. The buffer needs to know which stage is directly in front of the drum. The analyzer needs to know what "downstream blocked" means. Topology is not optional for correct DBR.

```go
type Pipeline struct {
    stages []PipelineStage
    edges  []Edge // from → to
}
```

### D-PULL: Submit at the head is gated by drum completions

The release gate at the pipeline head is the rope mechanism:
```go
func (p *Pipeline) Submit(ctx context.Context, item T) error {
    // Wait until aggregate WIP(release→drum) < ropeLength
    // OR ctx is canceled
}
```

This replaces per-stage Submit as the primary admission point for the pipeline.

## Two Constraint Dimensions

The system has two constraint dimensions simultaneously:

### Processing constraint (throughput drum)
Which stage is the bottleneck for throughput? Identified by utilization/idle/blocked classification. The rope for this: total items between release and drum. Elevation: move workers to the drum.

### Memory constraint (resource drum)
Is the system running out of memory? Identified by memctl headroom. The rope for this: total weight (memory cost) between release and the pipeline. This is a SECOND rope — not the same as the processing rope.

**Both ropes operate simultaneously:**
```
processingRope: items(release→processingDrum) < processingRopeLength
memoryRope:     weight(release→pipeline) < memoryBudget
```

An item is released only when BOTH ropes allow it. The tighter rope at any moment determines actual throughput.

**When the constraint moves between dimensions:**
- If processing is the constraint: rope length is set by drum throughput. Workers matter.
- If memory is the constraint: rope length is set by available headroom. Weight matters.
- The system should report WHICH constraint is active — "drum: embed (processing)" or "drum: memory (headroom=340MB)"

This is exactly what the blog post described: the constraint moved from embed throughput to available RAM after we elevated embed with ORT. The rope did the same thing — limited WIP to what the active constraint allows. Same mechanism, different constraint driving its length.

### Buffer for each constraint
- Processing buffer: queue in front of the processing drum. Sized by drum throughput × variability.
- Memory buffer: headroom reserve. Don't use 100% of available memory — keep a buffer. This is the "40% budget factor" from the blog post, but made explicit as a memory buffer.

### Fever chart for each constraint
- Processing fever chart: buffer penetration in front of the processing drum. Green/yellow/red.
- Memory fever chart: headroom consumption. Green (>1GB free) / Yellow (512MB-1GB) / Red (<512MB).

## The Complete Signal Model

| Signal | Formula | Used by |
|--------|---------|---------|
| Drum goodput | (Completed - Failed) / interval | Rope sizing |
| Drum error rate | Failed / Completed | Yield adjustment |
| Arrival rate | Submitted / interval at pipeline head | Load signal |
| Backlog | arrivalRate - goodput | Equilibrium signal |
| Aggregate WIP (release→drum) | sum of admitted across upstream stages | Rope mechanism |
| Aggregate weight (release→drum) | sum of admitted weight across upstream | Weight-aware rope |
| Drum starvation | drum idle AND drum input buffer empty | Step 2 violation |
| Buffer penetration | drum input depth / drum input capacity | Fever chart |
| Throughput-per-worker history | goodput delta after SetWorkers change | Diminishing returns |
| Item weight at drum | Weight(item) at drum stage | Weight-aware rope |
| Memory headroom | memctl.Watch MemAvailable/cgroup | Memory constraint |

## Implementation Plan

### Phase 1: Rename + goodput + starvation signal ✓
- `RopeWaitCount` → `WIPWaitCount` everywhere ✓
- Add goodput to IntervalStats: (Completed - Failed) / elapsed ✓
- Add drum starvation counter to Snapshot (analyzer-level, not per-stage Stats) ✓
- Add arrival rate as named IntervalStats field ✓
- Sticky constraint identity (persists until replaced) ✓
- Drum demo: examples/drum-demo/ ✓

## At Implementation Gate

```bash
evtctl contract '{"phase":"phase1-toc-signals","criteria":["rename RopeWait to WIPWait","add Goodput to IntervalStats","add ArrivalRate to IntervalStats","add DrumStarvationCount to Snapshot","sticky constraint identity","drum demo"]}'
evtctl plan ~/.claude/plans/unified-napping-peach.md
TASK_ID=27841
evtctl claim "$TASK_ID" claude
era store --type session -t "fluentfp,plan" "$(cat <<'MEMO'
Phase 1 of TOC DBR redesign: rename per-stage rope terminology to WIP limits,
add goodput/arrival rate signals to IntervalStats, add drum starvation tracking
to analyzer snapshots, make constraint identity sticky. Drum demo visualizes
correct vs wrong drum assignment.
MEMO
)"
```

## At Completion Gate

```bash
evtctl complete '{"criteria":[{"name":"rename RopeWait to WIPWait","status":"delivered","evidence":"14 occurrences renamed, zero remaining references, tests pass"},{"name":"add Goodput to IntervalStats","status":"delivered","evidence":"interval.go:52, clamped for atomic skew, 3 tests"},{"name":"add ArrivalRate to IntervalStats","status":"delivered","evidence":"interval.go:53, submitted/elapsed"},{"name":"add DrumStarvationCount to Snapshot","status":"delivered","evidence":"analyzer.go:99, consecutive starved intervals, reset on constraint change"},{"name":"sticky constraint identity","status":"delivered","evidence":"analyzer.go:319-327, constraint persists until replaced"},{"name":"drum demo","status":"delivered","evidence":"examples/drum-demo/main.go, peak queue 190 vs 2"}]}'
evtctl done 27841 "Phase 1 complete: rename + goodput + arrival rate + starvation + demo"
era store --type session -t "fluentfp,completion" "$(cat <<'MEMO'
Phase 1 delivered: RopeWait→WIPWait rename, Goodput/ArrivalRate in IntervalStats,
DrumStarvationCount in analyzer Snapshot, sticky constraint identity. Drum demo
shows peak queue 190 (no drum) vs 2 (correct drum) at same throughput.
/i lessons: clamp goodput for atomic counter skew, reset starvation count on
constraint change, doc should match actual classifier signals not idealized description.
MEMO
)"
git add toc/toc.go toc/rope_test.go toc/interval.go toc/interval_test.go toc/analyze/analyzer.go toc/analyze/analyzer_test.go examples/drum-demo/main.go
git commit -m "feat(toc): Phase 1 — rename RopeWait→WIPWait, add goodput/arrival rate/starvation signals

Rename per-stage rope terminology to WIP limits (D-RENAME).
Add Goodput and ArrivalRate to IntervalStats and StageAnalysis.
Add DrumStarvationCount to analyzer Snapshot (Step 2 violation detection).
Make constraint identity sticky until replaced (prevent premature abandonment).
Add drum demo showing correct vs wrong drum assignment.

Co-Authored-By: Claude <noreply@anthropic.com>"
```

### Phase 2: Pipeline topology + manual drum

#### Context
Phase 3's aggregate rope needs to compute total WIP from release point to drum.
That requires knowing which stages are upstream of the drum. Currently there's no
topology — stages register as flat lists. This phase builds the data structure.

Per Ted's direction (#27901): stats-only registration, explicit edges, build the
type without changing analyzer classification. Phase 3 consumes it.

#### Changes

**Create `toc/pipeline.go`:**

Configure-then-freeze, same as Analyzer/Reporter. Panics on misconfiguration
(consistent with all toc types — configuration errors are bugs, not runtime).

```go
// Pipeline is a DAG topology descriptor. Configure with AddStage/AddEdge,
// then call Freeze. After Freeze, immutable and safe for concurrent reads.
// Does NOT own stages — stores name + func() Stats only.
type Pipeline struct {
    mu      sync.Mutex
    frozen  bool
    stages  map[string]pipelineEntry
    order   []string
    edges   []edge
    forward map[string][]string  // from -> []to (built on Freeze)
    reverse map[string][]string  // to -> []from (built on Freeze)
    heads   []string             // zero in-degree stages (built on Freeze)
}
```

API:
- `NewPipeline() *Pipeline`
- `AddStage(name string, stats func() Stats)` — panics on empty/nil/dup/frozen
- `AddEdge(from, to string)` — panics on unknown/dup/self-loop/frozen
- `Freeze()` — validates acyclicity (Kahn's), non-empty. Builds adjacency, computes heads. Panics on cycle/empty/double-freeze

Query API (on immutable *Pipeline):
- `Stages() []string` — registration order
- `Heads() []string` — all zero in-degree stages
- `HeadsTo(target string) []string` — heads that can reach target (for Phase 3: only gate heads feeding the drum)
- `AncestorsOf(target string) []string` — all stages transitively upstream of target, BFS closest-first. Excludes target itself
- `DirectPredecessors(name string) []string` — immediate upstream
- `HasPath(from, to string) bool` — reachability query
- `StageStats(name string) func() Stats` — stats accessor for Phase 3 rope

No drum in Pipeline — drum is dynamic control state, topology is static.
Drum stays in Analyzer (SetDrum on Analyzer, or automatic identification).

**Create `toc/pipeline_test.go`:**

Table-driven tests: AddStage panics (empty name, nil stats, dup, frozen),
AddEdge panics (unknown, dup, self-loop, frozen), Freeze panics (cycle,
empty, double-freeze), Heads/HeadsTo/AncestorsOf/DirectPredecessors/HasPath for
linear and diamond DAGs, Stages ordering, StageStats lookup. Single-stage
pipeline (head = only stage, AncestorsOf returns empty).

**Modify `toc/analyze/analyzer.go`:**
- Add `pipeline *toc.Pipeline` field
- Add `WithPipeline(p *toc.Pipeline) Option`
- Add `Pipeline() *toc.Pipeline` accessor
- Move `SetDrum(name string)` to Analyzer (manual drum override)
- No classification logic changes — Phase 3 wires this

**Modify `toc/doc.go`:**
- Add Pipeline/PipelineBuilder to package doc and compile-time export verification

#### Design decisions

**D-TOPO-PASSIVE:** Pipeline is passive metadata, not a stage owner. Stores
`name + func() Stats` only. Avoids the generics problem (Stage[T,R] varies).

**D-TOPO-FREEZE:** Configure-then-freeze, same as Analyzer/Reporter. Panics
on misconfiguration — configuration errors are bugs, caught on first run
during development. Consistent with all toc types. No error returns.

**D-TOPO-NO-DRUM:** Drum is not stored in Pipeline. The graph is static; the
drum is dynamic (analyzer identifies it, or operator overrides). Drum stays
in Analyzer. Topology queries take a target parameter instead.

**D-TOPO-HEADS-TO:** `HeadsTo(target)` returns only heads that can reach the
target. If two heads exist but only one feeds the drum, the rope must not gate
the other. Target-relative head query is essential for correct rope scoping.

**D-TOPO-NO-PATH:** No PathFromHead. In a diamond DAG, one shortest path is
meaningless — the rope needs the upstream subgraph, not a path. Shared nodes
break per-path budgeting. Phase 3 uses `AncestorsOf(drum)` + `HeadsTo(drum)`
instead.

**D-TOPO-REACHABILITY:** `HasPath(from, to)` as a first-class query. Cheap
BFS reachability that Phase 3 needs for eligibility checks.

**D-TOPO-VALIDATE:** Build validates: acyclic (Kahn's), non-empty. No
disconnected-stage check — DAG properties guarantee reachability from some
head, and independent components (monitoring branches) are legitimate.

## At Implementation Gate

```bash
evtctl contract '{"phase":"phase2-pipeline-topology","criteria":["Pipeline with AddStage/AddEdge/Freeze","HeadsTo target-relative head query","AncestorsOf BFS traversal","DirectPredecessors","HasPath reachability","Freeze validation with Kahns cycle detection","Analyzer WithPipeline + SetDrum","Tests for all Pipeline methods"]}'
evtctl plan ~/.claude/plans/unified-napping-peach.md
TASK_ID=27842
evtctl claim "$TASK_ID" claude
era store --type session -t "fluentfp,plan" "$(cat <<'MEMO'
Phase 2: Pipeline topology type. Builder pattern with error returns. Immutable
after Build. No drum in topology (stays in Analyzer). Target-relative queries:
HeadsTo, AncestorsOf, HasPath. No PathFromHead (upstream subgraph, not paths).
MEMO
)"
```

## At Completion Gate

```bash
evtctl complete '<compose: {"criteria":[...status/evidence per criterion...]}'
evtctl done 27842 "Phase 2 complete: pipeline topology"
era store --type session -t "fluentfp,completion" "$(cat <<'MEMO'
<what delivered, /i lessons, insights>
MEMO
)"
git add toc/pipeline.go toc/pipeline_test.go toc/analyze/analyzer.go toc/doc.go
git commit -m "feat(toc): Phase 2 — Pipeline topology type with DAG validation

Co-Authored-By: Claude <noreply@anthropic.com>"
```

## Verification
```bash
go test ./toc/... -count=1 -race -timeout 60s
go build ./...
go vet ./...
```

### Phase 3: Aggregate rope — single-head count-based controller

#### Context
Per-stage WIP limits (MaxWIP) don't control aggregate pipeline WIP — items just
relocate to upstream queues (drum-demo proved this: 11MB vs 1MB at same throughput).
The true rope limits total WIP from release point (head) to drum (constraint).

Phase 3 scope (per Ted #27977): single-head, count-based rope. Weight/fairness/
multi-head/memory rope are future iterations. Controller approach — periodic
goroutine adjusts head stage's SetMaxWIP. No new admission path.

#### Rope length formula (Little's Law heuristic + yield adjustment)
```
requiredReleaseRate = drumGoodput / (1 - drumErrorRate)
upstreamFlowTime = sum(estimatedFlowTime for AncestorsOf(drum))
ropeLength = ceil(requiredReleaseRate * upstreamFlowTime * safetyFactor)
```

Flow time per stage estimated from IntervalStats:
`(ServiceTimeDelta + OutputBlockedDelta) / max(1, ItemsCompleted)`
This approximates sojourn time (service + output wait). Excludes idle time
(waiting for input is the rope's job to manage, not count). Better than
MeanServiceTime alone (which ignores queue/blocking delays).

Floor of 1. Safety factor default 1.5. All rate/time signals EWMA-smoothed
(α = 0.3) to prevent oscillation from interval noise. First tick seeds
EWMA to raw value; missing samples (zero completions) hold previous EWMA.

Yield inflation capped at 10× to prevent blow-up from noisy error rates
near 1.0: `requiredReleaseRate = min(drumGoodput * 10, drumGoodput / (1 - errorRate))`.

InitialRopeLength used only before first valid measurement. After warmup,
zero goodput holds last known rope length (does not collapse back to initial).

This is an **approximate soft control**, not a hard WIP cap. SetMaxWIP cannot
revoke existing permits and has a floor of 1. Overshoot: after a target
decrease, already-admitted items persist until completion — aggregate WIP
may exceed target for the duration of their flow time. The controller
cannot retract work in progress. Intermediate stage MaxWIP caps may also
prevent rope realization — the effective system WIP cap is the tightest
local cap on the path.

#### Head MaxWIP adjustment
```
aggregateWIP = sum(Stats.Admitted for each stage in AncestorsOf(drum))
downstreamWIP = aggregateWIP - headAdmitted
headMaxWIP = max(1, ropeLength - downstreamWIP)
headStage.SetMaxWIP(headMaxWIP)
```

This ensures: head only admits enough to keep total upstream WIP ≈ ropeLength.
When downstream drains, head opens. When downstream fills, head tightens.

#### Changes

**Create `toc/rope_controller.go`:**

```go
type RopeController struct {
    pipeline      *Pipeline
    drum          string
    head          string            // HeadsTo(drum)[0]
    ancestors     []string          // AncestorsOf(drum), cached
    setHeadWIP    func(int) int     // head stage's SetMaxWIP
    stageSnapshot func(string) IntervalStats
    interval      time.Duration
    safetyFactor  float64           // default 1.5
    initialLength int               // bootstrap rope length before first valid signal
    logger        *log.Logger
    warmedUp      bool              // true after first valid goodput measurement

    // EWMA state (written only by adjust goroutine, no lock needed).
    ewmaGoodput   float64
    ewmaErrorRate float64
    ewmaFlowTime  map[string]float64 // per-ancestor smoothed flow time

    // atomic stats for lock-free reads by Stats() callers.
    ropeLength, ropeWIP, adjustmentCount atomic.Int64
    drumGoodput, drumErrorRate           atomic.Int64 // float64 bits
}
```

Options:
- `WithRopeSafetyFactor(f float64)` — default 1.5
- `WithRopeLogger(l *log.Logger)` — default log.Default()
- `WithInitialRopeLength(n int)` — used until first valid measurement. Default 1 (conservative).

API:
- `NewRopeController(pipeline, drum, setHeadWIP, stageSnapshot, interval, opts...)`
  Panics if: not frozen, unknown drum, HeadsTo(drum) != 1 head,
  **non-linear path from head to drum** (for each internal ancestor on the
  path: in-degree=1 AND out-degree=1 within the head-to-drum subgraph;
  head: out-degree=1; drum predecessor: in-degree=1. Rejects fan-out,
  fan-in, side branches, and external inputs on the controlled path),
  nil funcs, interval <= 0
- `Run(ctx context.Context)` — blocks, adjusts every interval. Panics on double-call.
- `Stats() RopeStats` — RopeLength, RopeWIP, RopeUtilization, DrumGoodput,
  DrumErrorRate, AdjustmentCount, HeadAppliedWIP

Control loop: `runWithTicker(ctx, ticks)` pattern (same as Rebalancer/Reporter).
Each tick calls `adjust()`:
1. Read stageSnapshot(drum) → Goodput, ErrorRate (EWMA-smoothed)
2. If no valid signal yet → use InitialRopeLength, set headMaxWIP accordingly
3. Compute flow time per ancestor from IntervalStats
4. Compute rope length via Little's Law heuristic
5. Sum Admitted across ancestors → aggregate WIP
6. headMaxWIP = ropeLength - downstreamWIP (clamped ≥ 1)
7. setHeadWIP(headMaxWIP)

Edge cases: zero goodput before warmup → use InitialRopeLength, zero goodput
after warmup → hold last known rope length, 100% error rate → skip,
yield inflation capped at 10×, zero flow time → ropeLength=1,
negative Admitted → clamp to 0.

**Create `toc/rope_controller_test.go`:**

Tests with real Pipeline + mock stats/snapshot funcs + manual ticks:
- BasicAdjustment: known goodput/flow times → expected rope length
- ZeroGoodput: uses InitialRopeLength
- HighDownstreamWIP: head tightened to 1
- LowDownstreamWIP: head gets full rope length
- SafetyFactor: scales proportionally
- ErrorRateAdjustment: 50% errors → ~2× rope length
- ErrorRateInflationCap: 99% errors → capped at 10× (not 100×)
- FloorOfOne: extreme values → headMaxWIP never < 1
- EWMASmoothing: noisy signals produce stable rope length
- EWMAFirstSample: first tick seeds to raw value, not zero
- LinearChainValidation: constructor panics on branching topology
- LinearChainFanOut: constructor panics if head has extra successor
- ColdStart: InitialRopeLength used until first valid measurement
- HoldOnZeroGoodput: after warmup, zero goodput holds last rope length
- StopsOnCancel, PanicsOnDoubleRun, PanicsOnMultiHead
- Stats snapshot reflects current state

## At Implementation Gate

```bash
evtctl contract '{"phase":"phase3-aggregate-rope","criteria":["RopeController with Run/Stats","Rope length via flow time heuristic + yield adjustment","Head MaxWIP adjustment from aggregate upstream WIP","EWMA smoothing on rate/time signals","Linear chain validation in constructor","InitialRopeLength for cold start","Edge cases (zero goodput, high error rate, floor of 1)","Tests with manual ticks and mock stages"]}'
evtctl plan ~/.claude/plans/unified-napping-peach.md
TASK_ID=27843
evtctl claim "$TASK_ID" claude
era store --type session -t "fluentfp,plan" "$(cat <<'MEMO'
Phase 3: single-head count-based aggregate rope controller. Periodic goroutine
reads drum goodput + upstream service times, computes rope length via Little's Law
with yield adjustment, sets head MaxWIP to bound aggregate upstream WIP. No new
admission path — composes with existing per-stage SetMaxWIP.
MEMO
)"
```

## At Completion Gate

```bash
evtctl complete '<compose: {"criteria":[...status/evidence per criterion...]}'
evtctl done 27843 "Phase 3 complete: aggregate rope controller"
era store --type session -t "fluentfp,completion" "$(cat <<'MEMO'
<what delivered, /i lessons, insights>
MEMO
)"
git add toc/rope_controller.go toc/rope_controller_test.go
git commit -m "feat(toc): Phase 3 — aggregate rope controller (single-head, count-based)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

## Verification
```bash
go test ./toc/... -count=1 -race -timeout 60s
go build ./...
go vet ./...
```

### Phase 4: Buffer sizing + fever chart

#### Context
The buffer protects the drum from starvation. Size matters: too small → drum
starves on upstream hiccups; too large → excess memory. Goldratt's fever chart
(green/yellow/red) gives operators a simple health signal. We already have
BufferZone and CurrBufferPenetration in IntervalStats. Phase 4 adds:
recommended buffer size computation, and a memory headroom fever chart.

#### Changes

**Add to `toc/interval.go`:**

```go
// BufferCapacity computes how many items a buffer needs to hold to
// protect the drum from starvation over a given protection window.
// Goldratt's formula: ceil(throughput × protectionTime).
// Returns 0 if throughput <= 0 or protectionTime <= 0.
func BufferCapacity(throughput float64, protectionTime time.Duration) int

// MemoryFeverZone classifies memory headroom.
type MemoryFeverZone int
const (
    MemoryUnknown MemoryFeverZone = iota // limit=0 or unavailable
    MemoryGreen                          // < yellowAt consumed
    MemoryYellow                         // yellowAt to redAt consumed
    MemoryRed                            // >= redAt consumed
)

// MemoryFever classifies memory headroom into a fever zone.
// yellowAt and redAt are penetration thresholds (0.0-1.0).
// Returns MemoryUnknown if limit == 0.
func MemoryFever(headroom, limit uint64, yellowAt, redAt float64) MemoryFeverZone
```

Both functions take strategies as parameters:
- `BufferCapacity` takes `protectionTime` directly — caller computes it
  however they want (drum P95, upstream lead time, fixed duration)
- `MemoryFever` takes `yellowAt`/`redAt` thresholds — caller sets the
  sensitivity (Goldratt thirds 0.33/0.66, or tighter/looser)

No named strategy types. No built-in defaults baked in. Consumer code
passes the values.

**Add to `toc/interval_test.go`:**

Tests for BufferCapacity: positive throughput/time, zero throughput → 0,
zero duration → 0, negative inputs → 0, large values don't overflow.
Tests for MemoryFever: green/yellow/red thresholds, limit=0 → Unknown,
headroom > limit → clamp to Green, at limit → Red.

#### Design decisions

**D-BUFFER-FUNC:** `BufferCapacity` is pure math (throughput × time). No
strategy param — there's no logic to parameterize, just arithmetic.
Protection time is a plain `time.Duration` the caller computes.

**D-MEMORY-STRATEGY:** `MemoryFever` takes `yellowAt`/`redAt` thresholds as
params because the classification logic lives in the function. Consumer
controls sensitivity. `MemoryUnknown` for limit=0 — no silent default.

Note: existing `IntervalStats.BufferZone()` has hardcoded 0.33/0.66 — shipped
API, not changing. Consumers wanting custom thresholds use `CurrBufferPenetration`
directly.

## At Implementation Gate

```bash
evtctl contract '{"phase":"phase4-buffer-fever","criteria":["BufferCapacity pure function","MemoryFeverZone with Unknown state","MemoryFever function","Tests for buffer capacity edge cases","Tests for memory fever thresholds including Unknown"]}'
evtctl plan ~/.claude/plans/unified-napping-peach.md
TASK_ID=27845
evtctl claim "$TASK_ID" claude
era store --type session -t "fluentfp,plan" "$(cat <<'MEMO'
Phase 4: buffer sizing + fever charts. Pure functions — BufferRecommendation
from drumGoodput × variabilityWindow (P95/P50), MemoryFever from headroom/limit.
No new goroutines — composable with existing Reporter/RopeController.
MEMO
)"
```

## At Completion Gate

```bash
evtctl complete '<compose: {"criteria":[...status/evidence...]}'
evtctl done 27845 "Phase 4 complete: buffer sizing + fever chart"
era store --type session -t "fluentfp,completion" "$(cat <<'MEMO'
<what delivered, /i lessons, insights>
MEMO
)"
git add toc/interval.go toc/interval_test.go
git commit -m "feat(toc): Phase 4 — buffer sizing recommendation + memory fever chart

Co-Authored-By: Claude <noreply@anthropic.com>"
```

## Verification
```bash
go test ./toc/... -count=1 -race -timeout 60s
go build ./...
go vet ./...
```

### Phase 5: Five Focusing Steps — state classifier + constraint migration

#### Context
The Five Steps are a human decision framework, not a state machine. The runtime
pieces exist (Analyzer, RopeController, Rebalancer). Phase 5 ties them together:
a state classifier that reads their outputs and narrates which step the system is
in, plus a constraint migration protocol for when the drum moves.

Per Ted (#28210): light classifier, not an owner. Coordinator manages rope lifecycle
(clean rebuild on drum change). Not a framework — a function + example.

#### Changes

**Create `toc/focusing.go`:**

```go
// FocusingStep classifies the pipeline's operational state, inspired
// by Goldratt's Five Focusing Steps adapted for software pipelines.
type FocusingStep int
const (
    StepIdentify    FocusingStep = iota + 1 // searching for constraint
    StepExploit                              // constraint found, setting up rope/buffer
    StepSubordinate                          // rope active, healthy steady state
    StepElevate                              // moving resources to constraint
    StepReassess                             // constraint moved, rebuilding rope
)

// ClassifyStep determines the current focusing step from system state.
// Pure function — no side effects. Takes prev and curr constraint so
// the comparison logic lives with the classification.
func ClassifyStep(
    prevConstraint string,    // previous interval's constraint (empty = none)
    currConstraint string,    // current interval's constraint (empty = none)
    ropeActive bool,          // rope controller running for current constraint
    rebalancing bool,         // rebalancer made a change this interval
    starving bool,            // DrumStarvationCount > 0 (Step 2 violation)
) FocusingStep
```

Classification logic (priority order):
1. No constraint identified (currConstraint empty) → StepIdentify
2. Constraint changed (prev != curr, both non-empty) → StepReassess
3. Rebalancer actively moving workers → StepElevate
4. Constraint starving (Step 2 violation) → StepExploit (exploitation failing)
5. Constraint found, rope not yet active → StepExploit (setting up)
6. Constraint found, rope active, not starving → StepSubordinate

No MigrateRope function — it's just NewRopeController with a new drum.
The migration protocol is documented: cancel old context, call
NewRopeController with new drum. EWMA starts fresh (correct — old
drum's signals are irrelevant).

**Create `toc/focusing_test.go`:**

Tests for ClassifyStep: all step classifications, priority ordering
(Reassess > Elevate > Exploit > Subordinate), edge cases: constraint
lost (non-empty→empty = Identify), first identification (empty→non-empty),
rebalancing + starving simultaneously (Elevate wins), starving without
rope (Exploit).

**No periodic goroutine.** ClassifyStep is called by the consumer per
tick from whatever coordination loop they have.

#### Design decisions

**D-FOCUS-FUNC:** ClassifyStep is a pure function. The Five Steps are
a classification of system state, not a runtime loop. Consumer calls it
per interval from their own coordination logic.

**D-NO-MIGRATE:** No MigrateRope wrapper — NewRopeController already
does everything needed. Document the protocol: cancel old, build new.

## At Implementation Gate

```bash
evtctl contract '{"phase":"phase5-focusing-steps","criteria":["FocusingStep type with String method","ClassifyStep pure function","Tests for all step classifications","Documented constraint migration protocol"]}'
evtctl plan ~/.claude/plans/unified-napping-peach.md
TASK_ID=27848
evtctl claim "$TASK_ID" claude
era store --type session -t "fluentfp,plan" "$(cat <<'MEMO'
Phase 5: Five Focusing Steps as pure ClassifyStep function + MigrateRope
for constraint migration. No framework, no goroutine. Consumer calls
ClassifyStep per interval, calls MigrateRope when constraint moves.
MEMO
)"
```

## At Completion Gate

```bash
evtctl complete '<compose: {"criteria":[...status/evidence...]}'
evtctl done 27848 "Phase 5 complete: Five Focusing Steps classifier + constraint migration"
era store --type session -t "fluentfp,completion" "$(cat <<'MEMO'
<what delivered, /i lessons, insights>
MEMO
)"
git add toc/focusing.go toc/focusing_test.go
git commit -m "feat(toc): Phase 5 — Five Focusing Steps classifier + constraint migration

Co-Authored-By: Claude <noreply@anthropic.com>"
```

## Verification
```bash
go test ./toc/... -count=1 -race -timeout 60s
go build ./...
go vet ./...
```
