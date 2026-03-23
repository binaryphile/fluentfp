# Pipeline Pattern Catalog

Characterization of pipeline topologies, their observable resource signatures,
and recommendations for different application contexts.

## Pattern Classification Axes

| Axis | Values |
|------|--------|
| **Topology** | Linear, Fan-out (Tee), Fan-in (Merge), Diamond, DAG |
| **Constraint type** | CPU-bound, IO-bound, Memory-bound, Mixed |
| **Item uniformity** | Uniform weight, Variable weight, Batch-dependent |
| **Arrival pattern** | Steady, Bursty, Backlog-driven |
| **Worker model** | Fixed, Elastic (SetWorkers), Memory-budgeted |

## Pattern 1: Linear Chain (The Standard)

```
head → stage₁ → stage₂ → ... → stageₙ
```

**Topology:** Each stage has exactly one predecessor and one successor.

**Observable signature:**
- WIP profile: spike-then-drain without rope; steady-state with rope
- Memory trajectory: proportional to total items in flight
- Throughput: limited by slowest stage (the drum)
- Latency: sum of per-stage service times + queue wait

**When to use:**
- CPU-bound batch processing (era's walk → chunk → embed → store)
- Any sequential transformation pipeline
- When stages have clear ordering and no parallelizable branches

**Rope strategy:** Count-based or weight-based RopeController on head.
Linear chain validation passes. This is the only topology the current
rope controller supports.

**Example:** `examples/drum-demo/` — demonstrates resource cost across
three configurations (no limits, wrong stage, correct drum).

---

## Pattern 2: Fan-Out (Tee)

```
source → Tee → [branch₁, branch₂, ..., branchₙ]
```

**Topology:** One stage produces identical copies to N branches. All
branches process the same items independently.

**Observable signature:**
- WIP profile: N× multiplier from the tee point
- Memory trajectory: N× peak from a single-branch baseline
- Throughput: limited by the slowest branch
- Latency: max of branch latencies (for the overall result)

**When to use:**
- Parallel independent processing of the same data
- era's FTS + HNSW rebuild from the same finalize input
- Write-ahead + primary processing paths

**Rope strategy:** Rope on the pre-tee head. Each branch may have
different constraint characteristics. The aggregate rope must account
for all branches' WIP. Currently unsupported by RopeController (single
head, linear chain only). Future: multi-path rope with BOM ratios.

**Key risk:** Branch imbalance — if one branch is 10× slower than others,
the fast branches complete and idle while the slow branch determines
overall throughput.

---

## Pattern 3: Fan-In (Merge)

```
[source₁, source₂, ..., sourceₙ] → Merge → sink
```

**Topology:** Multiple independent sources feed a single downstream stage.

**Observable signature:**
- WIP profile: aggregate of all source WIP at the merge point
- Memory trajectory: sum of all source in-flight items
- Throughput: sum of source throughputs (if sink can keep up)
- Latency: per-item, determined by the source path + sink processing

**When to use:**
- Aggregating results from parallel independent producers
- Multi-tenant pipelines feeding a shared processing stage
- Merging after Tee+branch processing

**Rope strategy:** Multiple heads feeding the drum. HeadsTo(drum) returns
all source heads. The rope must gate each head proportionally — BOM
ratios determine how many items each source contributes per drum output.
Currently unsupported by RopeController.

**Key risk:** Source starvation — if the merge stage has a tight WIP
limit, faster sources can monopolize it, starving slower sources.
Fairness requires per-source admission budgets.

---

## Pattern 4: Diamond (Fan-Out + Fan-In)

```
        ┌→ branch₁ →┐
source ─┤            ├→ merge → sink
        └→ branch₂ →┘
```

**Topology:** Fan-out followed by fan-in. Common in map-reduce and
parallel processing patterns.

**Observable signature:**
- WIP profile: expands at fan-out, contracts at fan-in
- Memory trajectory: peak at the widest point (all branches active)
- Throughput: limited by the slowest branch (determines merge rate)
- Latency: max(branch latencies) + merge overhead

**When to use:**
- Parallel processing of different aspects of the same data
- era's ExtDocs + CallGraph → EdgeResolve pattern
- Any pattern where work splits, processes independently, then combines

**Rope strategy:** Aggregate WIP from source through all branches to
the merge point. The rope must track WIP across the diamond, not just
one path. PathFromHead is meaningless — AncestorsOf(merge) gives the
full upstream subgraph.

**Key risk:** Reconvergent WIP accounting. If both branches admit the
same source item independently, sum(Admitted) double-counts. Need
path-aware WIP tracking or normalized counting.

---

## Pattern 5: Staged Batch Pipeline

```
source → accumulate(n) → batch_process → scatter → per_item_finish
```

**Topology:** Linear with cardinality changes. Accumulator reduces N
items to 1 batch. Scatter expands 1 batch result to N items.

**Observable signature:**
- WIP profile: sawtooth — builds to batch size, drops on flush, repeats
- Memory trajectory: peaks at batch boundaries (batch + accumulated items)
- Throughput: batch-dependent — larger batches amortize overhead but
  increase latency and memory
- Latency: batch fill time + batch processing time

**When to use:**
- GPU/ML inference (batch embeddings)
- Bulk database writes (StoreBatch)
- Network request batching (HTTP/2 multiplexing)

**Rope strategy:** Weight-based rope where weight = batch memory cost.
The accumulator's weight grows as items accumulate. The rope limits
total accumulated weight, not item count (a batch of 64 embeddings
costs more than 64 individual items in queue).

**Key risk:** Batch size vs latency tradeoff. Larger batches improve
throughput but increase per-item latency. Dynamic batch sizing (#28309)
adjusts based on arrival rate and constraint utilization.

---

## Constraint Type Profiles

### CPU-Bound

**Characteristics:**
- Utilization near 100% at the constraint
- Adding workers helps linearly (until CPU cache/scheduling limits)
- ScalingHistory shows diminishing returns at core count
- Memory is secondary — items are small relative to processing time

**Rope sizing:** Little's Law works well. Flow time ≈ service time.
Safety factor 1.2-1.5×.

**Example:** Text chunking, hashing, compression.

### IO-Bound

**Characteristics:**
- High idle/blocked time at the constraint
- Workers help dramatically (parallelism hides latency)
- ScalingHistory shows gains well beyond core count
- Memory may become constraint with many concurrent connections

**Rope sizing:** Little's Law needs full flow time (service + blocked).
Higher safety factor (2-3×) because IO variance is high.

**Example:** HTTP API calls, database queries, file I/O.

### Memory-Bound

**Characteristics:**
- MemoryFever in yellow/red zones
- Adding workers doesn't help (each worker adds memory cost)
- ScalingHistory shows gains plateau or reverse with more workers
- Processing throughput is secondary to memory headroom

**Rope sizing:** Memory rope governs via BudgetAllocator. Processing
rope may be looser. BudgetStrategy (DefaultBudget) computes workers
from per-worker memory cost.

**Example:** ORT embedding (native memory outside Go heap), image
processing, large model inference.

### Mixed (Constraint Migration)

**Characteristics:**
- Constraint moves between CPU and memory as load changes
- FocusingStep cycles: Identify → Exploit → Elevate → back to Identify
- Elevation of CPU constraint may trigger memory constraint
- DrumStarvationCount spikes during transitions

**Rope sizing:** Both processing and memory ropes active. LimitManager
composes via min. The tighter rope governs at any moment.

**Example:** era's fullIndex pipeline — embed is CPU-bound at low
concurrency, memory-bound when elevated to many workers.

---

## Configuration Recommendations

### By Application Type

| Application | Pattern | Constraint | Workers | Rope | Buffer |
|-------------|---------|------------|---------|------|--------|
| Data pipeline (era) | Linear | CPU/Memory | Elastic | Count + Memory | Drum P95 |
| API gateway | Fan-in | IO | Fixed high | Count per source | Arrival × P99 |
| ML inference | Staged batch | Memory | Budget-limited | Weight | Batch size × 2 |
| Stream processing | Linear | IO | Elastic | Count | Low (latency) |
| Map-reduce | Diamond | CPU | Per-branch | Aggregate weight | Per-branch |

### Elevation Tradeoffs (#27919)

When elevating the constraint (Step 4), consider:

| Tradeoff | More workers | Fewer workers |
|----------|-------------|---------------|
| CPU | Higher throughput | Lower cache contention |
| Memory | Higher total footprint | Lower per-item cost |
| GC | More allocation pressure | Less GC pause impact |
| Latency | More parallel processing | Less context switching |
| Batch size | Can process larger batches | Memory for larger batches |

GC policy (GOMEMLIMIT) trades CPU for memory: tighter heap limit forces
more frequent GC, freeing memory for WIP but consuming CPU cycles.

Process isolation (#27922) trades IPC overhead for deterministic memory
reclamation: fork child processes for memory-heavy stages, OS reclaims
on exit. Bypasses GC entirely for native allocations (ORT/CGo).

### Dynamic Batch Sizing (#28309)

The WeightedBatcher threshold should adapt to:
- Arrival rate: higher arrival → larger batches (amortize overhead)
- Constraint utilization: saturated → smaller batches (reduce latency)
- Memory headroom: tight → smaller batches (reduce peak)
- Error rate: high errors → smaller batches (reduce waste)

This is a control variable the rope can adjust — another knob alongside
worker count and WIP limits.
